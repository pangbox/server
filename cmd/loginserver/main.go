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
	"context"
	"flag"

	"github.com/pangbox/server/common/hash"
	"github.com/pangbox/server/common/topology"
	"github.com/pangbox/server/database"
	"github.com/pangbox/server/database/accounts"
	"github.com/pangbox/server/gameconfig"
	"github.com/pangbox/server/login"
	log "github.com/sirupsen/logrus"
	"github.com/xo/dburl"
)

//go:generate go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo -platform-specific=true

var (
	listenAddr  = ":10101"
	topologyURL = "h2c://localhost:41141"
	databaseURI = "sqlite://pangbox.sqlite3"
)

func init() {
	flag.StringVar(&topologyURL, "topology_url", topologyURL, "URL of topology server")
	flag.StringVar(&listenAddr, "addr", listenAddr, "Address to listen on for game server connections.")
	flag.StringVar(&databaseURI, "database", databaseURI, "Database URI.")
	flag.Parse()
}

func main() {
	ctx := context.Background()

	url, err := dburl.Parse(databaseURI)
	if err != nil {
		log.Fatalf("Error parsing database URL: %v", err)
	}

	db, err := database.OpenDBWithDriver(url.Driver, url.DSN)
	if err != nil {
		log.Fatalf("Failed to open DB: %v\n", err)
	}

	topologyClient, err := topology.NewClient(topology.ClientOptions{
		BaseURL: topologyURL,
	})
	if err != nil {
		log.Fatalf("Error creating topology client: %v", err)
	}

	log.Println("Listening for login server on", listenAddr)
	loginServer := login.New(login.Options{
		TopologyClient: topologyClient,
		AccountsService: accounts.NewService(accounts.Options{
			Database: db,
			Hasher:   hash.Bcrypt{},
		}),
		ConfigProvider: gameconfig.Default(),
	})
	log.Fatalln(loginServer.Listen(ctx, listenAddr))
}
