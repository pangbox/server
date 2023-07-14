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
	"os"
	"runtime"

	_ "github.com/pangbox/server/migrations"
	"github.com/pangbox/server/minibox"
	"github.com/rs/zerolog"
	_ "modernc.org/sqlite"
)

//go:generate go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo -manifest minibox.manifest -platform-specific=true

func cliMain() {
	ctx := context.Background()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log := zerolog.
		New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger()

	log.Info().Msg("initializing minibox")

	server := minibox.NewServer(ctx, log)
	if err := server.ConfigureDatabase(dbOpts); err != nil {
		log.Fatal().Err(err).Msg("error setting up database")
	}

	if err := server.ConfigureServices(opts); err != nil {
		log.Fatal().Err(err).Msg("error setting up services - try setting -pangya_dir")
	}

	log.Info().Str("address", opts.WebAddr).Msg("listening for web server connections")
	log.Info().Str("address", opts.AdminAddr).Msg("listening for admin web server connections")
	log.Info().Str("address", opts.QAAuthAddr).Msg("listening for qa auth server connections")
	log.Info().Str("address", opts.LoginAddr).Msg("listening for login server connections")
	log.Info().Str("address", opts.GameAddr).Msg("listening for game server connections")
	log.Info().Str("address", opts.MessageAddr).Msg("listening for message server connections")
	runtime.Goexit()
}
