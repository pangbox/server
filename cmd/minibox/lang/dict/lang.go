// Copyright (c) 2012 The polyglot Authors.
// Copyright (c) 2023 John Chadwick
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. The names of the authors may not be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHORS ``AS IS'' AND ANY EXPRESS OR
// IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
// THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// SPDX-FileCopyrightText: Copyright (c) 2012 The polyglot Authors.
// SPDX-FileCopyrightText: Copyright (c) 2023 John Chadwick
// SPDX-License-Identifier: BSD-3-Clause
//
// Based on https://github.com/lxn/polyglot.

package dict

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/pangbox/server/cmd/minibox/lang"
)

var (
	// ErrInvalidLocale is returned if a specified locale is invalid.
	ErrInvalidLocale = errors.New("invalid locale")
)

// Dict provides translated strings appropriate for a specific locale.
type Dict struct {
	dirPath                string
	locales                []string
	locale2SourceKey2Trans map[string]map[string]string
}

// NewDict returns a new Dict with the specified locale.
func NewDict(locale string) (*Dict, error) {
	locales := localesChainForLocale(locale)
	if len(locales) == 0 {
		return nil, ErrInvalidLocale
	}

	d := &Dict{
		locales:                locales,
		locale2SourceKey2Trans: make(map[string]map[string]string),
	}

	if err := d.loadTranslations(); err != nil {
		return nil, err
	}

	return d, nil
}

// DirPath returns the translations directory path of the Dict.
func (d *Dict) DirPath() string {
	return d.dirPath
}

// Locale returns the locale of the Dict.
func (d *Dict) Locale() string {
	return d.locales[0]
}

// Translation returns a translation of the source string to the locale of the
// Dict or the source string, if no matching translation was found.
//
// Provided context arguments are used for disambiguation.
func (d *Dict) Translation(source string, context ...string) string {
	if d == nil {
		return source
	}

	for _, locale := range d.locales {
		if sourceKey2Trans, ok := d.locale2SourceKey2Trans[locale]; ok {
			if trans, ok := sourceKey2Trans[sourceKey(source, context)]; ok && trans != "" {
				return trans
			}
		}
	}

	return source
}

func (d *Dict) loadTranslation(reader io.Reader, locale string) error {
	var trf lang.Translation

	if err := json.NewDecoder(reader).Decode(&trf); err != nil {
		return err
	}

	sourceKey2Trans, ok := d.locale2SourceKey2Trans[locale]
	if !ok {
		sourceKey2Trans = make(map[string]string)

		d.locale2SourceKey2Trans[locale] = sourceKey2Trans
	}

	for _, m := range trf.Messages {
		if m.Translation != "" {
			sourceKey2Trans[sourceKey(m.Source, m.Context)] = m.Translation
		}
	}

	return nil
}

func (d *Dict) loadTranslations() error {
	dirPath := "."

	entries, err := lang.TranslationFS.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := path.Join(dirPath, entry.Name())
		if locale := d.matchingLocaleFromFileName(entry.Name()); locale != "" {
			file, err := lang.TranslationFS.Open(fullPath)
			if err != nil {
				return err
			}
			defer file.Close()

			if err := d.loadTranslation(file, locale); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Dict) matchingLocaleFromFileName(name string) string {
	for _, locale := range d.locales {
		if name == fmt.Sprintf("%s.json", locale) {
			return locale
		}
	}

	return ""
}

func sourceKey(source string, context []string) string {
	if len(context) == 0 {
		return source
	}

	return fmt.Sprintf("__%s__%s__", source, strings.Join(context, "__"))
}

func localesChainForLocale(locale string) []string {
	parts := strings.Split(locale, "_")
	if len(parts) > 2 {
		return nil
	}

	if len(parts[0]) != 2 {
		return nil
	}

	for _, r := range parts[0] {
		if r < rune('a') || r > rune('z') {
			return nil
		}
	}

	if len(parts) == 1 {
		return []string{parts[0]}
	}

	if len(parts[1]) < 2 || len(parts[1]) > 3 {
		return nil
	}

	for _, r := range parts[1] {
		if r < rune('A') || r > rune('Z') {
			return nil
		}
	}

	return []string{locale, parts[0]}
}
