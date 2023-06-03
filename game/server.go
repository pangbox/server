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

package game

import (
	"context"
	"net"

	"github.com/pangbox/server/common"
	"github.com/pangbox/server/database/accounts"
	"github.com/pangbox/server/gen/proto/go/topologypb/topologypbconnect"
	"github.com/pangbox/server/pangya/iff"
	log "github.com/sirupsen/logrus"
)

// Options specify the options used to construct the game server.
type Options struct {
	TopologyClient  topologypbconnect.TopologyServiceClient
	AccountsService *accounts.Service
	PangyaIFF       *iff.Archive
	ServerID        uint32
	ChannelName     string
}

// Server provides an implementation of the PangYa game server.
type Server struct {
	baseServer      *common.BaseServer
	topologyClient  topologypbconnect.TopologyServiceClient
	accountsService *accounts.Service
	pangyaIFF       *iff.Archive
	serverID        uint32
	channelName     string
}

// New creates a new instance of the game server.
func New(opts Options) *Server {
	return &Server{
		baseServer:      &common.BaseServer{},
		topologyClient:  opts.TopologyClient,
		accountsService: opts.AccountsService,
		pangyaIFF:       opts.PangyaIFF,
		serverID:        opts.ServerID,
		channelName:     opts.ChannelName,
	}
}

// Listen listens for connections on a given address and blocks indefinitely.
func (s *Server) Listen(ctx context.Context, addr string) error {
	logger := log.WithField("server", "GameServer")
	return s.baseServer.Listen(logger, addr, func(logger *log.Entry, socket net.Conn) error {
		conn := Conn{
			ServerConn: common.ServerConn[ClientMessage, ServerMessage]{
				Socket:    socket,
				Log:       logger,
				ClientMsg: ClientMessageTable,
				ServerMsg: ServerMessageTable,
			},
			s: s,
		}
		return conn.Handle(ctx)
	})
}

func (s *Server) Shutdown(shutdownCtx context.Context) error {
	// TODO: Need to shut down connection threads.
	return s.baseServer.Close()
}
