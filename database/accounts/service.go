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
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pangbox/server/common/hash"
	"github.com/pangbox/server/gen/dbmodels"
	"github.com/pangbox/server/pangya"
	"github.com/rs/zerolog"
)

// Enumeration of possible errors that can be returned from authenticate.
var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrUnknownUsername = errors.New("unknown user")
)

const sessionTimeout = 15 * time.Minute

// Options specifies options for account services.
type Options struct {
	Logger   zerolog.Logger
	Database *sql.DB
	Hasher   hash.Hasher
}

// Service implements account services using the database.
type Service struct {
	log     zerolog.Logger
	db      *sql.DB
	queries *dbmodels.Queries
	hasher  hash.Hasher
}

// NewService creates new account services using the database.
func NewService(opts Options) *Service {
	return &Service{
		log:     opts.Logger,
		db:      opts.Database,
		queries: dbmodels.New(opts.Database),
		hasher:  opts.Hasher,
	}
}

func (s *Service) GetPlayer(ctx context.Context, playerID int64) (dbmodels.GetPlayerRow, error) {
	return s.queries.GetPlayer(ctx, playerID)
}

func (s *Service) Register(ctx context.Context, username, password string) (dbmodels.Player, error) {
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return dbmodels.Player{}, err
	}
	return s.queries.CreatePlayer(ctx, dbmodels.CreatePlayerParams{
		Username:     username,
		PasswordHash: hash,
		Pang:         20000, // TODO
	})
}

// Authenticate authenticates a user using the database.
func (s *Service) Authenticate(ctx context.Context, username, password string) (dbmodels.Player, error) {
	player, err := s.queries.GetPlayerByUsername(ctx, username)
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
	return s.queries.SetPlayerNickname(ctx, dbmodels.SetPlayerNicknameParams{
		PlayerID: playerID,
		Nickname: sql.NullString{Valid: true, String: nickname},
	})
}

// HasCharacters returns whether or not the player has characters.
func (s *Service) HasCharacters(ctx context.Context, playerID int64) (bool, error) {
	return s.queries.PlayerHasCharacters(ctx, playerID)
}

// GetCharacters returns the characters for a given player.
func (s *Service) GetCharacters(ctx context.Context, playerID int64) ([]pangya.PlayerCharacterData, error) {
	characters, err := s.queries.GetCharactersByPlayer(ctx, playerID)
	if err != nil {
		return nil, err
	}
	result := make([]pangya.PlayerCharacterData, len(characters))
	for i, character := range characters {
		result[i] = pangya.PlayerCharacterData{
			CharTypeID: uint32(character.CharacterTypeID),
			ID:         uint32(character.CharacterID),
			HairColor:  uint8(character.HairColor),
			Shirt:      uint8(character.Shirt),
			PartTypeIDs: [24]uint32{
				uint32(character.Part00ItemTypeID), uint32(character.Part01ItemTypeID), uint32(character.Part02ItemTypeID), uint32(character.Part03ItemTypeID),
				uint32(character.Part04ItemTypeID), uint32(character.Part05ItemTypeID), uint32(character.Part06ItemTypeID), uint32(character.Part07ItemTypeID),
				uint32(character.Part08ItemTypeID), uint32(character.Part09ItemTypeID), uint32(character.Part10ItemTypeID), uint32(character.Part11ItemTypeID),
				uint32(character.Part12ItemTypeID), uint32(character.Part13ItemTypeID), uint32(character.Part14ItemTypeID), uint32(character.Part15ItemTypeID),
				uint32(character.Part16ItemTypeID), uint32(character.Part17ItemTypeID), uint32(character.Part18ItemTypeID), uint32(character.Part19ItemTypeID),
				uint32(character.Part20ItemTypeID), uint32(character.Part21ItemTypeID), uint32(character.Part22ItemTypeID), uint32(character.Part23ItemTypeID),
			},
			PartIDs: [24]uint32{
				uint32(character.Part00ItemID.Int64), uint32(character.Part01ItemID.Int64), uint32(character.Part02ItemID.Int64), uint32(character.Part03ItemID.Int64),
				uint32(character.Part04ItemID.Int64), uint32(character.Part05ItemID.Int64), uint32(character.Part06ItemID.Int64), uint32(character.Part07ItemID.Int64),
				uint32(character.Part08ItemID.Int64), uint32(character.Part09ItemID.Int64), uint32(character.Part10ItemID.Int64), uint32(character.Part11ItemID.Int64),
				uint32(character.Part12ItemID.Int64), uint32(character.Part13ItemID.Int64), uint32(character.Part14ItemID.Int64), uint32(character.Part15ItemID.Int64),
				uint32(character.Part16ItemID.Int64), uint32(character.Part17ItemID.Int64), uint32(character.Part18ItemID.Int64), uint32(character.Part19ItemID.Int64),
				uint32(character.Part20ItemID.Int64), uint32(character.Part21ItemID.Int64), uint32(character.Part22ItemID.Int64), uint32(character.Part23ItemID.Int64),
			},
			AuxParts: [5]uint32{
				uint32(character.AuxPart0ID.Int64),
				uint32(character.AuxPart1ID.Int64),
				uint32(character.AuxPart2ID.Int64),
				uint32(character.AuxPart3ID.Int64),
				uint32(character.AuxPart4ID.Int64),
			},
			CutInID: uint32(character.CutInID.Int64),
		}
	}
	return result, nil
}

