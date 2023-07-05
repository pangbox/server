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

package gameserver

import (
	"github.com/pangbox/server/gen/dbmodels"
	"github.com/pangbox/server/pangya"
)

func playerInfoFromDB(player *dbmodels.GetPlayerRow, connID uint32) pangya.PlayerInfo {
	return pangya.PlayerInfo{
		Username: player.Username,
		Nickname: player.Nickname.String,
		PlayerID: uint32(player.PlayerID),
		ConnID:   connID,
		// TODO
	}
}

func playerStatsFromDB(player *dbmodels.GetPlayerRow) pangya.PlayerStats {
	return pangya.PlayerStats{
		Pang:    uint64(player.Pang),
		Rank:    byte(player.Rank),
		TotalXP: uint32(player.Exp),
		// TODO
	}
}

func playerEquippedConsumablesFromDB(player *dbmodels.GetPlayerRow) [10]uint32 {
	return [10]uint32{
		uint32(player.Slot0TypeID),
		uint32(player.Slot1TypeID),
		uint32(player.Slot2TypeID),
		uint32(player.Slot3TypeID),
		uint32(player.Slot4TypeID),
		uint32(player.Slot5TypeID),
		uint32(player.Slot6TypeID),
		uint32(player.Slot7TypeID),
		uint32(player.Slot8TypeID),
		uint32(player.Slot9TypeID),
	}
}

func playerEquippedClubSetFromDB(player *dbmodels.GetPlayerRow) pangya.PlayerClubData {
	return pangya.PlayerClubData{
		Item: pangya.PlayerItem{
			ID:     uint32(player.ClubID.Int64),
			TypeID: uint32(player.ClubTypeID.Int64),
		},
		// TODO: stats/enchantments are not implemented.
		Stats: pangya.ClubStats{
			UpgradeStats: [5]uint16{8, 9, 8, 3, 3},
		},
	}
}

func playerEquippedItemsFromDB(player *dbmodels.GetPlayerRow) pangya.PlayerEquipment {
	comet := player.BallTypeID
	if comet == 0 {
		comet = 0x14000000
	}
	return pangya.PlayerEquipment{
		CaddieID:    uint32(player.CaddieID.Int64),
		CharacterID: uint32(player.CharacterID.Int64),
		ClubSetID:   uint32(player.ClubID.Int64),
		CometTypeID: uint32(comet),
		Items: pangya.PlayerEquippedItems{
			ItemIDs: playerEquippedConsumablesFromDB(player),
		},
		BackgroundID:     uint32(player.BackgroundID.Int64),
		FrameID:          uint32(player.FrameID.Int64),
		StickerID:        uint32(player.StickerID.Int64),
		SlotID:           uint32(player.SlotID.Int64),
		CutInID:          uint32(player.CutInID.Int64),
		TitleID:          uint32(player.TitleID.Int64),
		BackgroundTypeID: uint32(player.BackgroundTypeID.Int64),
		FrameTypeID:      uint32(player.FrameTypeID.Int64),
		StickerTypeID:    uint32(player.StickerTypeID.Int64),
		SlotTypeID:       uint32(player.SlotTypeID.Int64),
		CutInTypeID:      uint32(player.CutInTypeID.Int64),
		TitleTypeID:      uint32(player.TitleTypeID.Int64),
		MascotID:         uint32(player.MascotTypeID),
		PosterID: [2]uint32{
			uint32(player.Poster0TypeID.Int64),
			uint32(player.Poster1TypeID.Int64),
		},
	}
}

