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
	"database/sql"
	"testing"

	"github.com/pangbox/server/gen/dbmodels"
	"github.com/stretchr/testify/assert"
)

func TestPostgreSQL(t *testing.T) {
	t.Skip("postgresql support is currently broken")
	RunPostgreSQLTest(t, testCreateUser)
	RunPostgreSQLTest(t, testCreateUserUsernameUnique)
}

func TestSQLite(t *testing.T) {
	RunSQLiteTest(t, testCreateUser)
	RunSQLiteTest(t, testCreateUserUsernameUnique)
}

func testCreateUser(t *testing.T, db dbmodels.DBTX) {
	ctx := context.Background()

	queries := dbmodels.New(db)
	user1, err := queries.CreatePlayer(ctx, dbmodels.CreatePlayerParams{
		Username:     "test",
		Nickname:     sql.NullString{String: "testnick", Valid: true},
		PasswordHash: "xxx",
	})
	assert.NoError(t, err)
	user2, err := queries.CreatePlayer(ctx, dbmodels.CreatePlayerParams{
		Username:     "test2",
		Nickname:     sql.NullString{String: "testnick2", Valid: true},
		PasswordHash: "xxx",
	})
	assert.NoError(t, err)
	assert.NotEqual(t, user1.PlayerID, user2.PlayerID)
}

func testCreateUserUsernameUnique(t *testing.T, db dbmodels.DBTX) {
	ctx := context.Background()

	queries := dbmodels.New(db)
	_, err := queries.CreatePlayer(ctx, dbmodels.CreatePlayerParams{
		Username:     "test",
		Nickname:     sql.NullString{String: "testnick", Valid: true},
		PasswordHash: "xxx",
	})
	assert.NoError(t, err)
	_, err = queries.CreatePlayer(ctx, dbmodels.CreatePlayerParams{
		Username:     "test",
		Nickname:     sql.NullString{String: "testnick", Valid: true},
		PasswordHash: "xxx",
	})
	assert.ErrorContains(t, err, "UNIQUE")
}
