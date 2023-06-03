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

package topology

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/pangbox/server/gen/proto/go/topologypb"
	"github.com/pangbox/server/gen/proto/go/topologypb/topologypbconnect"
)

// Ensure that we are always implementing the full Topology service.
var _ = topologypbconnect.TopologyServiceHandler(&Server{})

// Server implements TopologyServiceServer.
type Server struct {
	storage Storage
}

// NewServer creates a new Topology server.
func NewServer(storage Storage) *Server {
	return &Server{storage}
}

// AddServer implements TopologyServiceServer.
func (s *Server) AddServer(ctx context.Context, request *connect.Request[topologypb.AddServerRequest]) (*connect.Response[topologypb.AddServerResponse], error) {
	err := s.storage.Put(uint16(request.Msg.Server.Id), &topologypb.ServerEntry{
		Server: request.Msg.Server,
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&topologypb.AddServerResponse{}), nil
}

// ListServers implements TopologyServiceServer.
func (s *Server) ListServers(ctx context.Context, request *connect.Request[topologypb.ListServersRequest]) (*connect.Response[topologypb.ListServersResponse], error) {
	// Get full server list.
	entries, err := s.storage.List()
	if err != nil {
		return nil, err
	}

	// Do filtering.
	filteredServers := []*topologypb.Server{}
	for _, entry := range entries {
		if request.Msg.Type != topologypb.Server_TYPE_UNSPECIFIED && entry.Server.Type != request.Msg.Type {
			continue
		}
		filteredServers = append(filteredServers, entry.Server)
	}

	return connect.NewResponse(&topologypb.ListServersResponse{Server: filteredServers}), nil
}

// GetServer implements TopologyServiceServer.
func (s *Server) GetServer(ctx context.Context, request *connect.Request[topologypb.GetServerRequest]) (*connect.Response[topologypb.GetServerResponse], error) {
	entry, err := s.storage.Get(uint16(request.Msg.Id))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&topologypb.GetServerResponse{Server: entry.Server}), nil
}