type NewCharacterParams struct {
	CharTypeID         uint32
	HairColor          uint8
	Shirt              uint8
	Mastery            uint32
	DefaultPartTypeIDs [24]uint32
}

// AddCharacter adds a character for a given player.
func (s *Service) AddCharacter(ctx context.Context, playerID int64, params NewCharacterParams) (dbmodels.Character, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dbmodels.Character{}, err
	}
	defer tx.Rollback()

	queries := s.queries.WithTx(tx)

	item, err := queries.AddItemToInventory(ctx, dbmodels.AddItemToInventoryParams{
		PlayerID:   playerID,
		ItemTypeID: int64(params.CharTypeID),
	})
	if err != nil {
		return dbmodels.Character{}, err
	}

	character, err := queries.CreateCharacter(ctx, dbmodels.CreateCharacterParams{
		PlayerID:         playerID,
		ItemID:           item.ItemID,
		HairColor:        int64(params.HairColor),
		Shirt:            int64(params.Shirt),
		Mastery:          int64(params.Mastery),
		Part00ItemTypeID: int64(params.DefaultPartTypeIDs[0]),
		Part01ItemTypeID: int64(params.DefaultPartTypeIDs[1]),
		Part02ItemTypeID: int64(params.DefaultPartTypeIDs[2]),
		Part03ItemTypeID: int64(params.DefaultPartTypeIDs[3]),
		Part04ItemTypeID: int64(params.DefaultPartTypeIDs[4]),
		Part05ItemTypeID: int64(params.DefaultPartTypeIDs[5]),
		Part06ItemTypeID: int64(params.DefaultPartTypeIDs[6]),
		Part07ItemTypeID: int64(params.DefaultPartTypeIDs[7]),
		Part08ItemTypeID: int64(params.DefaultPartTypeIDs[8]),
		Part09ItemTypeID: int64(params.DefaultPartTypeIDs[9]),
		Part10ItemTypeID: int64(params.DefaultPartTypeIDs[10]),
		Part11ItemTypeID: int64(params.DefaultPartTypeIDs[11]),
		Part12ItemTypeID: int64(params.DefaultPartTypeIDs[12]),
		Part13ItemTypeID: int64(params.DefaultPartTypeIDs[13]),
		Part14ItemTypeID: int64(params.DefaultPartTypeIDs[14]),
		Part15ItemTypeID: int64(params.DefaultPartTypeIDs[15]),
		Part16ItemTypeID: int64(params.DefaultPartTypeIDs[16]),
		Part17ItemTypeID: int64(params.DefaultPartTypeIDs[17]),
		Part18ItemTypeID: int64(params.DefaultPartTypeIDs[18]),
		Part19ItemTypeID: int64(params.DefaultPartTypeIDs[19]),
		Part20ItemTypeID: int64(params.DefaultPartTypeIDs[20]),
		Part21ItemTypeID: int64(params.DefaultPartTypeIDs[21]),
		Part22ItemTypeID: int64(params.DefaultPartTypeIDs[22]),
		Part23ItemTypeID: int64(params.DefaultPartTypeIDs[23]),
	})

	if err != nil {
		return dbmodels.Character{}, err
	}

	err = tx.Commit()
	if err != nil {
		return dbmodels.Character{}, err
	}

	return character, nil
}

