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

package common

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"io"
	"math/rand"
	"net"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-restruct/restruct"
	"github.com/pangbox/pangcrypt"
	log "github.com/sirupsen/logrus"
)

type HelloMessage interface {
	SetKey(key uint8)
}

// ServerConn provides base functionality for PangYa-compatible servers.
type ServerConn[ClientMsg Message, ServerMsg Message] struct {
	sendMu sync.RWMutex

	socket net.Conn
	key    uint8
	log    *log.Entry

	ClientMsg MessageTable[ClientMsg]
	ServerMsg MessageTable[ServerMsg]
}

func NewServerConn[C Message, S Message](
	socket net.Conn,
	log *log.Entry,
	clientMsg MessageTable[C],
	serverMsg MessageTable[S],
) *ServerConn[C, S] {
	key := uint8(rand.Intn(16))
	return &ServerConn[C, S]{
		socket:    socket,
		key:       key,
		log:       log,
		ClientMsg: clientMsg,
		ServerMsg: serverMsg,
	}
}

// RemoteAddr returns the address of the remotely connected endpoint.
func (c *ServerConn[_, _]) RemoteAddr() net.Addr {
	return c.socket.RemoteAddr()
}

// Log returns a log.Entry for logging.
func (c *ServerConn[_, _]) Log() *log.Entry {
	return c.log
}

// SendHello sends the initial handshake bytes to the client.
func (c *ServerConn[_, _]) SendHello(hello HelloMessage) error {
	hello.SetKey(c.key)

	data, err := restruct.Pack(binary.LittleEndian, hello)
	if err != nil {
		return err
	}

	c.sendMu.Lock()
	defer c.sendMu.Unlock()

	_, err = c.socket.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// ReadPacket attempts to read a single packet from the socket.
// It is not safe to call ReadPacket from multiple goroutines.
func (c *ServerConn[_, _]) ReadPacket() ([]byte, error) {
	packetHeaderBytes := [4]byte{}

	read, err := c.socket.Read(packetHeaderBytes[:])
	if err != nil {
		return nil, err
	} else if read != len(packetHeaderBytes) {
		return nil, io.EOF
	}

	remaining := binary.LittleEndian.Uint16(packetHeaderBytes[1:3])
	packet := make([]byte, len(packetHeaderBytes)+int(remaining))
	copy(packet[:4], packetHeaderBytes[:])
	read, err = c.socket.Read(packet[4:])
	if err != nil {
		return nil, err
	} else if read != len(packet[4:]) {
		return nil, io.EOF
	}

	return pangcrypt.ClientDecrypt(packet, c.key)
}

// ParsePacket attempts to construct a packet from packet data.
func (c *ServerConn[ClientMsg, _]) ParsePacket(packet []byte) (ClientMsg, error) {
	msgid := binary.LittleEndian.Uint16(packet[:2])

	c.log.Debug(hex.Dump(packet))

	message, err := c.ClientMsg.Build(msgid)
	if err != nil {
		return message, err
	}

	err = restruct.Unpack(packet[2:], binary.LittleEndian, message)
	if err != nil {
		return message, err
	}

	return message, nil
}

// ReadMessage reads a single packet and parses it.
func (c *ServerConn[ClientMsg, _]) ReadMessage() (ClientMsg, error) {
	var message ClientMsg

	data, err := c.ReadPacket()
	if err != nil {
		return message, err
	}

	return c.ParsePacket(data)
}

// SendMessage sends a message to the client. It is safe to call SendMessage
// from multiple goroutines.
func (c *ServerConn[_, ServerMsg]) SendMessage(_ context.Context, msg ServerMsg) error {
	// TODO: need to handle context cancellation

	data, err := restruct.Pack(binary.LittleEndian, msg)
	if err != nil {
		return err
	}

	id, err := c.ServerMsg.ID(msg)
	if err != nil {
		return err
	}

	msgid := [2]byte{}
	binary.LittleEndian.PutUint16(msgid[:], id)
	data = append(msgid[:], data...)

	data, err = pangcrypt.ServerEncrypt(data, c.key, 0)
	if err != nil {
		return err
	}

	c.sendMu.Lock()
	defer c.sendMu.Unlock()

	written, err := c.socket.Write(data)
	if err != nil {
		return err
	} else if written != len(data) {
		return io.EOF
	}

	return nil
}

// DebugMsg prints a message.
func (c *ServerConn[_, ServerMsg]) DebugMsg(msg ServerMsg) error {
	data, err := restruct.Pack(binary.LittleEndian, msg)
	if err != nil {
		return err
	}

	id, err := c.ServerMsg.ID(msg)
	if err != nil {
		return err
	}

	msgid := [2]byte{}
	binary.LittleEndian.PutUint16(msgid[:], id)
	data = append(msgid[:], data...)

	spew.Dump(data)
	return nil
}

// SendRaw sends raw bytes into a PangYa packet.
func (c *ServerConn[_, ServerMsg]) SendRaw(data []byte) error {
	data, err := pangcrypt.ServerEncrypt(data, c.key, 0)
	if err != nil {
		return err
	}
	written, err := c.socket.Write(data)
	if err != nil {
		return err
	} else if written != len(data) {
		return io.EOF
	}
	return nil
}
