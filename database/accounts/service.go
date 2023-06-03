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

package accounts

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"time"

	"github.com/go-restruct/restruct"
	"github.com/google/uuid"
	"github.com/pangbox/server/common/hash"
	"github.com/pangbox/server/gen/dbmodels"
	"github.com/pangbox/server/pangya"
)

// Enumeration of possible errors that can be returned from authenticate.
var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrUnknownUsername = errors.New("unknown user")
)

const sessionTimeout = 15 * time.Minute

// Options specifies options for account services.
type Options struct {
	Database dbmodels.DBTX
	Hasher   hash.Hasher
}

// Service implements account services using the database.
type Service struct {
	database *dbmodels.Queries
	hasher   hash.Hasher
}

// NewService creates new account services using the database.
func NewService(opts Options) *Service {
	return &Service{
		database: dbmodels.New(opts.Database),
		hasher:   opts.Hasher,
	}
}

func (s *Service) GetPlayer(ctx context.Context, playerID int64) (dbmodels.Player, error) {
	return s.database.GetPlayer(ctx, playerID)
}

func (s *Service) Register(ctx context.Context, username, password string) (dbmodels.Player, error) {
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return dbmodels.Player{}, err
	}
	return s.database.CreatePlayer(ctx, dbmodels.CreatePlayerParams{
		Username:     username,
		PasswordHash: hash,
	})
}

// Authenticate authenticates a user using the database.
func (s *Service) Authenticate(ctx context.Context, username, password string) (dbmodels.Player, error) {
	player, err := s.database.GetPlayerByUsername(ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		return dbmodels.Player{}, ErrUnknownUsername
	} else if err != nil {
		return dbmodels.Player{}, err
	}
	if !s.hasher.CheckHash(password, player.PasswordHash) {
		return dbmodels.Player{}, ErrInvalidPassword
	}
	return player, nil
}

// SetNickname sets the player's nickname.
func (s *Service) SetNickname(ctx context.Context, playerID int64, nickname string) (dbmodels.Player, error) {
	return s.database.SetPlayerNickname(ctx, dbmodels.SetPlayerNicknameParams{
		PlayerID: playerID,
		Nickname: sql.NullString{Valid: true, String: nickname},
	})
}

// HasCharacters returns whether or not the player has characters.
func (s *Service) HasCharacters(ctx context.Context, playerID int64) (bool, error) {
	return s.database.PlayerHasCharacters(ctx, playerID)
}

// GetCharacters returns the characters for a given player.
func (s *Service) GetCharacters(ctx context.Context, playerID int64) ([]pangya.PlayerCharacterData, error) {
	characters, err := s.database.GetCharactersByPlayer(ctx, playerID)
	if err != nil {
		return nil, err
	}
	result := make([]pangya.PlayerCharacterData, len(characters))
	for i, character := range characters {
		if err := restruct.Unpack(character.CharacterData, binary.LittleEndian, &result[i]); err != nil {
			return nil, err
		}
		result[i].ID = uint32(character.CharacterID)
	}
	return result, nil
}

// AddCharacter adds a character for a given player.
func (s *Service) AddCharacter(ctx context.Context, playerID int64, data pangya.PlayerCharacterData) error {
	blob, err := restruct.Pack(binary.LittleEndian, &data)
	if err != nil {
		return err
	}
	_, err = s.database.CreateCharacter(ctx, dbmodels.CreateCharacterParams{
		PlayerID:        playerID,
		CharacterTypeID: int64(data.CharTypeID),
		CharacterData:   blob,
	})
	return err
}

// AddSession adds a new session for a player.
func (s *Service) AddSession(ctx context.Context, playerID int64, address string) (dbmodels.Session, error) {
	sessionKey, err := uuid.NewRandom()
	if err != nil {
		return dbmodels.Session{}, err
	}
	return s.database.CreateSession(ctx, dbmodels.CreateSessionParams{
		PlayerID:         playerID,
		SessionKey:       sessionKey.String(),
		SessionAddress:   address,
		SessionExpiresAt: time.Now().Add(sessionTimeout).Unix(),
	})
}

// GetSession gets a session by its ID.
func (s *Service) GetSession(ctx context.Context, sessionID int64) (dbmodels.Session, error) {
	return s.database.GetSession(ctx, sessionID)
}

// GetSessionByKey gets a session by its session key.
func (s *Service) GetSessionByKey(ctx context.Context, sessionKey string) (dbmodels.Session, error) {
	return s.database.GetSessionByKey(ctx, sessionKey)
}

// UpdateSessionExpiry bumps the session expiry value for a session.
func (s *Service) UpdateSessionExpiry(ctx context.Context, sessionID int64) (dbmodels.Session, error) {
	return s.database.UpdateSessionExpiry(ctx, dbmodels.UpdateSessionExpiryParams{
		SessionID:        sessionID,
		SessionExpiresAt: time.Now().Add(sessionTimeout).Unix(),
	})
}

// DeleteExpiredSessions deletes sessions that have expired.
func (s *Service) DeleteExpiredSessions(ctx context.Context) error {
	return s.database.DeleteExpiredSessions(ctx, time.Now().Unix())
}
