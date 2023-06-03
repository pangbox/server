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
	"database/sql"
	"errors"

	"github.com/pressly/goose/v3"
)

type Dialect int

const (
	DialectPostgreSQL Dialect = 1
	DialectSQLite     Dialect = 2
)

var (
	dialect Dialect
)

func AutoIncrementColumnType() string {
	switch dialect {
	case DialectPostgreSQL:
		return " SERIAL PRIMARY KEY "
	case DialectSQLite:
		return " INTEGER PRIMARY KEY "
	}
	panic("no dialect set")
}

func OpenDBWithDriver(driver string, dsn string) (*sql.DB, error) {
	switch driver {
	case "pgx":
		dialect = DialectPostgreSQL
		goose.SetDialect("postgres")
	case "sqlite", "sqlite3":
		dialect = DialectSQLite
		goose.SetDialect("sqlite3")
		driver = "sqlite"
	default:
		return nil, errors.New("unsupported SQL dialect")
	}
	return sql.Open(driver, dsn)
}
