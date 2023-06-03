// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package dbmodels

import (
	"database/sql"
)

type Character struct {
	CharacterID     int64
	PlayerID        int64
	CharacterTypeID int64
	CharacterData   []byte
}

type Player struct {
	PlayerID     int64
	Username     string
	Nickname     sql.NullString
	PasswordHash string
	Pang         int64
	Rank         int64
}

type Session struct {
	SessionID        int64
	PlayerID         int64
	SessionKey       string
	SessionAddress   string
	SessionExpiresAt int64
}
