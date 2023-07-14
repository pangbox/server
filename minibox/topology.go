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
	"crypto/tls"
	"net"
	"net/http"
	"strconv"

	"github.com/pangbox/server/common/bufconn"
	"github.com/pangbox/server/common/topology"
	"github.com/pangbox/server/gen/proto/go/topologypb"
	"github.com/pangbox/server/gen/proto/go/topologypb/topologypbconnect"
	"github.com/rs/zerolog"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type TopologyServerOptions struct {
	Logger zerolog.Logger

	ServerIP       string
	GameServerName string

	GamePort    uint16
	MessagePort uint16
}

type TopologyServer struct {
	pipe    *bufconn.Listener
	service *Service
	client  topologypbconnect.TopologyServiceClient
}

func NewLocalTopology(ctx context.Context) *TopologyServer {
	listener := bufconn.Listen(65536)

	h2transport := &http2.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return listener.Dial()
		},
	}
	client := &http.Client{Transport: h2transport}

	topology := &TopologyServer{
		pipe:    listener,
		service: NewService(ctx),
		client:  topologypbconnect.NewTopologyServiceClient(client, "https://localhost"),
	}

	return topology
}

func (t *TopologyServer) Configure(opts TopologyServerOptions) error {
	log := opts.Logger
	server := topology.NewServer(topology.NewMemoryStorage([]*topologypb.ServerEntry{
		{
			Server: &topologypb.Server{
				Type:     topologypb.Server_TYPE_GAME_SERVER,
				Name:     opts.GameServerName,
				Id:       20202,
				NumUsers: 1,
				MaxUsers: 2000,
				Address:  opts.ServerIP,
				Port:     uint32(opts.GamePort),
				Flags:    0x800,
			},
		},
		{
			Server: &topologypb.Server{
				Type:     topologypb.Server_TYPE_MESSAGE_SERVER,
				Name:     "MessageServer1",
				Id:       30303,
				NumUsers: 1,
				MaxUsers: 2000,
				Address:  opts.ServerIP,
				Port:     uint32(opts.MessagePort),
				Flags:    0x1000,
			},
		},
	}))

	_, handler := topologypbconnect.NewTopologyServiceHandler(server)

	spawn := func(ctx context.Context, service *Service) {
		httpserver := &http.Server{
			Handler: h2c.NewHandler(handler, &http2.Server{}),
			BaseContext: func(l net.Listener) context.Context {
				return ctx
			},
		}

		service.SetShutdownFunc(func(shutdownCtx context.Context) error {
			return httpserver.Shutdown(shutdownCtx)
		})

		if ctx.Err() != nil {
			log.Error().Err(ctx.Err()).Msg("cancelled before topology server could start")
		}

		err := httpserver.Serve(t.pipe)
		if err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("error serving topology server")
		}
	}

	t.service.Configure(spawn)

	return nil
}

func (t *TopologyServer) Client() topologypbconnect.TopologyServiceClient {
	return t.client
}

func (t *TopologyServer) Running() bool {
	return t.service.Running()
}

func (t *TopologyServer) Start() error {
	return t.service.Start()
}

func (t *TopologyServer) Stop() error {
	return t.service.Stop()
}

func getPort(addr string) (uint16, error) {
	_, portstr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}

	port, err := strconv.Atoi(portstr)
	if err != nil {
		return 0, err
	}

	return uint16(port), nil
}
