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

// ClientMessageID is the type used to identify client messages.
type ClientMessageID uint16

var ClientMessageTable = common.NewMessageTable(map[uint16]ClientMessage{
	0x0001: &ClientLogin{},
	0x0003: &ClientSelectServer{},
	0x0006: &ClientSetNickname{},
	0x0007: &ClientCheckNickname{},
	0x0008: &ClientSelectCharacter{},
	0x000B: &ClientReconnect{},
})

type ClientLoginMessage struct {
}

// ClientLogin is the payload associated with the ClientLogin packet.
// It is sent by the client when the client logs in.
type ClientLogin struct {
	ClientMessage_
	Username common.PString
	Password common.PString
}

type ClientSelectServer struct {
	ClientMessage_
	Unknown1 uint32
}

type ClientSetNickname struct {
	ClientMessage_
	Nickname common.PString
}

type ClientCheckNickname struct {
	ClientMessage_
	Nickname common.PString
}

type ClientSelectCharacter struct {
	ClientMessage_
	CharacterID uint32
	HairColor   uint8
	Unknown     uint8
}

type ClientReconnect struct {
	ClientMessage_
}
