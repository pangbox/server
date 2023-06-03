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

package main

import (
	"encoding/binary"
	"flag"
	"io"
	"math/rand"
	"net"

	"github.com/davecgh/go-spew/spew"
	"github.com/manifoldco/promptui"
	"github.com/pangbox/pangcrypt"
	"github.com/pangbox/server/common"
	log "github.com/sirupsen/logrus"
)

func readServerPacket(k byte, r io.Reader) ([]byte, error) {
	packetHeaderBytes := [3]byte{}

	read, err := r.Read(packetHeaderBytes[:])
	if err != nil {
		return nil, err
	} else if read != len(packetHeaderBytes) {
		return nil, io.EOF
	}

	remaining := binary.LittleEndian.Uint16(packetHeaderBytes[1:3])
	packet := make([]byte, len(packetHeaderBytes)+int(remaining))
	copy(packet[:3], packetHeaderBytes[:])
	_, err = io.ReadFull(r, packet[3:])
	if err != nil {
		return nil, err
	}

	return pangcrypt.ServerDecrypt(packet, k)
}

func sendClientPacket(k byte, w io.Writer, data []byte) {
	salt := rand.Intn(0x100)
	outpkt, err := pangcrypt.ClientEncrypt(data, k, byte(salt))

	if err != nil {
		log.Fatal("While encrypting outgoing Login packet:", err)
	}

	if n, err := w.Write(outpkt); err != nil {
		log.Fatal("While sending client packet:", err)
	} else if n < len(outpkt) {
		log.Fatalf("Short write on out: %d of %d", n, len(outpkt))
	}
}

func main() {
	loginAddr := flag.String("login_addr", "127.0.0.1:10101", "address of login server")
	flag.Parse()

	sock, err := net.Dial("tcp", *loginAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Read credentials.
	prompt := promptui.Prompt{Label: "Username"}
	username, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
	prompt = promptui.Prompt{Label: "Password", Mask: '*'}
	password, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Read hello.
	hello := [14]byte{}
	if c, err := sock.Read(hello[:]); err != nil {
		log.Fatal(err)
	} else if c < len(hello) {
		log.Fatal("short read on hello packet")
	}
	key := hello[6]

	log.Printf("Connected to %s with key %d.", sock.RemoteAddr(), key)

	sendClientPacket(key, sock, common.NewPacketBuilder().
		PutUint16(0x0001).
		PutPString(username).
		PutPString(password).
		PutBytes([]byte{
			0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF,
			0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF,
		}).MustBuild(),
	)

	for {
		packet, err := readServerPacket(key, sock)
		if err != nil {
			log.Fatal(err)
		}
		spew.Dump(packet)
	}
}
