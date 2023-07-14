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

package login

import (
	"context"
	"net"

	"github.com/pangbox/server/common"
	"github.com/pangbox/server/database/accounts"
	"github.com/pangbox/server/gameconfig"
	"github.com/pangbox/server/gen/proto/go/topologypb/topologypbconnect"
	"github.com/rs/zerolog"
)

// Options specify the options to use to instantiate the login server.
type Options struct {
	Logger          zerolog.Logger
	TopologyClient  topologypbconnect.TopologyServiceClient
	AccountsService *accounts.Service
	ConfigProvider  gameconfig.Provider
}

// Server provides an implementation of the PangYa login server.
type Server struct {
	log             zerolog.Logger
	topologyClient  topologypbconnect.TopologyServiceClient
	accountsService *accounts.Service
	baseServer      *common.BaseServer
	configProvider  gameconfig.Provider
}

// New creates a new instance of the login server.
func New(opts Options) *Server {
	return &Server{
		log:             opts.Logger.With().Str("server", "login").Logger(),
		topologyClient:  opts.TopologyClient,
		accountsService: opts.AccountsService,
		baseServer:      &common.BaseServer{},
		configProvider:  opts.ConfigProvider,
	}
}

// Listen listens for connections on the given port and blocks indefinitely.
func (s *Server) Listen(ctx context.Context, addr string) error {
	return s.baseServer.Listen(s.log, addr, func(log zerolog.Logger, socket net.Conn) error {
		conn := Conn{
			ServerConn: common.NewServerConn(
				socket,
				log,
				ClientMessageTable,
				ServerMessageTable,
			),
			s: s,
		}
		return conn.Handle(ctx)
	})
}

func (s *Server) Shutdown(shutdownCtx context.Context) error {
	// TODO: Need to shut down connection threads.
	return s.baseServer.Close()
}
