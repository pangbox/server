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
	"runtime"

	_ "github.com/pangbox/server/migrations"
	"github.com/pangbox/server/minibox"
	log "github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

//go:generate go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo -manifest minibox.manifest -platform-specific=true

func cliMain() {
	ctx := context.Background()
	log.SetLevel(log.DebugLevel)
	log.Println("Welcome to Pangbox. Main thread started.")

	server := minibox.NewServer(ctx, log.WithContext(ctx))
	if err := server.ConfigureDatabase(dbOpts); err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}

	if err := server.ConfigureServices(opts); err != nil {
		log.Fatalf("Error setting up services: %v -- try setting -pangya_dir?", err)
	}

	log.Infof("Web server listening on %s", opts.WebAddr)
	log.Infof("QA auth server listening on %s", opts.QAAuthAddr)
	log.Infof("Login server listening on %s", opts.LoginAddr)
	log.Infof("Game server listening on %s", opts.GameAddr)
	log.Infof("Message server listening on %s", opts.MessageAddr)
	runtime.Goexit()
}
