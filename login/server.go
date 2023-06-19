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
	"github.com/pangbox/server/gen/proto/go/topologypb/topologypbconnect"
	log "github.com/sirupsen/logrus"
)

// Options specify the options to use to instantiate the login server.
type Options struct {
	TopologyClient  topologypbconnect.TopologyServiceClient
	AccountsService *accounts.Service
}

// Server provides an implementation of the PangYa login server.
type Server struct {
	topologyClient  topologypbconnect.TopologyServiceClient
	accountsService *accounts.Service
	baseServer      *common.BaseServer
}

// New creates a new instance of the login server.
func New(opts Options) *Server {
	return &Server{
		topologyClient:  opts.TopologyClient,
		accountsService: opts.AccountsService,
		baseServer:      &common.BaseServer{},
	}
}

// Listen listens for connections on the given port and blocks indefinitely.
func (s *Server) Listen(ctx context.Context, addr string) error {
	logger := log.WithField("server", "LoginServer")
	return s.baseServer.Listen(logger, addr, func(logger *log.Entry, socket net.Conn) error {
		conn := Conn{
			ServerConn: common.NewServerConn(
				socket,
				logger,
				ClientMessageTable,
				ServerMessageTable,
			),
			topologyClient:  s.topologyClient,
			accountsService: s.accountsService,
		}
		return conn.Handle(ctx)
	})
}

func (s *Server) Shutdown(shutdownCtx context.Context) error {
	// TODO: Need to shut down connection threads.
	return s.baseServer.Close()
}
