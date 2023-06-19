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

package minibox

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrServiceRunning = errors.New("service is already running")      // +checklocksignore
var ErrServiceStopped = errors.New("service is already stopped")      // +checklocksignore
var ErrServiceNotConfigured = errors.New("service is not configured") // +checklocksignore
var ErrStopping = errors.New("stopping service")                      // +checklocksignore

const ShutdownTimeout = 10 * time.Second

type ServiceShutdownFunc func(shutdownCtx context.Context) error
type ServiceSpawnFunc func(ctx context.Context, service *Service)

type Service struct {
	mu sync.RWMutex

	pctx   context.Context
	ctx    context.Context
	cancel context.CancelCauseFunc

	running bool

	spawn    ServiceSpawnFunc
	shutdown ServiceShutdownFunc
}

func NewService(ctx context.Context) *Service {
	service := new(Service)
	service.pctx = ctx
	return service
}

func (s *Service) start() {
	s.running = true
	ctx, cancel := context.WithCancelCause(s.pctx)
	go func() {
		s.spawn(ctx, s)
		s.stopCtx(ctx, cancel)
	}()
	s.ctx, s.cancel = ctx, cancel
}

func (s *Service) stop() {
	s.running = false
	s.cancel(ErrStopping)
	s.ctx = nil
	s.cancel = nil
	s.shutdown = nil
}

func (s *Service) stopCtx(ctx context.Context, cancel context.CancelCauseFunc) {
	// This function exists to solve a race condition when restarting a service.
	// If we restart fast enough, the goroutine of the previous run could run
	// stop after we've already started. Still cancel the old context
	// defensively even though it should already be cancelled.

	s.mu.Lock()
	defer s.mu.Unlock()

	if ctx == s.ctx {
		s.stop()
	} else {
		cancel(ErrStopping)
	}
}

func (s *Service) SetShutdownFunc(shutdown ServiceShutdownFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.shutdown = shutdown
}

func (s *Service) Running() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.running
}

func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return ErrServiceRunning
	}

	if s.spawn == nil {
		return ErrServiceNotConfigured
	}

	s.start()
	return nil
}

func (s *Service) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(s.pctx, ShutdownTimeout)
	defer cancel()

	return s.StopContext(shutdownCtx)
}

func (s *Service) StopContext(shutdownCtx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return ErrServiceStopped
	}

	var err error
	if s.shutdown != nil {
		err = s.shutdown(shutdownCtx)
	}

	s.stop()

	return err
}

func (s *Service) Configure(spawn ServiceSpawnFunc) error {
	shutdownCtx, cancel := context.WithTimeout(s.pctx, ShutdownTimeout)
	defer cancel()

	return s.ConfigureContext(shutdownCtx, spawn)
}

func (s *Service) ConfigureContext(shutdownCtx context.Context, spawn ServiceSpawnFunc) error {
	var err error

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		err = s.shutdown(shutdownCtx)
		s.stop()
	}

	s.spawn = spawn
	s.start()
	return err
}
