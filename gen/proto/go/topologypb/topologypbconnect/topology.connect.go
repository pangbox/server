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

// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: topologypb/topology.proto

package topologypbconnect

import (
	context "context"
	errors "errors"
	connect_go "github.com/bufbuild/connect-go"
	topologypb "github.com/pangbox/server/gen/proto/go/topologypb"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect_go.IsAtLeastVersion0_1_0

const (
	// TopologyServiceName is the fully-qualified name of the TopologyService service.
	TopologyServiceName = "TopologyService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// TopologyServiceAddServerProcedure is the fully-qualified name of the TopologyService's AddServer
	// RPC.
	TopologyServiceAddServerProcedure = "/TopologyService/AddServer"
	// TopologyServiceListServersProcedure is the fully-qualified name of the TopologyService's
	// ListServers RPC.
	TopologyServiceListServersProcedure = "/TopologyService/ListServers"
	// TopologyServiceGetServerProcedure is the fully-qualified name of the TopologyService's GetServer
	// RPC.
	TopologyServiceGetServerProcedure = "/TopologyService/GetServer"
)

// TopologyServiceClient is a client for the TopologyService service.
type TopologyServiceClient interface {
	AddServer(context.Context, *connect_go.Request[topologypb.AddServerRequest]) (*connect_go.Response[topologypb.AddServerResponse], error)
	ListServers(context.Context, *connect_go.Request[topologypb.ListServersRequest]) (*connect_go.Response[topologypb.ListServersResponse], error)
	GetServer(context.Context, *connect_go.Request[topologypb.GetServerRequest]) (*connect_go.Response[topologypb.GetServerResponse], error)
}

// NewTopologyServiceClient constructs a client for the TopologyService service. By default, it uses
// the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewTopologyServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) TopologyServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &topologyServiceClient{
		addServer: connect_go.NewClient[topologypb.AddServerRequest, topologypb.AddServerResponse](
			httpClient,
			baseURL+TopologyServiceAddServerProcedure,
			opts...,
		),
		listServers: connect_go.NewClient[topologypb.ListServersRequest, topologypb.ListServersResponse](
			httpClient,
			baseURL+TopologyServiceListServersProcedure,
			opts...,
		),
		getServer: connect_go.NewClient[topologypb.GetServerRequest, topologypb.GetServerResponse](
			httpClient,
			baseURL+TopologyServiceGetServerProcedure,
			opts...,
		),
	}
}

// topologyServiceClient implements TopologyServiceClient.
type topologyServiceClient struct {
	addServer   *connect_go.Client[topologypb.AddServerRequest, topologypb.AddServerResponse]
	listServers *connect_go.Client[topologypb.ListServersRequest, topologypb.ListServersResponse]
	getServer   *connect_go.Client[topologypb.GetServerRequest, topologypb.GetServerResponse]
}

// AddServer calls TopologyService.AddServer.
func (c *topologyServiceClient) AddServer(ctx context.Context, req *connect_go.Request[topologypb.AddServerRequest]) (*connect_go.Response[topologypb.AddServerResponse], error) {
	return c.addServer.CallUnary(ctx, req)
}

// ListServers calls TopologyService.ListServers.
func (c *topologyServiceClient) ListServers(ctx context.Context, req *connect_go.Request[topologypb.ListServersRequest]) (*connect_go.Response[topologypb.ListServersResponse], error) {
	return c.listServers.CallUnary(ctx, req)
}

// GetServer calls TopologyService.GetServer.
func (c *topologyServiceClient) GetServer(ctx context.Context, req *connect_go.Request[topologypb.GetServerRequest]) (*connect_go.Response[topologypb.GetServerResponse], error) {
	return c.getServer.CallUnary(ctx, req)
}

// TopologyServiceHandler is an implementation of the TopologyService service.
type TopologyServiceHandler interface {
	AddServer(context.Context, *connect_go.Request[topologypb.AddServerRequest]) (*connect_go.Response[topologypb.AddServerResponse], error)
	ListServers(context.Context, *connect_go.Request[topologypb.ListServersRequest]) (*connect_go.Response[topologypb.ListServersResponse], error)
	GetServer(context.Context, *connect_go.Request[topologypb.GetServerRequest]) (*connect_go.Response[topologypb.GetServerResponse], error)
}

// NewTopologyServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewTopologyServiceHandler(svc TopologyServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	topologyServiceAddServerHandler := connect_go.NewUnaryHandler(
		TopologyServiceAddServerProcedure,
		svc.AddServer,
		opts...,
	)
	topologyServiceListServersHandler := connect_go.NewUnaryHandler(
		TopologyServiceListServersProcedure,
		svc.ListServers,
		opts...,
	)
	topologyServiceGetServerHandler := connect_go.NewUnaryHandler(
		TopologyServiceGetServerProcedure,
		svc.GetServer,
		opts...,
	)
	return "/.TopologyService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case TopologyServiceAddServerProcedure:
			topologyServiceAddServerHandler.ServeHTTP(w, r)
		case TopologyServiceListServersProcedure:
			topologyServiceListServersHandler.ServeHTTP(w, r)
		case TopologyServiceGetServerProcedure:
			topologyServiceGetServerHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedTopologyServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedTopologyServiceHandler struct{}

func (UnimplementedTopologyServiceHandler) AddServer(context.Context, *connect_go.Request[topologypb.AddServerRequest]) (*connect_go.Response[topologypb.AddServerResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("TopologyService.AddServer is not implemented"))
}

func (UnimplementedTopologyServiceHandler) ListServers(context.Context, *connect_go.Request[topologypb.ListServersRequest]) (*connect_go.Response[topologypb.ListServersResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("TopologyService.ListServers is not implemented"))
}

func (UnimplementedTopologyServiceHandler) GetServer(context.Context, *connect_go.Request[topologypb.GetServerRequest]) (*connect_go.Response[topologypb.GetServerResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("TopologyService.GetServer is not implemented"))
}
