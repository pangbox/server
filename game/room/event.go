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

package room

import (
	gamemodel "github.com/pangbox/server/game/model"
	gamepacket "github.com/pangbox/server/game/packet"
	"github.com/pangbox/server/pangya"
)

type lobbyEvent struct{}
type LobbyEvent interface{ isLobbyEvent() }
type roomEvent struct{}
type RoomEvent interface{ isRoomEvent() }

func (lobbyEvent) isLobbyEvent() {}
func (roomEvent) isRoomEvent()   {}

type LobbyPlayerJoin struct {
	lobbyEvent
	Entry gamemodel.LobbyPlayer
	Conn  *gamepacket.ServerConn
}

type LobbyPlayerUpdate struct {
	lobbyEvent
	Entry gamemodel.LobbyPlayer
}

type LobbyPlayerUpdateRoom struct {
	lobbyEvent
	ConnID     uint32
	RoomNumber int16
}

type LobbyPlayerLeave struct {
	lobbyEvent
	ConnID uint32
}

type LobbyRoomCreate struct {
	lobbyEvent
	Room gamemodel.RoomState
}

type LobbyRoomUpdate struct {
	lobbyEvent
	Room gamemodel.RoomState
}

type LobbyRoomRemove struct {
	lobbyEvent
	Room gamemodel.RoomState
}

type RoomGetInfo struct {
	roomEvent
}

type RoomPlayerJoin struct {
	roomEvent
	Entry      *gamemodel.RoomPlayerEntry
	PlayerData pangya.PlayerData
	Conn       *gamepacket.ServerConn
}

type RoomPlayerLeave struct {
	roomEvent
	ConnID uint32
}

type RoomAction struct {
	roomEvent
	ConnID uint32
	Action gamemodel.RoomAction
}

type RoomPlayerIdle struct {
	roomEvent
	ConnID uint32
	Idle   bool
}

type RoomPlayerReady struct {
	roomEvent
	ConnID uint32
	Ready  bool
}

type RoomPlayerKick struct {
	roomEvent
	ConnID     uint32
	KickConnID uint32
}

type RoomLoadingProgress struct {
	roomEvent
	ConnID   uint32
	Progress uint8
}

type RoomSettingsChange struct {
	roomEvent
	ConnID  uint32
	Changes []gamemodel.RoomSettingsChange
}

type RoomStartGame struct {
	roomEvent
	ConnID uint32
}

type RoomGameReady struct {
	roomEvent
	ConnID uint32
}

type RoomGameShotCommit struct {
	roomEvent
	ConnID           uint32
	ShotStrength     float32
	ShotAccuracy     float32
	ShotEnglishCurve float32
	ShotEnglishSpin  float32
	Unknown2         [30]byte
	Unknown3         [4]float32
}

type RoomGameShotRotate struct {
	roomEvent
	ConnID uint32
	Angle  float32
}

type RoomGameShotPower struct {
	roomEvent
	ConnID uint32
	Level  uint8
}

type RoomGameShotClubChange struct {
	roomEvent
	ConnID uint32
	Club   uint8
}

type RoomGameShotItemUse struct {
	roomEvent
	ConnID     uint32
	ItemTypeID uint32
}

type RoomGameTypingIndicator struct {
	roomEvent
	ConnID uint32
	Status int16
}

type RoomGameShotCometRelief struct {
	roomEvent
	ConnID  uint32
	X, Y, Z float32
}

type RoomGameTurn struct {
	roomEvent
	ConnID uint32
}

type RoomGameTurnEnd struct {
	roomEvent
	ConnID uint32
}

type RoomGameHoleEnd struct {
	roomEvent
	ConnID uint32
}

type RoomGameShotSync struct {
	roomEvent
	ConnID uint32
	Data   gamemodel.ShotSyncData
}

type ChatMessage struct {
	lobbyEvent
	roomEvent
	Nickname string
	Message  string
}
