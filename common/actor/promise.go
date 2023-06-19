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

var ErrClosed = errors.New("promise closed")

// Promise is a simple promise-like object.
type Promise[T any] struct {
	once   sync.Once
	doneCh chan struct{}
	value  T
	err    error
}

func NewPromise[T any]() *Promise[T] {
	return &Promise[T]{
		doneCh: make(chan struct{}),
	}
}

func (p *Promise[T]) Resolve(value T) {
	p.once.Do(func() {
		p.value = value
		close(p.doneCh)
	})
}

func (p *Promise[T]) Reject(err error) {
	p.once.Do(func() {
		p.err = err
		close(p.doneCh)
	})
}

func (p *Promise[T]) Close() {
	p.Reject(ErrClosed)
}

func (p *Promise[T]) Wait(ctx context.Context) (T, error) {
	var t T
	select {
	case <-p.doneCh:
		if p.err != nil {
			return t, p.err
		}
		return p.value, nil
	case <-ctx.Done():
		return t, ctx.Err()
	}
}