func (s *Service) AddClubSet(ctx context.Context, playerID int64, clubsetTypeID uint32) (dbmodels.Inventory, error) {
	return s.queries.AddItemToInventory(ctx, dbmodels.AddItemToInventoryParams{
		PlayerID:   playerID,
		ItemTypeID: int64(clubsetTypeID),
	})
}

func (s *Service) SetCharacter(ctx context.Context, playerID int64, characterID int64) error {
	_, err := s.queries.SetPlayerCharacter(ctx, dbmodels.SetPlayerCharacterParams{
		PlayerID:    playerID,
		CharacterID: sql.NullInt64{Valid: true, Int64: characterID},
	})
	return err
}

func (s *Service) SetCaddie(ctx context.Context, playerID int64, caddieID int64) error {
	_, err := s.queries.SetPlayerCaddie(ctx, dbmodels.SetPlayerCaddieParams{
		PlayerID: playerID,
		CaddieID: sql.NullInt64{Valid: true, Int64: caddieID},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) getConsumablesWith(ctx context.Context, tx *dbmodels.Queries, playerID int64) ([10]uint32, error) {
	row, err := tx.GetPlayerConsumables(ctx, playerID)
	if err != nil {
		return [10]uint32{}, err
	}
	return [10]uint32{
		uint32(row.Slot0TypeID),
		uint32(row.Slot1TypeID),
		uint32(row.Slot2TypeID),
		uint32(row.Slot3TypeID),
		uint32(row.Slot4TypeID),
		uint32(row.Slot5TypeID),
		uint32(row.Slot6TypeID),
		uint32(row.Slot7TypeID),
		uint32(row.Slot8TypeID),
		uint32(row.Slot9TypeID),
	}, nil
}

func (s *Service) setConsumablesWith(ctx context.Context, tx *dbmodels.Queries, playerID int64, newSlots [10]uint32) (dbmodels.Player, error) {
	ret, err := tx.SetPlayerConsumables(ctx, dbmodels.SetPlayerConsumablesParams{
		PlayerID:    playerID,
		Slot0TypeID: int64(newSlots[0]),
		Slot1TypeID: int64(newSlots[1]),
		Slot2TypeID: int64(newSlots[2]),
		Slot3TypeID: int64(newSlots[3]),
		Slot4TypeID: int64(newSlots[4]),
		Slot5TypeID: int64(newSlots[5]),
		Slot6TypeID: int64(newSlots[6]),
		Slot7TypeID: int64(newSlots[7]),
		Slot8TypeID: int64(newSlots[8]),
		Slot9TypeID: int64(newSlots[9]),
	})
	if err != nil {
		return dbmodels.Player{}, err
	}

	return ret, nil
}

func (s *Service) updateConsumables(player *dbmodels.GetPlayerRow, ret dbmodels.Player) {
	if player != nil {
		player.Slot0TypeID = ret.Slot0TypeID
		player.Slot1TypeID = ret.Slot1TypeID
		player.Slot2TypeID = ret.Slot2TypeID
		player.Slot3TypeID = ret.Slot3TypeID
		player.Slot4TypeID = ret.Slot4TypeID
		player.Slot5TypeID = ret.Slot5TypeID
		player.Slot6TypeID = ret.Slot6TypeID
		player.Slot7TypeID = ret.Slot7TypeID
		player.Slot8TypeID = ret.Slot8TypeID
		player.Slot9TypeID = ret.Slot9TypeID
	}
}

func (s *Service) decrementConsumableQuantityWith(ctx context.Context, tx *dbmodels.Queries, playerID, itemTypeID int64) (int64, error) {
	items, err := tx.GetItemsByTypeID(ctx, dbmodels.GetItemsByTypeIDParams{
		PlayerID:   playerID,
		ItemTypeID: itemTypeID,
	})
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, fmt.Errorf("item %08x not in inventory", itemTypeID)
	} else if len(items) != 1 {
		return 0, errors.New("cardinality error")
	}
	item := items[0]
	quantity := item.Quantity.Int64
	if !item.Quantity.Valid || quantity == 0 {
		// Items without quantities are not consumable and can't be equipped here.
		return 0, errors.New("invalid item quantity in inventory")
	} else if quantity > 1 {
		newValue, err := tx.SetItemQuantity(ctx, dbmodels.SetItemQuantityParams{
			PlayerID: playerID,
			ItemID:   item.ItemID,
			Quantity: sql.NullInt64{Int64: quantity - 1, Valid: true},
		})
		s.log.Debug().
			Int64("old quantity", quantity).
			Int64("new quantity", newValue.Quantity.Int64).
			Msg("decrement consumable quantity")
		if err != nil {
			return 0, err
		}
		return item.ItemID, nil
	} else {
		s.log.Debug().Msg("remove consumable")
		return item.ItemID, tx.RemoveItemFromInventory(ctx, dbmodels.RemoveItemFromInventoryParams{
			PlayerID: playerID,
			ItemID:   item.ItemID,
		})
	}
}

func (s *Service) incrementConsumableQuantityWith(ctx context.Context, tx *dbmodels.Queries, playerID, itemTypeID, amount int64) error {
	items, err := tx.GetItemsByTypeID(ctx, dbmodels.GetItemsByTypeIDParams{
		PlayerID:   playerID,
		ItemTypeID: itemTypeID,
	})
	if err != nil {
		return err
	}
	if len(items) > 1 {
		return errors.New("cardinality error")
	} else if len(items) == 0 {
		_, err = tx.AddItemToInventory(ctx, dbmodels.AddItemToInventoryParams{
			PlayerID:   playerID,
			ItemTypeID: itemTypeID,
			Quantity:   sql.NullInt64{Valid: true, Int64: amount},
		})
		if err != nil {
			s.log.Error().Err(err).Msg("adding item to inventory")
		}
		return nil
	}
	item := items[0]
	if !item.Quantity.Valid || item.Quantity.Int64 == 0 {
		// Items without quantities are not consumable and can't be equipped here.
		return errors.New("invalid item quantity in inventory")
	}
	quantity := item.Quantity.Int64
	_, err = tx.SetItemQuantity(ctx, dbmodels.SetItemQuantityParams{
		PlayerID: playerID,
		ItemID:   item.ItemID,
		Quantity: sql.NullInt64{Int64: quantity + amount, Valid: true},
	})
	return err
}

func (s *Service) GetConsumables(ctx context.Context, playerID int64) ([10]uint32, error) {
	return s.getConsumablesWith(ctx, s.queries, playerID)
}

func (s *Service) SetConsumables(ctx context.Context, playerID int64, newSlots [10]uint32, player *dbmodels.GetPlayerRow) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := s.queries.WithTx(tx)

	// TODO: check that there is enough in the inventory

	ret, err := queries.SetPlayerConsumables(ctx, dbmodels.SetPlayerConsumablesParams{
		PlayerID:    playerID,
		Slot0TypeID: int64(newSlots[0]),
		Slot1TypeID: int64(newSlots[1]),
		Slot2TypeID: int64(newSlots[2]),
		Slot3TypeID: int64(newSlots[3]),
		Slot4TypeID: int64(newSlots[4]),
		Slot5TypeID: int64(newSlots[5]),
		Slot6TypeID: int64(newSlots[6]),
		Slot7TypeID: int64(newSlots[7]),
		Slot8TypeID: int64(newSlots[8]),
		Slot9TypeID: int64(newSlots[9]),
	})
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if player != nil {
		player.Slot0TypeID = ret.Slot0TypeID
		player.Slot1TypeID = ret.Slot1TypeID
		player.Slot2TypeID = ret.Slot2TypeID
		player.Slot3TypeID = ret.Slot3TypeID
		player.Slot4TypeID = ret.Slot4TypeID
		player.Slot5TypeID = ret.Slot5TypeID
		player.Slot6TypeID = ret.Slot6TypeID
		player.Slot7TypeID = ret.Slot7TypeID
		player.Slot8TypeID = ret.Slot8TypeID
		player.Slot9TypeID = ret.Slot9TypeID
	}

	return nil
}