func playerEquippedCharacterFromDB(player *dbmodels.GetPlayerRow) pangya.PlayerCharacterData {
	return pangya.PlayerCharacterData{
		CharTypeID: uint32(player.CharacterTypeID.Int64),
		ID:         uint32(player.CharacterID.Int64),
		HairColor:  uint8(player.HairColor),
		Shirt:      uint8(player.Shirt),
		PartTypeIDs: [24]uint32{
			uint32(player.Part00ItemTypeID), uint32(player.Part01ItemTypeID), uint32(player.Part02ItemTypeID), uint32(player.Part03ItemTypeID),
			uint32(player.Part04ItemTypeID), uint32(player.Part05ItemTypeID), uint32(player.Part06ItemTypeID), uint32(player.Part07ItemTypeID),
			uint32(player.Part08ItemTypeID), uint32(player.Part09ItemTypeID), uint32(player.Part10ItemTypeID), uint32(player.Part11ItemTypeID),
			uint32(player.Part12ItemTypeID), uint32(player.Part13ItemTypeID), uint32(player.Part14ItemTypeID), uint32(player.Part15ItemTypeID),
			uint32(player.Part16ItemTypeID), uint32(player.Part17ItemTypeID), uint32(player.Part18ItemTypeID), uint32(player.Part19ItemTypeID),
			uint32(player.Part20ItemTypeID), uint32(player.Part21ItemTypeID), uint32(player.Part22ItemTypeID), uint32(player.Part23ItemTypeID),
		},
		PartIDs: [24]uint32{
			uint32(player.Part00ItemID.Int64), uint32(player.Part01ItemID.Int64), uint32(player.Part02ItemID.Int64), uint32(player.Part03ItemID.Int64),
			uint32(player.Part04ItemID.Int64), uint32(player.Part05ItemID.Int64), uint32(player.Part06ItemID.Int64), uint32(player.Part07ItemID.Int64),
			uint32(player.Part08ItemID.Int64), uint32(player.Part09ItemID.Int64), uint32(player.Part10ItemID.Int64), uint32(player.Part11ItemID.Int64),
			uint32(player.Part12ItemID.Int64), uint32(player.Part13ItemID.Int64), uint32(player.Part14ItemID.Int64), uint32(player.Part15ItemID.Int64),
			uint32(player.Part16ItemID.Int64), uint32(player.Part17ItemID.Int64), uint32(player.Part18ItemID.Int64), uint32(player.Part19ItemID.Int64),
			uint32(player.Part20ItemID.Int64), uint32(player.Part21ItemID.Int64), uint32(player.Part22ItemID.Int64), uint32(player.Part23ItemID.Int64),
		},
		AuxParts: [5]uint32{
			uint32(player.AuxPart0ID.Int64),
			uint32(player.AuxPart1ID.Int64),
			uint32(player.AuxPart2ID.Int64),
			uint32(player.AuxPart3ID.Int64),
			uint32(player.AuxPart4ID.Int64),
		},
		CutInID: uint32(player.CutInID.Int64),
	}
}

func playerDataFromDB(player *dbmodels.GetPlayerRow, connID uint32) pangya.PlayerData {
	return pangya.PlayerData{
		UserInfo:          playerInfoFromDB(player, connID),
		PlayerStats:       playerStatsFromDB(player),
		EquippedItems:     playerEquippedItemsFromDB(player),
		EquippedCharacter: playerEquippedCharacterFromDB(player),
		EquippedClub:      playerEquippedClubSetFromDB(player),
	}
}

func (c *Conn) getPlayerInfo() pangya.PlayerInfo {
	return playerInfoFromDB(&c.player, c.connID)
}

func (c *Conn) getPlayerStats() pangya.PlayerStats {
	return playerStatsFromDB(&c.player)
}

func (c *Conn) getPlayerEquippedConsumables() [10]uint32 {
	return playerEquippedConsumablesFromDB(&c.player)
}

func (c *Conn) getPlayerEquippedItems() pangya.PlayerEquipment {
	return playerEquippedItemsFromDB(&c.player)
}

func (c *Conn) getPlayerEquippedCharacter() pangya.PlayerCharacterData {
	return *c.currentCharacter
}

func (c *Conn) getPlayerEquippedClubSet() pangya.PlayerClubData {
	return playerEquippedClubSetFromDB(&c.player)
}

func (c *Conn) getPlayerData() pangya.PlayerData {
	return pangya.PlayerData{
		UserInfo:          c.getPlayerInfo(),
		PlayerStats:       c.getPlayerStats(),
		EquippedItems:     c.getPlayerEquippedItems(),
		EquippedCharacter: c.getPlayerEquippedCharacter(),
		EquippedClub:      c.getPlayerEquippedClubSet(),
	}
}
