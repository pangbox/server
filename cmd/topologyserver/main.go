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

package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/pangbox/server/common/topology"
	"github.com/pangbox/server/gen/proto/go/topologypb"
	"github.com/pangbox/server/gen/proto/go/topologypb/topologypbconnect"
	"github.com/rs/zerolog"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
)

//go:generate go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo -platform-specific=true

var (
	listenAddr       = ":41141"
	staticServerList = ""
	useH2C           = true
)

func init() {
	flag.StringVar(&listenAddr, "addr", listenAddr, "Address to listen on for topology server connections.")
	flag.StringVar(&staticServerList, "static_server_list", staticServerList, "Filename of static server list in JSON format.")
	flag.BoolVar(&useH2C, "use_h2c", useH2C, "Whether or not to enable H2C support.")
	flag.Parse()
}

func main() {
	log := zerolog.
		New(os.Stderr).
		With().
		Timestamp().
		Logger()

	if staticServerList == "" {
		log.Fatal().Msg("topology server requires a static server list for now")
	}

	jsonData, err := os.ReadFile(staticServerList)
	if err != nil {
		log.Fatal().Err(err).Msg("error reading static server list")
	}

	config := &topologypb.Configuration{}
	if err := protojson.Unmarshal(jsonData, config); err != nil {
		log.Fatal().Err(err).Msg("error parsing static server list")
	}

	serverList := []*topologypb.ServerEntry{}
	for _, server := range config.Servers {
		serverList = append(serverList, &topologypb.ServerEntry{
			Server: server,
		})
	}

	storage := topology.NewMemoryStorage(serverList)

	server := topology.NewServer(storage)

	_, handler := topologypbconnect.NewTopologyServiceHandler(server)
	httpserver := &http.Server{Handler: handler}
	if useH2C {
		httpserver.Handler = h2c.NewHandler(httpserver.Handler, &http2.Server{})
	}

	log.Info().Str("address", listenAddr).Msg("listening for message service connections")
	if err := httpserver.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("error in login server")
	}
}