func (s *Service) SetComet(ctx context.Context, playerID int64, cometTypeID uint32, player *dbmodels.GetPlayerRow) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	queries := s.queries.WithTx(tx)

	cometID := int64(0)
	if cometTypeID != 0 {
		cometID, err = s.decrementConsumableQuantityWith(ctx, queries, playerID, int64(cometTypeID))
		if err != nil {
			return 0, fmt.Errorf("set comet: %w", err)
		}
	}

	ret, err := queries.SetPlayerComet(ctx, dbmodels.SetPlayerCometParams{
		PlayerID:   playerID,
		BallTypeID: int64(cometTypeID),
	})
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	if player != nil {
		player.BallTypeID = ret.BallTypeID
	}

	return cometID, nil
}

func (s *Service) SetClubSet(ctx context.Context, playerID int64, clubsetID int64) error {
	_, err := s.queries.SetPlayerClubSet(ctx, dbmodels.SetPlayerClubSetParams{
		PlayerID: playerID,
		ClubID:   sql.NullInt64{Valid: true, Int64: clubsetID},
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PurchaseItem(ctx context.Context, playerID, pangTotal, pointTotal, itemTypeID, quantity int64) (dbmodels.SetPlayerCurrencyRow, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dbmodels.SetPlayerCurrencyRow{}, err
	}
	defer tx.Rollback()

	queries := s.queries.WithTx(tx)

	currency, err := queries.GetPlayerCurrency(ctx, playerID)
	if err != nil {
		return dbmodels.SetPlayerCurrencyRow{}, fmt.Errorf("getting player currency: %w", err)
	}

	if quantity != 0 {
		if err := s.incrementConsumableQuantityWith(ctx, queries, playerID, itemTypeID, quantity); err != nil {
			return dbmodels.SetPlayerCurrencyRow{}, fmt.Errorf("adding consumable quantity to inventory: %w", err)
		}
	} else {
		item, err := queries.AddItemToInventory(ctx, dbmodels.AddItemToInventoryParams{
			PlayerID:   playerID,
			ItemTypeID: itemTypeID,
		})
		if err != nil {
			return dbmodels.SetPlayerCurrencyRow{}, fmt.Errorf("adding item to inventory: %w", err)
		}
		// TODO: should use IFF data
		if itemTypeID >= 0x4000000 && itemTypeID < 0x40000FF {
			if _, err := queries.CreateCharacter(ctx, dbmodels.CreateCharacterParams{
				PlayerID: playerID,
				ItemID:   item.ItemID,
			}); err != nil {
				return dbmodels.SetPlayerCurrencyRow{}, fmt.Errorf("creating character entry: %w", err)
			}
		}
	}

	newCurrency, err := queries.SetPlayerCurrency(ctx, dbmodels.SetPlayerCurrencyParams{
		PlayerID: playerID,
		Pang:     currency.Pang - pangTotal,
		Points:   currency.Points - pointTotal,
	})
	if err != nil {
		return dbmodels.SetPlayerCurrencyRow{}, fmt.Errorf("setting player currency: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return dbmodels.SetPlayerCurrencyRow{}, err
	}

	return newCurrency, nil
}

func (s *Service) UseItem(ctx context.Context, playerID, itemTypeID int64, player *dbmodels.GetPlayerRow) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := s.queries.WithTx(tx)

	_, err = s.decrementConsumableQuantityWith(ctx, queries, playerID, itemTypeID)
	if err != nil {
		return err
	}

	slots, err := s.getConsumablesWith(ctx, queries, playerID)
	if err != nil {
		return err
	}
	for i := 9; i >= 0; i-- {
		if slots[i] == uint32(itemTypeID) {
			slots[i] = 0
			break
		}
	}
	ret, err := s.setConsumablesWith(ctx, queries, playerID, slots)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	s.updateConsumables(player, ret)

	return nil
}

func (s *Service) AddPang(ctx context.Context, playerID, pang int64) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	queries := s.queries.WithTx(tx)

	currency, err := queries.GetPlayerCurrency(ctx, playerID)
	if err != nil {
		return 0, err
	}

	newCurrency, err := queries.SetPlayerCurrency(ctx, dbmodels.SetPlayerCurrencyParams{
		PlayerID: playerID,
		Pang:     currency.Pang + pang,
		Points:   currency.Points,
	})
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return newCurrency.Pang, nil
}

type DecorationTypeIDs struct {
	BackgroundTypeID uint32
	FrameTypeID      uint32
	StickerTypeID    uint32
	SlotTypeID       uint32
	CutInTypeID      uint32
	TitleTypeID      uint32
}

func (s *Service) getDecorationIDWith(ctx context.Context, tx *dbmodels.Queries, playerID int64, typeID uint32) (sql.NullInt64, error) {
	if typeID == 0 {
		return sql.NullInt64{}, nil
	}
	items, err := tx.GetItemsByTypeID(ctx, dbmodels.GetItemsByTypeIDParams{
		PlayerID:   playerID,
		ItemTypeID: int64(typeID),
	})
	if err != nil {
		return sql.NullInt64{}, err
	}
	if len(items) == 0 {
		s.log.Warn().Msgf("missing decoration 0x%08x", typeID)
		return sql.NullInt64{}, nil
	}
	return sql.NullInt64{Valid: true, Int64: items[0].ItemID}, nil
}

func (s *Service) SetDecoration(ctx context.Context, playerId int64, decoration DecorationTypeIDs) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := s.queries.WithTx(tx)

	backgroundID, err := s.getDecorationIDWith(ctx, queries, playerId, decoration.BackgroundTypeID)
	if err != nil {
		return err
	}

	frameID, err := s.getDecorationIDWith(ctx, queries, playerId, decoration.FrameTypeID)
	if err != nil {
		return err
	}

	stickerID, err := s.getDecorationIDWith(ctx, queries, playerId, decoration.StickerTypeID)
	if err != nil {
		return err
	}

	slotID, err := s.getDecorationIDWith(ctx, queries, playerId, decoration.SlotTypeID)
	if err != nil {
		return err
	}

	cutInID, err := s.getDecorationIDWith(ctx, queries, playerId, decoration.CutInTypeID)
	if err != nil {
		return err
	}

	titleID, err := s.getDecorationIDWith(ctx, queries, playerId, decoration.TitleTypeID)
	if err != nil {
		return err
	}

	_, err = queries.SetPlayerDecoration(ctx, dbmodels.SetPlayerDecorationParams{
		BackgroundID: backgroundID,
		FrameID:      frameID,
		StickerID:    stickerID,
		SlotID:       slotID,
		CutInID:      cutInID,
		TitleID:      titleID,
		PlayerID:     playerId,
	})
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) getPartIDWith(ctx context.Context, tx *dbmodels.Queries, playerID int64, itemID uint32) (sql.NullInt64, error) {
	if itemID == 0 {
		return sql.NullInt64{}, nil
	}
	item, err := tx.GetItem(ctx, dbmodels.GetItemParams{
		PlayerID: playerID,
		ItemID:   int64(itemID),
	})
	if err != nil {
		return sql.NullInt64{}, err
	}
	return sql.NullInt64{Valid: true, Int64: item.ItemID}, nil
}

