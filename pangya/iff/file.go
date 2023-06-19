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

package iff

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/go-restruct/restruct"
	"github.com/pangbox/pangfiles/pak"
	log "github.com/sirupsen/logrus"
)

type Archive struct {
	ItemMap map[uint32]*Item
}

// Filenames to look for to find client IFF.
var iffSearchOrder = []string{
	"pangya_gb.iff",
	"pangya_us.iff", // older US, before global
	"pangya_jp.iff",
	"pangya_eu.iff",
	"pangya_th.iff",
	"pangya_sg.iff", // nb: uses jp key
	"pangya_idnes.iff",
	"pangya.iff", // kr (present in some gb ver too)
}

func LoadFromPak(fs pak.FS) (*Archive, error) {
	data, err := findPangYaIFF(fs)
	if err != nil {
		return nil, err
	}
	return Load(data)
}

func Load(data []byte) (*Archive, error) {
	archive := &Archive{}
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for _, f := range r.File {
		log.Debugf("Found IFF: %s", f.Name)
		if f.Name == "Item.iff" {
			r, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer r.Close()
			data, err := io.ReadAll(r)
			if err != nil {
				return nil, err
			}
			archive.loadItems(data)
		}
	}
	return archive, nil
}

func (a *Archive) loadItems(data []byte) error {
	file, err := LoadItems(data)
	if err != nil {
		return err
	}
	a.ItemMap = make(map[uint32]*Item)
	for _, item := range file.Records {
		a.ItemMap[item.ID] = &item
	}
	return nil
}

func LoadItems(data []byte) (*File[Item], error) {
	recordCount := binary.LittleEndian.Uint16(data[:2])
	recordLength := (len(data) - 0x8) / int(recordCount)
	version := Version(binary.LittleEndian.Uint32(data[4:8]))

	switch version {
	case Version11:
		switch recordLength {
		case 0x78:
			return loadItemVersion[ItemV11_78](data)
		case 0x98:
			return loadItemVersion[ItemV11_98](data)
		case 0xB0:
			return loadItemVersion[ItemV11_B0](data)
		case 0xC0:
			return loadItemVersion[ItemV11_C0](data)
		case 0xC4:
			return loadItemVersion[ItemV11_C4](data)
		case 0xD8:
			// JP4xx has the common times after the model name
			// this is back to normal in JP5xx
			var testItem ItemV11_D8_2
			restruct.Unpack(data[8:], binary.LittleEndian, &testItem)
			if testItem.StartTime.IsZero() || testItem.StartTime.IsValid() {
				return loadItemVersion[ItemV11_D8_2](data)
			} else {
				return loadItemVersion[ItemV11_D8_1](data)
			}
		default:
			return nil, fmt.Errorf("unknown item iff v%d record size %d (please report)", version, recordLength)
		}
	case Version13:
		switch recordLength {
		case 0xE0:
			return loadItemVersion[ItemV13_E0](data)
		case 0xF8:
			return loadItemVersion[ItemV13_F8](data)
		default:
			return nil, fmt.Errorf("unknown item iff v%d record size %d (please report)", version, recordLength)
		}
	default:
		return nil, errors.New("unknown item iff version")
	}
}

func loadItemVersion[T itemGeneric](data []byte) (*File[Item], error) {
	result := &File[Item]{}
	f, err := LoadFile[T](data)
	if err != nil {
		return nil, err
	}
	result.Header = f.Header
	for _, record := range f.Records {
		result.Records = append(result.Records, record.Generic())
	}
	size, err := restruct.SizeOf(f)
	if err != nil {
		return nil, err
	}
	if len(data) > size {
		return nil, fmt.Errorf("short read: read %d of %d bytes", size, len(data))
	}
	return result, nil
}

func LoadFile[T any](data []byte) (*File[T], error) {
	file := &File[T]{}
	if err := restruct.Unpack(data, binary.LittleEndian, file); err != nil {
		return nil, err
	}
	return file, nil
}

func findPangYaIFF(fs pak.FS) ([]byte, error) {
	var errs error
	if fs.NumFiles() == 0 {
		return nil, fmt.Errorf("no pak files found, aborting IFF search")
	}
	for _, fn := range iffSearchOrder {
		data, err := fs.ReadFile(fn)
		if err == nil {
			return data, err
		}
		errs = errors.Join(errs, fmt.Errorf("trying: %q: %w", fn, err))
	}
	return nil, fmt.Errorf("error finding IFF file: %w", errs)
}
