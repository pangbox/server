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

package gamemodel

import (
	"github.com/pangbox/server/common"
	"github.com/pangbox/server/pangya"
)

type RoomStatusFlag uint16

const (
	RoomStateUnknown1 RoomStatusFlag = 1 << iota
	RoomStateUnknown2
	RoomStateAway
	RoomStateMaster
	RoomStateUnknown3
	RoomStateUnknown4
	RoomStateUnknown5
	RoomStateUnknown6
	RoomStateUnknown7
	RoomStateReady
)

type RoomPlayerEntry struct {
	ConnID           uint32
	Nickname         string `struct:"[22]byte"`
	GuildName        string `struct:"[17]byte"`
	Slot             uint8
	PlayerFlags      uint32 // (I think GM flag goes here!)
	TitleID          uint32
	CharTypeID       uint32
	BackgroundTypeID uint32
	FrameTypeID      uint32
	StickerTypeID    uint32
	SlotTypeID       uint32
	CutInTypeID      uint32
	TitleTypeID      uint32
	StatusFlags      RoomStatusFlag
	Rank             uint8
	Unknown          uint16
	GuildID          uint32
	GuildEmblemImage string `struct:"[12]byte"`
	PlayerID         uint32
	Unknown2         uint32
	Unknown3         uint8 // <-- suspect, this may not be in the right place
	LoungeState      uint32
	Unknown4         uint16
	X                float32
	Y                float32
	Z                float32
	Angle            float32
	Unknown5         uint32
	ShopUnknown      uint32
	ShopName         string `struct:"[64]byte"`
	MascotTypeID     uint32
	GlobalID         string `struct:"[22]byte"`
	Unknown6         [106]byte
	Guest            bool `struct:"byte"`
	AverageScore     float32
	CharacterData    pangya.PlayerCharacterData
}

type RoomActionRotation struct {
	Z float32
}

type RoomActionPosition struct {
	X, Y, Z float32
}

type RoomAction struct {
	ActionType  byte
	Rotation    *RoomActionRotation `struct-if:"ActionType == 0"`
	PositionAbs *RoomActionRotation `struct-if:"ActionType == 4"`
	PositionRel *RoomActionRotation `struct-if:"ActionType == 6"`
	Emote       *common.PString     `struct-if:"ActionType == 7"`
	Departure   *uint32             `struct-if:"ActionType == 8"`
}

type GamePhase int

const (
	LobbyPhase GamePhase = 1
	WaitingLoad
	InGame
)

type RoomState struct {
	Active          bool
	Open            bool
	ShotTimerMS     uint32
	GameTimerMS     uint32
	NumUsers        uint8
	MaxUsers        uint8
	RoomType        byte
	NumHoles        byte
	CurrentHole     byte
	HoleProgression byte
	Course          byte
	RoomName        string
	RoomNumber      int16
	Password        string
	OwnerConnID     uint32
	NaturalWind     uint32

	GamePhase    GamePhase
	ShotSync     *ShotSyncData
	HoleInfo     *HoleInfo
	ActiveConnID uint32
}

type RoomSettingsChange struct {
	Type             byte
	RoomName         *common.PString `struct-if:"Type == 0"`
	Password         *common.PString `struct-if:"Type == 1"`
	RoomType         *byte           `struct-if:"Type == 2"`
	Course           *byte           `struct-if:"Type == 3"`
	NumHoles         *uint8          `struct-if:"Type == 4"`
	HoleProgression  *uint8          `struct-if:"Type == 5"`
	ShotTimerSeconds *uint8          `struct-if:"Type == 6"`
	MaxUsers         *uint8          `struct-if:"Type == 7"`
	GameTimerMinutes *uint8          `struct-if:"Type == 8"`
	ArtifactID       *uint32         `struct-if:"Type == 13"`
	NaturalWind      *uint32         `struct-if:"Type == 14"`
}

// Used when viewing room info in the lobby.
type RoomInfo struct {
	PlayerCount uint32
	NumHoles    uint8
	Unknown     uint32
	Course      uint8
	RoomType    uint8
	Mode        uint8
	Trophy      uint32
	Users       []RoomInfoPlayer `struct:"sizefrom=UserCount"`
}

type RoomInfoPlayer struct {
	ConnID      uint32
	Rank        uint8
	Unknown     uint8
	PlayerFlags uint32
	TitleID     uint32
	Unknown2    uint32
}

type ShotSyncData struct {
	ActiveConnID uint32
	X, Y, Z      float32
	Unknown1     [3]byte
	Pang         uint32
	BonusPang    uint32
	Unknown2     [11]byte
}

type HoleInfo struct {
	Par  uint8
	TeeX float32
	TeeZ float32
	PinX float32
	PinZ float32
}
