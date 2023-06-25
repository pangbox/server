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

type PlayerInfo struct {
	Username         string `struct:"[22]byte"`
	Nickname         string `struct:"[22]byte"`
	GuildName        string `struct:"[17]byte"`
	GuildEmblemImage string `struct:"[24]byte"`
	ConnID           uint32
	Unknown          [12]byte
	Unknown2         uint32
	Unknown3         uint32
	Unknown4         uint16
	Unknown5         [6]byte
	Unknown6         [16]byte
	GlobalID         string `struct:"[128]byte"`
	PlayerID         uint32
}

type PlayerStats struct {
	TotalStrokes      uint32
	TotalPutts        uint32
	Time              uint32
	TimeHitting       uint32
	LongestDrive      float32
	PangyaHits        uint32
	Timeouts          uint32
	OBs               uint32
	TotalDistance     uint32
	TotalHoles        uint32
	HoleUnfinished    uint32 //Holes that you don't end up putting/chipping in due to mode; e.g. match play
	TotalHIO          uint32
	BunkersHit        uint16
	FairwaysHit       uint32
	TotalAlbatross    uint32
	Warnings          uint32
	PuttIns           uint32
	LongestPutt       float32
	LongestChip       float32
	TotalXP           uint32
	Level             byte
	Pang              uint64
	TotalScore        int32
	Difficulty1Score  uint8 //Not 100% sure on these
	Difficulty2Score  uint8
	Difficulty3Score  uint8
	Difficulty4Score  uint8
	Difficulty5Score  uint8
	UnknownFlag       uint8 //possibly total?
	BestPang1         uint64
	BestPang2         uint64
	BestPang3         uint64
	BestPang4         uint64
	BestPang5         uint64
	BestPangTotal     uint64
	GamesPlayed       uint32
	TeamHole          uint32
	TeamWin           uint32
	TeamGame          uint32
	LadderMMR         uint32
	LadderHoles       uint32
	LadderWins        uint32
	LadderLosses      uint32
	LadderDraws       uint32
	ComboNum          uint32
	ComboDenom        uint32
	Quits             uint32
	PangBattleTotal   int32
	PangBattleWins    uint32
	PangBattleLosses  uint32
	PangBattleAllIn   uint32
	PangBattleCombo   uint32
	PangBattleUnknown uint32   //could be first medal - there are 6
	Unknown24         [10]byte //other 5 medals?  However, this also could be school related stuff according to jp
	GameCountSeason   uint32
	Unknown26         [8]byte //
}

type PlayerEquippedItems struct {
	ItemIDs [10]uint32
}

type PlayerEquipment struct {
	CaddieID    uint32
	CharacterID uint32
	ClubSetID   uint32
	CometTypeID uint32

	Items PlayerEquippedItems

	BackgroundID uint32
	FrameID      uint32
	StickerID    uint32
	SlotID       uint32
	CutInID      uint32
	TitleID      uint32

	BackgroundTypeID uint32
	FrameTypeID      uint32
	StickerTypeID    uint32
	SlotTypeID       uint32
	CutInTypeID      uint32
	TitleTypeID      uint32

	MascotID uint32
	PosterID [2]uint32
}

type PlayerCourseData struct {
	CourseID     uint8
	TotalStrokes uint32
	TotalPutts   uint32
	NumHoles     uint32
	Unknown      uint32
	Unknown2     uint32
	Unknown3     uint32
	TotalScore   uint32
	BestScore    int8
	BestPang     uint32
	Unknown4     uint32
	CharTypeID   uint32
	Unknown5     int8
}

type PlayerSeasonData struct {
	Courses [21]PlayerCourseData
}

type PlayerSeasonHistory struct {
	Seasons [12]PlayerSeasonData
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
	Unknown4    [16]byte
	Stats       [5]byte
	CardChar    [4]uint32
	CardCaddie  [4]uint32
	CardNPC     [4]uint32
}

type PlayerItem struct {
	ID     uint32
	TypeID uint32
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
	UserInfo          PlayerInfo
	PlayerStats       PlayerStats
	Trophy            [13][3]uint16
	EquippedItems     PlayerEquipment
	SeasonHistory     PlayerSeasonHistory
	EquippedCharacter PlayerCharacterData
	EquippedCaddie    PlayerCaddieData
	EquippedClub      PlayerClubData
	EquippedMascot    PlayerMascotData
}
