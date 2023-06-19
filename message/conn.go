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

import (
	"fmt"

	"github.com/pangbox/server/common"
)

// Conn holds the state for a connection to the server.
type Conn struct {
	*common.ServerConn[ClientMessage, ServerMessage]
}

// Handle runs the main connection loop.
func (c *Conn) Handle() error {
	log := c.Log()

	err := c.SendHello(&ConnectMessage{
		Unknown1: 0x0900,
		Unknown2: 0x0000,
		Unknown3: 0x002E,
		Unknown4: 0x0101,
		Unknown5: 0x0000,
	})
	if err != nil {
		return fmt.Errorf("sending hello: %w", err)
	}

	for {
		msg, err := c.ReadMessage()
		if err != nil {
			log.WithError(err).Error("Error receiving packet")
			return err
		}

		// TODO: messageng needs impl; should probably use old message server for now?
		log.Printf("%#v\n", msg)
	}
}
