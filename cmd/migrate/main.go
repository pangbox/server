// Copyright (C) 2023, John Chadwick <john@jchw.io>
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
// SPDX-FileCopyrightText: Copyright (c) 2023 John Chadwick
// SPDX-License-Identifier: ISC

package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pangbox/server/database"
	_ "github.com/pangbox/server/migrations"
	"github.com/pressly/goose/v3"
	log "github.com/sirupsen/logrus"
	"github.com/xo/dburl"
	_ "modernc.org/sqlite"
)

const usageString = `Usage: %s COMMAND [ARGUMENTS...]

COMMANDS:
  up                   Migrate the DB to the most recent version available
  up-by-one            Migrate the DB up by 1
  up-to VERSION        Migrate the DB to a specific VERSION
  down                 Roll back the version by 1
  down-to VERSION      Roll back to a specific VERSION
  fix                  Apply sequential ordering to migrations
  redo                 Re-run the latest migration
  reset                Roll back all migrations
  status               Dump the migration status for the current DB
  version              Print the current version of the database
  create NAME [sql|go] Creates new migration file with the current timestamp

FLAGS:
`

func main() {
	dbstring := flag.String("database", "pgx://localhost", "Database URL.")
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, usageString, os.Args[0])
		flag.CommandLine.PrintDefaults()
		return
	}

	url, err := dburl.Parse(*dbstring)
	if err != nil {
		log.Fatalf("error parsing URL: %v", err)
	}

	db, err := database.OpenDBWithDriver(url.Driver, url.DSN)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()

	arguments := []string{}
	if len(args) > 1 {
		arguments = append(arguments, args[1:]...)
	}

	if err := goose.Run(args[0], db, ".", arguments...); err != nil {
		log.Fatalf("goose %v: %v", args[0], err)
	}
}
