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

package web

import (
	_ "embed"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

//go:embed data/translation.xml
var translation []byte
var translationB64 = []byte(base64.StdEncoding.EncodeToString(translation))

//go:embed data/extracontents.xml
var extraContents []byte

//go:embed data/pangya_default.xml
var pangyaDefault []byte

func (l *Handler) serveTranslations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Length", strconv.Itoa(len(translationB64)))
	w.Write(translationB64)
}

func (l *Handler) serveExtraContents(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Length", strconv.Itoa(len(extraContents)))
	w.Write(extraContents)
}

func (l *Handler) servePangyaDefault(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Length", strconv.Itoa(len(pangyaDefault)))
	w.Write(pangyaDefault)
}
