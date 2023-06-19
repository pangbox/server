// Copyright (C) 2018-2023, John Chadwick <john@jchw.io>
//
// Permission to use, copy, modify, and/or distribute this software for any purpose
// with or without fee is hereby granted, provided that the above copyright notice
// and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
// FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
// OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
// TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
// THIS SOFTWARE.
//
// SPDX-FileCopyrightText: Copyright (c) 2018-2023 John Chadwick
// SPDX-License-Identifier: ISC

package actor

import (
	"context"
	"errors"
	"sync"
)

var ErrActorDead = errors.New("actor dead")

type Message[T any] struct {
	Context context.Context
	Value   T
	Promise *Promise[any]
}

type Base[T any] struct {
	mutex sync.RWMutex
	task  *Task[T]
	err   error
}

type Task[T any] struct {
	wg     sync.WaitGroup
	msgCh  chan Message[T]
	doneCh chan struct{}
	cancel context.CancelFunc
	ctx    context.Context
}

// TryStart returns true if started, or false if there's already a running task.
func (b *Base[T]) TryStart(ctx context.Context, callback func(context.Context, *Task[T]) error) bool {
	// Fast/low contention path
	b.mutex.RLock()
	task := b.task
	b.mutex.RUnlock()

	if task != nil {
		return false
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Need to re-check now.
	if b.task != nil {
		return false
	}

	ctx, cancel := context.WithCancel(ctx)
	task = &Task[T]{
		msgCh:  make(chan Message[T]),
		doneCh: make(chan struct{}),
		ctx:    ctx,
		cancel: cancel,
	}

	// These signal the status to elsewhere.
	b.task = task
	b.err = nil

	go func() {
		defer func() {
			b.mutex.Lock()
			b.task = nil
			b.mutex.Unlock()

			task.wg.Wait()

			// At this point, nothing can see this task anymore.
			// It should be safe to close the message channel.
			close(task.msgCh)
			close(task.doneCh)
		}()
		defer cancel()
		if err := callback(ctx, task); err != nil {
			b.mutex.Lock()
			b.err = err
			b.mutex.Unlock()
		}
	}()

	return true
}

// Shutdown shuts down the task, if it's running. Shutdown will wait for the
// current instance of the task to fully shut down.
func (b *Base[T]) Shutdown(ctx context.Context) error {
	b.mutex.RLock()
	task := b.task
	b.mutex.RUnlock()

	if task != nil {
		task.cancel()
		select {
		case <-task.doneCh:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// Active returns true if the task is currently running. Note that by the time
// this function returns, the value it returns may already be stale.
func (b *Base[T]) Active() bool {
	b.mutex.RLock()
	task := b.task
	b.mutex.RUnlock()

	return task != nil
}

// Err returns the last error, if a task returns one.
func (b *Base[T]) Err() error {
	b.mutex.RLock()
	err := b.err
	b.mutex.RUnlock()

	return err
}

func (b *Base[T]) acquireTask() *Task[T] {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	task := b.task
	if task != nil {
		task.wg.Add(1)
	}

	return task
}

func (t *Task[T]) release() {
	if t != nil {
		t.wg.Done()
	}
}

// TrySend tries to send if it wouldn't block. The boolean result is set to true
// when the message is successfully sent, false otherwise.
func (b *Base[T]) TrySend(ctx context.Context, value T) (*Promise[any], bool) {
	msg := Message[T]{
		Context: ctx,
		Value:   value,
		Promise: NewPromise[any](),
	}

	task := b.acquireTask()
	defer task.release()

	if task == nil {
		return nil, false
	}

	defer task.wg.Done()

	select {
	case task.msgCh <- msg:
		return msg.Promise, true
	default:
		msg.Promise.Close()
		return nil, false
	}
}

// Send will block until the message is sent or the context is cancelled.
func (b *Base[T]) Send(ctx context.Context, value T) (*Promise[any], error) {
	msg := Message[T]{
		Context: ctx,
		Value:   value,
		Promise: NewPromise[any](),
	}

	task := b.acquireTask()
	defer task.release()

	if task == nil {
		return nil, ErrActorDead
	}

	select {
	case task.msgCh <- msg:
		return msg.Promise, nil
	case <-ctx.Done():
		msg.Promise.Close()
		return nil, ctx.Err()
	}
}

// Receive receives the next message in the mailbox. Note that the promise needs
// to be resolved or rejected for every message.
func (t *Task[T]) Receive() (Message[T], error) {
	var msg Message[T]

	msgCh := t.msgCh

	select {
	case msg = <-msgCh:
		return msg, nil
	case <-t.ctx.Done():
		return msg, t.ctx.Err()
	}
}
