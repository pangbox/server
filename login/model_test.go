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

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/go-restruct/restruct"
	"github.com/stretchr/testify/assert"
)

func TestMessageStructure(t *testing.T) {
	tests := []struct {
		data  []byte
		value interface{}
	}{
		{
			data: []byte{
				/* 0x00 */ 0x01, 0x50, 0x61, 0x6e, 0x67, 0x62, 0x6f, 0x78,
				/* 0x08 */ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				/* 0x10 */ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				/* 0x18 */ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				/* 0x20 */ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				/* 0x28 */ 0x00, 0xea, 0x4e, 0x00, 0x00, 0xd0, 0x07, 0x00,
				/* 0x30 */ 0x00, 0x01, 0x00, 0x00, 0x00, 0x31, 0x32, 0x37,
				/* 0x38 */ 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31, 0x00, 0x00,
				/* 0x40 */ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xea,
				/* 0x48 */ 0x4e, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00,
				/* 0x50 */ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				/* 0x58 */ 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			value: &ServerList{
				Count: 1,
				Servers: []ServerEntry{
					{
						ServerName: "Pangbox",
						ServerID:   20202,
						NumUsers:   1,
						MaxUsers:   2000,
						IPAddress:  "127.0.0.1",
						Port:       20202,
						Flags:      0x800,
					},
				},
			},
		},
	}

	for _, test := range tests {
		v := reflect.New(reflect.TypeOf(test.value).Elem())
		err := restruct.Unpack(test.data, binary.LittleEndian, v.Interface())
		assert.Nil(t, err, "unpack")
		assert.Equal(t, test.value, v.Interface(), "unpack")

		data, err := restruct.Pack(binary.LittleEndian, test.value)
		assert.Nil(t, err, "pack")
		assert.Equal(t, test.data, data, "pack")

		size, err := restruct.SizeOf(test.value)
		assert.Nil(t, err, "sizing")
		assert.Equal(t, len(test.data), size, "sizing")
	}
}
