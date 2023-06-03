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

package pangya

type UserInfo struct {
	Username      string `struct:"[22]byte"`
	Nickname      string `struct:"[22]byte"`
	Unknown       [33]byte
	GMFlag        byte
	Unknown2      [7]byte
	ConnnectionID uint32
	Unknown3      [32]byte
	ChatFlag      byte
	Unknown4      [139]byte
	PlayerID      uint32
}

type PlayerStats struct {
	Unknown           uint32
	TotalStrokes      uint32
	TotalPlayTime     uint32
	AverageStrokeTime uint32
	Unknown2          [12]byte
	OBRate            uint32
	TotalDistance     uint32
	TotalHoles        uint32
	Unknown3          uint32
	HIO               uint32
	Unknown4          [26]byte
	Experience        uint32
	Rank              Rank
	Pangs             uint64
	Unknown5          [58]byte
	QuitRateY         uint32
	Unknown6          [32]byte
	GameComboX        uint32
	GameComboY        uint32
	QuitRateX         uint32
	TotalPangsWin     uint64
	Unknown7          [38]byte
}

type PlayerEquippedItems struct {
	ItemIDs [10]uint32
}

type Decorations struct {
	Background uint32
	Frame      uint32
	Sticker    uint32
	Slot       uint32
	Unknown    uint32
	Title      uint32
}

type PlayerEquipment struct {
	CaddieID    uint32
	CharacterID uint32
	ClubSetID   uint32
	AztecIffID  uint32

	Items PlayerEquippedItems

	Unknown13 uint32
	Unknown14 uint32
	Unknown15 uint32
	Unknown16 uint32
	Unknown17 uint32
	Unknown18 uint32

	Decorations Decorations

	MascotID uint32

	Unknown26 uint32
	Unknown27 uint32
}

type PlayerCharacterData struct {
	CharTypeID  uint32
	ID          uint32
	HairColor   uint8
	Shirt       uint8
	Unknown1    byte
	Unknown2    byte
	PartTypeIDs [24]uint32
	PartIDs     [24]uint32
	Unknown3    [216]byte
	AuxParts    [5]uint32
	CutInID     uint32
	Unknown4    [12]byte
	Stats       [5]byte
	Mastery     int
	CardChar    [4]uint32
	CardCaddie  [4]uint32
	CardNPC     [4]uint32
}

type PlayerItem struct {
	ID    uint32
	IFFID uint32
}

type PlayerCaddieData struct {
	Item    PlayerItem
	Unknown [17]byte
}

type ClubStats struct {
	UpgradeStats [5]uint16
}

type PlayerClubData struct {
	Item    PlayerItem
	Unknown [10]byte
	Stats   ClubStats
}

type PlayerMascotData struct {
	Item     PlayerItem
	Unknown  [5]byte
	Text     string `struct:"[16]byte"`
	Unknown2 [33]byte
}

type PlayerData struct {
	UserInfo          UserInfo
	PlayerStats       PlayerStats
	Unknown           [78]byte
	Items             PlayerEquipment
	JunkData          [10836]byte
	EquippedCharacter PlayerCharacterData
	EquippedCaddie    PlayerCaddieData
	EquippedClub      PlayerClubData
	EquippedMascot    PlayerMascotData
}
