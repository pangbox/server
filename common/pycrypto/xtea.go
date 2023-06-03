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

package pycrypto

import (
	"fmt"

	"github.com/pangbox/pangfiles/crypto/pyxtea"
)

var Keys = []pyxtea.Key{
	pyxtea.KeyUS,
	pyxtea.KeyJP,
	pyxtea.KeyTH,
	pyxtea.KeyEU,
	pyxtea.KeyID,
	pyxtea.KeyKR,
}

var regionToKey = map[string]pyxtea.Key{
	"us": pyxtea.KeyUS,
	"jp": pyxtea.KeyJP,
	"th": pyxtea.KeyTH,
	"eu": pyxtea.KeyEU,
	"id": pyxtea.KeyID,
	"kr": pyxtea.KeyKR,
}

var keyToRegion = map[pyxtea.Key]string{
	pyxtea.KeyUS: "us",
	pyxtea.KeyJP: "jp",
	pyxtea.KeyTH: "th",
	pyxtea.KeyEU: "eu",
	pyxtea.KeyID: "id",
	pyxtea.KeyKR: "kr",
}

func GetRegionKey(regionCode string) (pyxtea.Key, error) {
	key, ok := regionToKey[regionCode]
	if !ok {
		return pyxtea.Key{}, fmt.Errorf("invalid region %q (valid regions: us, jp, th, eu, id, kr)", regionCode)
	}
	return key, nil
}

func GetKeyRegion(key pyxtea.Key) string {
	region, ok := keyToRegion[key]
	if !ok {
		panic("programming error: unexpected key")
	}
	return region
}
