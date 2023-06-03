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
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-restruct/restruct"
	"github.com/pangbox/server/common"
	"github.com/pangbox/server/game"
	"github.com/pangbox/server/login"
	"github.com/pangbox/server/message"
)

func GetMessageTable(server string, origin string) (common.AnyMessageTable, error) {
	switch server {
	case "login":
		switch origin {
		case "server":
			return login.ServerMessageTable.Any(), nil
		case "client":
			return login.ClientMessageTable.Any(), nil
		default:
			return nil, fmt.Errorf("unexpected origin %q; valid values are server, client", origin)
		}
	case "game":
		switch origin {
		case "server":
			return game.ServerMessageTable.Any(), nil
		case "client":
			return game.ClientMessageTable.Any(), nil
		default:
			return nil, fmt.Errorf("unexpected origin %q; valid values are server, client", origin)
		}
	case "message":
		switch origin {
		case "server":
			return message.ServerMessageTable.Any(), nil
		case "client":
			return message.ClientMessageTable.Any(), nil
		default:
			return nil, fmt.Errorf("unexpected origin %q; valid values are server, client", origin)
		}
	default:
		return nil, fmt.Errorf("unexpected server %q; valid values are login, game, message", server)
	}
}

func ParseHex(input []byte) []byte {
	output := []byte{}

	for i, n := range strings.Split(string(input), ",") {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		v, err := strconv.ParseUint(n, 0, 8)
		if err != nil {
			log.Fatalf("Unexpected sequence %d in hex string: %q", i, n)
		}
		output = append(output, byte(v))
	}

	return output
}

func main() {
	var err error
	hex := flag.Bool("hex", false, "If set, parse hex input.")
	flag.Parse()

	args := flag.Args()

	if len(args) < 2 || len(args) > 3 {
		fmt.Fprintf(os.Stderr, "Usage: %v login|game|message server|client [FILE...] [-hex]", os.Args[0])
		flag.PrintDefaults()
		return
	}

	input := os.Stdin
	if len(args) == 3 && args[2] != "-" {
		input, err = os.Open(args[2])
		if err != nil {
			log.Fatalf("Error opening input file: %v", err)
		}
	}

	messageTable, err := GetMessageTable(args[0], args[1])
	if err != nil {
		log.Fatalf("Error getting message table: %v", err)
	}

	var packet []byte

	packet, err = io.ReadAll(input)
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}

	if *hex {
		packet = ParseHex(packet)
	}

	msgid := binary.LittleEndian.Uint16(packet[:2])

	message, err := messageTable.Build(msgid)
	if err != nil {
		log.Fatalf("Error building packet: %v", err)
	}

	err = restruct.Unpack(packet[2:], binary.LittleEndian, message)
	if err != nil {
		log.Fatalf("Error parsing packet: %v; partial result: %#v; data: % 02x", err, message, packet)
	}

	spew.Dump(message)
}
