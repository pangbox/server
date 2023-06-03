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
	"errors"
	"fmt"

	"github.com/pangbox/pangfiles/pak"
	log "github.com/sirupsen/logrus"
)

type Archive struct {
}

// Filenames to look for to find client IFF.
var iffSearchOrder = []string{
	"pangya_gb.iff",
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
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for _, f := range r.File {
		log.Debugf("Found IFF: %s", f.Name)
	}
	return &Archive{}, nil
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
