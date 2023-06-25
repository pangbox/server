#!/bin/sh
go run github.com/kyleconroy/sqlc/cmd/sqlc generate --no-remote

# Workaround for LEFT JOIN nullability issues in sqlc.
sed -i -E 's/FIXNULL([[:space:]]+)int64/       \1sql.NullInt64/g; s/FIXNULL//g' gen/dbmodels/player.sql.go