func (s *Service) SetCharacterParts(ctx context.Context, playerID int64, data pangya.PlayerCharacterData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := s.queries.WithTx(tx)

	part00Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[0])
	if err != nil {
		return err
	}
	part01Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[1])
	if err != nil {
		return err
	}
	part02Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[2])
	if err != nil {
		return err
	}
	part03Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[3])
	if err != nil {
		return err
	}
	part04Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[4])
	if err != nil {
		return err
	}
	part05Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[5])
	if err != nil {
		return err
	}
	part06Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[6])
	if err != nil {
		return err
	}
	part07Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[7])
	if err != nil {
		return err
	}
	part08Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[8])
	if err != nil {
		return err
	}
	part09Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[9])
	if err != nil {
		return err
	}
	part10Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[10])
	if err != nil {
		return err
	}
	part11Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[11])
	if err != nil {
		return err
	}
	part12Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[12])
	if err != nil {
		return err
	}
	part13Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[13])
	if err != nil {
		return err
	}
	part14Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[14])
	if err != nil {
		return err
	}
	part15Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[15])
	if err != nil {
		return err
	}
	part16Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[16])
	if err != nil {
		return err
	}
	part17Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[17])
	if err != nil {
		return err
	}
	part18Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[18])
	if err != nil {
		return err
	}
	part19Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[19])
	if err != nil {
		return err
	}
	part20Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[20])
	if err != nil {
		return err
	}
	part21Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[21])
	if err != nil {
		return err
	}
	part22Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[22])
	if err != nil {
		return err
	}
	part23Item, err := s.getPartIDWith(ctx, queries, playerID, data.PartIDs[23])
	if err != nil {
		return err
	}
	cutIn, err := s.getPartIDWith(ctx, queries, playerID, data.CutInID)
	if err != nil {
		return err
	}

	queries.SetCharacterParts(ctx, dbmodels.SetCharacterPartsParams{
		CharacterID:      int64(data.ID),
		Part00ItemID:     part00Item,
		Part01ItemID:     part01Item,
		Part02ItemID:     part02Item,
		Part03ItemID:     part03Item,
		Part04ItemID:     part04Item,
		Part05ItemID:     part05Item,
		Part06ItemID:     part06Item,
		Part07ItemID:     part07Item,
		Part08ItemID:     part08Item,
		Part09ItemID:     part09Item,
		Part10ItemID:     part10Item,
		Part11ItemID:     part11Item,
		Part12ItemID:     part12Item,
		Part13ItemID:     part13Item,
		Part14ItemID:     part14Item,
		Part15ItemID:     part15Item,
		Part16ItemID:     part16Item,
		Part17ItemID:     part17Item,
		Part18ItemID:     part18Item,
		Part19ItemID:     part19Item,
		Part20ItemID:     part20Item,
		Part21ItemID:     part21Item,
		Part22ItemID:     part22Item,
		Part23ItemID:     part23Item,
		Part00ItemTypeID: int64(data.PartTypeIDs[0]),
		Part01ItemTypeID: int64(data.PartTypeIDs[1]),
		Part02ItemTypeID: int64(data.PartTypeIDs[2]),
		Part03ItemTypeID: int64(data.PartTypeIDs[3]),
		Part04ItemTypeID: int64(data.PartTypeIDs[4]),
		Part05ItemTypeID: int64(data.PartTypeIDs[5]),
		Part06ItemTypeID: int64(data.PartTypeIDs[6]),
		Part07ItemTypeID: int64(data.PartTypeIDs[7]),
		Part08ItemTypeID: int64(data.PartTypeIDs[8]),
		Part09ItemTypeID: int64(data.PartTypeIDs[9]),
		Part10ItemTypeID: int64(data.PartTypeIDs[10]),
		Part11ItemTypeID: int64(data.PartTypeIDs[11]),
		Part12ItemTypeID: int64(data.PartTypeIDs[12]),
		Part13ItemTypeID: int64(data.PartTypeIDs[13]),
		Part14ItemTypeID: int64(data.PartTypeIDs[14]),
		Part15ItemTypeID: int64(data.PartTypeIDs[15]),
		Part16ItemTypeID: int64(data.PartTypeIDs[16]),
		Part17ItemTypeID: int64(data.PartTypeIDs[17]),
		Part18ItemTypeID: int64(data.PartTypeIDs[18]),
		Part19ItemTypeID: int64(data.PartTypeIDs[19]),
		Part20ItemTypeID: int64(data.PartTypeIDs[20]),
		Part21ItemTypeID: int64(data.PartTypeIDs[21]),
		Part22ItemTypeID: int64(data.PartTypeIDs[22]),
		Part23ItemTypeID: int64(data.PartTypeIDs[23]),
		CutInID:          cutIn,
	})

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}

