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

package common

import (
	"net"
	"sync"

	"github.com/rs/zerolog"
)

type BaseHandlerFunc func(zerolog.Logger, net.Conn) error

type BaseServer struct {
	mu sync.RWMutex

	// +checklocks:mu
	listener net.Listener
}

func (b *BaseServer) setListener(l net.Listener) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.listener = l
}

// Listen implements a basic TCP server connection loop, dispatching connections
// to the handle function. It has some basic provisions for logging and error
// handling as well.
func (b *BaseServer) Listen(log zerolog.Logger, addr string, handler BaseHandlerFunc) error {
	var err error

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	b.setListener(listener)

	for {
		socket, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			log := log.With().Str("remote address", socket.RemoteAddr().String()).Logger()
			log.Info().Msg("entering connection thread")
			defer func() {
				if r := recover(); r != nil {
					log.Error().Any(zerolog.ErrorFieldName, r).Msg("panic in connection")
				}
				log.Info().Msg("exiting connection thread")
				if err := socket.Close(); err != nil {
					log.Warn().Err(err).Msg("error closing socket")
				}
			}()
			if err := handler(log, socket); err != nil {
				log.Error().Err(err).Msg("error in connection")
			}
		}()
	}
}

// Close stops accepting connections.
func (b *BaseServer) Close() error {
	var err error

	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.listener != nil {
		err = b.listener.Close()
	}
	return err
}
