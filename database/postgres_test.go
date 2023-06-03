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

package database

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pangbox/server/gen/dbmodels"
	_ "github.com/pangbox/server/migrations"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func RunPostgreSQLTest(t *testing.T, cb func(*testing.T, dbmodels.DBTX)) {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.RunContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("Could not terminate container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Could not determine dsn: %v", err)
	}

	db, err := OpenDBWithDriver("pgx", dsn)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}

	if err = goose.Up(db, "."); err != nil {
		t.Fatalf("Failed to run migrations forward: %v", err)
	}

	cb(t, db)

	err = goose.Down(db, ".")
	if err != nil {
		t.Fatalf("Failed to run migrations backward: %v", err)
	}
}
