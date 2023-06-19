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

package minibox

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/pangbox/rugburn/slipstrm/embedded"
	"github.com/pangbox/rugburn/slipstrm/patcher"
)

type RugburnOptions struct {
	PangyaDir string
}

type RugburnPatcher struct {
	mu sync.RWMutex

	// +checklocks:mu
	path string
	// +checklocks:mu
	calc int64

	// +checklocks:mu
	haveOrig bool
	// +checklocks:mu
	rugburnVer string
	// +checklocks:mu
	rugburnVerErr error
}

func NewRugburnPatcher() *RugburnPatcher {
	return new(RugburnPatcher)
}

func (p *RugburnPatcher) Configure(opts RugburnOptions) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.path = filepath.Join(opts.PangyaDir, "ijl15.dll")
}

func (p *RugburnPatcher) recalc() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	finfo, err := os.Stat(p.path)
	if err != nil {
		return err
	}

	ncalc := finfo.Size() ^ finfo.ModTime().Unix()
	if p.calc == ncalc {
		return nil
	}

	ijl15, err := os.ReadFile(p.path)
	if err != nil {
		return err
	}

	orig := patcher.CheckOriginalData(ijl15)
	if orig {
		p.haveOrig = true
	} else {
		orig, err := patcher.UnpackOriginal(ijl15)
		if err == nil {
			p.haveOrig = patcher.CheckOriginalData(orig)
		}
	}
	p.rugburnVer, p.rugburnVerErr = patcher.GetRugburnVersion(ijl15)
	p.calc = ncalc
	return nil
}

func (p *RugburnPatcher) RugburnVersion() (string, error) {
	if err := p.recalc(); err != nil {
		return "", err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.rugburnVer, p.rugburnVerErr
}

func (p *RugburnPatcher) HaveOriginal() bool {
	if err := p.recalc(); err != nil {
		return false
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.haveOrig
}

func (p *RugburnPatcher) Patch() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ijl15, err := os.ReadFile(p.path)
	if err != nil {
		return err
	}

	if !patcher.CheckOriginalData(ijl15) {
		ijl15, err = patcher.UnpackOriginal(ijl15)
		if err != nil {
			return err
		}
		if !patcher.CheckOriginalData(ijl15) {
			return errors.New("couldn't recover original ijl15.dll for patching")
		}
	}

	rugburn, err := patcher.Patch(log.Default(), ijl15, embedded.RugburnDLL, embedded.Version)
	if err != nil {
		return err
	}

	err = os.WriteFile(p.path, rugburn, 0644)
	_ = p.recalc()
	return err
}

func (p *RugburnPatcher) Unpatch() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	rugburn, err := os.ReadFile(p.path)
	if err != nil {
		return err
	}

	ijl15, err := patcher.UnpackOriginal(rugburn)
	if err != nil {
		return err
	}

	err = os.WriteFile(p.path, ijl15, 0644)
	_ = p.recalc()
	return err
}
