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
	"testing"

	"github.com/pangbox/server/gen/dbmodels"
	_ "github.com/pangbox/server/migrations"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func RunSQLiteTest(t *testing.T, cb func(*testing.T, dbmodels.DBTX)) {
	t.Helper()

	db, err := OpenDBWithDriver("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}

	err = goose.Up(db, ".")
	if err != nil {
		t.Fatalf("Failed to run migrations forward: %v", err)
	}

	cb(t, db)

	err = goose.Down(db, ".")
	if err != nil {
		t.Fatalf("Failed to run migrations backward: %v", err)
	}
}
