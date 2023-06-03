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
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// PacketBuilder is used for ad-hoc building of packets, mainly intended for
// use in the shim client.
type PacketBuilder struct {
	buf []byte
	err error
}

func NewPacketBuilder() PacketBuilder {
	return PacketBuilder{make([]byte, 0, 256), nil}
}

func (builder PacketBuilder) PutPString(s string) PacketBuilder {
	c := len(s)
	if c > 0xFFFF {
		builder.err = errors.Join(builder.err, fmt.Errorf("string too big (%d)", c))
		return builder
	}
	builder.buf = binary.LittleEndian.AppendUint16(builder.buf, uint16(c))
	builder.buf = append(builder.buf, s...)
	return builder
}

func (builder PacketBuilder) PutString(s string, l int) PacketBuilder {
	if l > len(s) {
		s += strings.Repeat("\000", l-len(s))
	} else if l < len(s) {
		s = s[:l]
	}
	builder.buf = append(builder.buf, s...)
	return builder
}

func (builder PacketBuilder) PutBytes(raw []byte) PacketBuilder {
	builder.buf = append(builder.buf, raw...)
	return builder
}

func (builder PacketBuilder) PutUint8(v uint8) PacketBuilder {
	builder.buf = append(builder.buf, v)
	return builder
}

func (builder PacketBuilder) PutUint16(v uint16) PacketBuilder {
	builder.buf = binary.LittleEndian.AppendUint16(builder.buf, v)
	return builder
}

func (builder PacketBuilder) PutUint32(v uint32) PacketBuilder {
	builder.buf = binary.LittleEndian.AppendUint32(builder.buf, v)
	return builder
}

func (builder PacketBuilder) PutInt8(v int8) PacketBuilder {
	return builder.PutUint8(uint8(v))
}

func (builder PacketBuilder) PutInt16(v int16) PacketBuilder {
	return builder.PutUint16(uint16(v))
}

func (builder PacketBuilder) PutInt32(v int32) PacketBuilder {
	return builder.PutUint32(uint32(v))
}

func (builder PacketBuilder) Build() ([]byte, error) {
	if builder.err != nil {
		return nil, builder.err
	}
	return builder.buf, nil
}

func (builder PacketBuilder) MustBuild() []byte {
	b, err := builder.Build()
	if err != nil {
		panic(err)
	}
	return b
}
