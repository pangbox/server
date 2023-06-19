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

package message

import "github.com/pangbox/server/common"

var ServerMessageTable = common.NewMessageTable(map[uint16]ServerMessage{
	0x0001: &Server0001{},
})

// ConnectMessage is the message sent by the server when connecting.
type ConnectMessage struct {
	Unknown1 uint16
	Unknown2 uint16
	Unknown3 uint16
	Unknown4 uint16
	Key      uint16
	Unknown5 uint16
}

func (c *ConnectMessage) SetKey(key uint8) {
	c.Key = uint16(key)
}

// Server0001 is an unknown message.
type Server0001 struct {
	ServerMessage_
}
