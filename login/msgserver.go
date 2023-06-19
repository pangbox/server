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

package login

import "github.com/pangbox/server/common"

// These are the known possible server message IDs.
var ServerMessageTable = common.NewMessageTable(map[uint16]ServerMessage{
	0x0001: &ServerLogin{},
	0x0002: &ServerGameServerList{},
	0x0003: &ServerGameSessionKey{},
	0x0006: &ServerMacros{},
	0x0009: &ServerMessageServerList{},
	0x000E: &ServerNicknameCheckResponse{},
	0x0010: &ServerLoginSessionKey{},
	0x0011: &Server0011{},
	0x0040: &ServerGameGuardCheck{},
	0x004D: &ServerLobbiesList{},
})

// ConnectMessage is the message sent by the server when connecting.
type ConnectMessage struct {
	Unknown1 uint16
	Unknown2 uint16
	Unknown3 uint16
	Key      uint16
	Unknown4 uint16
	ServerID uint16
	Unknown6 uint16
}

func (c *ConnectMessage) SetKey(key uint8) {
	c.Key = uint16(key)
}

// ServerEntry represents a server in a ServerListMessage.
type ServerEntry struct {
	ServerName string `struct:"[40]byte"`
	ServerID   uint32
	MaxUsers   uint32
	NumUsers   uint32
	IPAddress  string `struct:"[18]byte"`
	Port       uint16
	Unknown1   uint16
	Flags      uint16
	Unknown2   [16]byte
}

type ServerList struct {
	Count   uint8 `struct:"sizeof=Servers"`
	Servers []ServerEntry
}

const (
	LoginStatusSuccess      = 0
	LoginStatusError        = 1
	LoginStatusSetNickname  = 216
	LoginStatusSetCharacter = 217
)

type LoginSuccess struct {
	Username common.PString
	UserID   uint32
	Unknown  [14]byte
	Nickname common.PString
}

type LoginSetNickname struct {
	Unknown uint32
}

type LoginSetCharacter struct {
}

const (
	LoginErrorInvalidCredentials    = 0
	LoginErrorAlreadyLoggedIn       = 5100019
	LoginErrorDuplicateConn         = 5100107
	LoginErrorInvalidReconnectToken = 5157002
)

type LoginError struct {
	Error uint32
}

type ServerLogin struct {
	ServerMessage_
	Status byte

	Success      *LoginSuccess      `struct-if:"Status == 0"`
	Error        *LoginError        `struct-if:"Status == 1"`
	SetNickname  *LoginSetNickname  `struct-if:"Status == 216"`
	SetCharacter *LoginSetCharacter `struct-if:"Status == 217"`
}

type ServerGameServerList struct {
	ServerMessage_
	ServerList
}
type ServerGameSessionKey struct {
	ServerMessage_
	Unknown    uint32
	SessionKey common.PString
}
type ServerMacros struct {
	ServerMessage_
}
type ServerMessageServerList struct {
	ServerMessage_
	ServerList
}
type ServerNicknameCheckResponse struct {
	ServerMessage_
	Unknown  uint32
	Nickname common.PString
}
type ServerLoginSessionKey struct {
	ServerMessage_
	SessionKey common.PString
}
type Server0011 struct {
	ServerMessage_
	Unknown byte
}
type ServerGameGuardCheck struct {
	ServerMessage_
}
type ServerLobbiesList struct {
	ServerMessage_
}
