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

import "github.com/pangbox/server/pangya"

func (c *Conn) getPlayerInfo() pangya.PlayerInfo {
	return pangya.PlayerInfo{
		Username: c.player.Username,
		Nickname: c.player.Nickname.String,
		PlayerID: uint32(c.player.PlayerID),
		ConnID:   c.connID,
		// TODO
	}
}

func (c *Conn) getPlayerStats() pangya.PlayerStats {
	return pangya.PlayerStats{
		Pang: uint64(c.player.Pang),
		// TODO
	}
}

func (c *Conn) getPlayerEquippedConsumables() [10]uint32 {
	return [10]uint32{
		uint32(c.player.Slot0TypeID),
		uint32(c.player.Slot1TypeID),
		uint32(c.player.Slot2TypeID),
		uint32(c.player.Slot3TypeID),
		uint32(c.player.Slot4TypeID),
		uint32(c.player.Slot5TypeID),
		uint32(c.player.Slot6TypeID),
		uint32(c.player.Slot7TypeID),
		uint32(c.player.Slot8TypeID),
		uint32(c.player.Slot9TypeID),
	}
}

func (c *Conn) getPlayerEquippedItems() pangya.PlayerEquipment {
	comet := c.player.BallTypeID
	if comet == 0 {
		comet = 0x14000000
	}
	return pangya.PlayerEquipment{
		CaddieID:    uint32(c.player.CaddieID.Int64),
		CharacterID: uint32(c.player.CharacterID.Int64),
		ClubSetID:   uint32(c.player.ClubID.Int64),
		CometTypeID: uint32(comet),
		Items: pangya.PlayerEquippedItems{
			ItemIDs: c.getPlayerEquippedConsumables(),
		},
		BackgroundID:     uint32(c.player.BackgroundID.Int64),
		FrameID:          uint32(c.player.FrameID.Int64),
		StickerID:        uint32(c.player.StickerID.Int64),
		SlotID:           uint32(c.player.SlotID.Int64),
		CutInID:          uint32(c.player.CutInID.Int64),
		TitleID:          uint32(c.player.TitleID.Int64),
		BackgroundTypeID: uint32(c.player.BackgroundTypeID.Int64),
		FrameTypeID:      uint32(c.player.FrameTypeID.Int64),
		StickerTypeID:    uint32(c.player.StickerTypeID.Int64),
		SlotTypeID:       uint32(c.player.SlotTypeID.Int64),
		CutInTypeID:      uint32(c.player.CutInTypeID.Int64),
		TitleTypeID:      uint32(c.player.TitleTypeID.Int64),
		MascotID:         uint32(c.player.MascotTypeID),
		PosterID: [2]uint32{
			uint32(c.player.Poster0TypeID.Int64),
			uint32(c.player.Poster1TypeID.Int64),
		},
	}
}

func (c *Conn) getPlayerEquippedCharacter() pangya.PlayerCharacterData {
	return *c.currentCharacter
}

func (c *Conn) getPlayerEquippedClubSet() pangya.PlayerClubData {
	return pangya.PlayerClubData{
		Item: pangya.PlayerItem{
			ID:     uint32(c.player.ClubID.Int64),
			TypeID: uint32(c.player.ClubTypeID.Int64),
		},
		// TODO: stats/enchantments are not implemented.
		Stats: pangya.ClubStats{
			UpgradeStats: [5]uint16{8, 9, 8, 3, 3},
		},
	}
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