// AddSession adds a new session for a player.
func (s *Service) AddSession(ctx context.Context, playerID int64, address string) (dbmodels.Session, error) {
	sessionKey, err := uuid.NewRandom()
	if err != nil {
		return dbmodels.Session{}, err
	}
	return s.queries.CreateSession(ctx, dbmodels.CreateSessionParams{
		PlayerID:         playerID,
		SessionKey:       sessionKey.String(),
		SessionAddress:   address,
		SessionExpiresAt: time.Now().Add(sessionTimeout).Unix(),
	})
}

// GetSession gets a session by its ID.
func (s *Service) GetSession(ctx context.Context, sessionID int64) (dbmodels.Session, error) {
	return s.queries.GetSession(ctx, sessionID)
}

// GetSessionByKey gets a session by its session key.
func (s *Service) GetSessionByKey(ctx context.Context, sessionKey string) (dbmodels.Session, error) {
	return s.queries.GetSessionByKey(ctx, sessionKey)
}

// UpdateSessionExpiry bumps the session expiry value for a session.
func (s *Service) UpdateSessionExpiry(ctx context.Context, sessionID int64) (dbmodels.Session, error) {
	return s.queries.UpdateSessionExpiry(ctx, dbmodels.UpdateSessionExpiryParams{
		SessionID:        sessionID,
		SessionExpiresAt: time.Now().Add(sessionTimeout).Unix(),
	})
}

// DeleteExpiredSessions deletes sessions that have expired.
func (s *Service) DeleteExpiredSessions(ctx context.Context) error {
	return s.queries.DeleteExpiredSessions(ctx, time.Now().Unix())
}

func (s *Service) GetPlayerInventory(ctx context.Context, playerID int64) ([]dbmodels.Inventory, error) {
	return s.queries.GetPlayerInventory(ctx, playerID)
}

func (s *Service) AddExp(ctx context.Context, playerID int64, add int) (pangya.Rank, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	queries := s.queries.WithTx(tx)

	rank, err := queries.GetPlayerRank(ctx, playerID)
	if err != nil {
		return 0, 0, err
	}

	newRank, newExp := pangya.AddExperience(pangya.Rank(rank.Rank), int(rank.Exp), add)
	values, err := queries.SetPlayerRank(ctx, dbmodels.SetPlayerRankParams{
		PlayerID: playerID,
		Rank:     int64(newRank),
		Exp:      int64(newExp),
	})
	if err != nil {
		return 0, 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, 0, err
	}

	return pangya.Rank(values.Rank), int(values.Exp), nil
}
