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
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-restruct/restruct"
	"github.com/pangbox/server/common"
	gamepacket "github.com/pangbox/server/game/packet"
	"github.com/pangbox/server/login"
	"github.com/pangbox/server/message"
	"github.com/rs/zerolog"
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
			return gamepacket.ServerMessageTable.Any(), nil
		case "client":
			return gamepacket.ClientMessageTable.Any(), nil
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

func ParseHex(input []byte) ([]byte, error) {
	output := []byte{}

	for i, n := range strings.Split(string(input), ",") {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		v, err := strconv.ParseUint(n, 0, 8)
		if err != nil {
			return nil, fmt.Errorf("unexpected sequence %d in hex string: %q", i, n)
		}
		output = append(output, byte(v))
	}

	return output, nil
}

func main() {
	var err error
	hex := flag.Bool("hex", false, "If set, parse hex input.")
	flag.Parse()

	log := zerolog.
		New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger()

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
			log.Fatal().Err(err).Msg("error opening input file")
		}
	}

	messageTable, err := GetMessageTable(args[0], args[1])
	if err != nil {
		log.Fatal().Err(err).Msg("error getting message table")
	}

	var packet []byte

	packet, err = io.ReadAll(input)
	if err != nil {
		log.Fatal().Err(err).Msg("error reading input")
	}

	if *hex {
		packet, err = ParseHex(packet)
		if err != nil {
			log.Fatal().Err(err).Msg("error parsing hex")
		}
	}

	msgid := binary.LittleEndian.Uint16(packet[:2])

	message, err := messageTable.Build(msgid)
	if err != nil {
		log.Fatal().Err(err).Msg("error building packet")
	}

	err = restruct.Unpack(packet[2:], binary.LittleEndian, message)
	if err != nil {
		fmt.Fprintln(os.Stderr, "-----BEGIN PARTIAL DATA-----")
		spew.Fdump(os.Stderr, message)
		fmt.Fprintln(os.Stderr, "-----END PARTIAL DATA-----")
		log.Fatal().Err(err).Msg("error parsing packet")
	}

	spew.Dump(message)

	sz, err := restruct.SizeOf(message)
	if err != nil {
		log.Fatal().Err(err).Msg("error getting packet size")
	}

	if sz != len(packet[2:]) {
		extra := len(packet[2:]) - sz
		log.Warn().Msgf("warning: %[1]d (%[1]08x) extra bytes... (%d total, %d parsed)", extra, len(packet[2:]), sz)
	}
}
