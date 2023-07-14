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

package authserv

import (
	"net/http"
	"strconv"

	"github.com/rs/zerolog"
)

type Options struct {
	Logger zerolog.Logger
}

type Listener struct {
	log zerolog.Logger
}

func New(opts Options) *Listener {
	return &Listener{
		log: opts.Logger,
	}
}

func serveData(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func (l *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.log.Debug().Str("method", r.Method).Str("url", r.URL.String()).Msg("qa auth http request")
	switch r.URL.Path {
	case "/Secure/Login/LoginForGame.php", "/qalogin":
		serveData(w, ([]byte)(`<result>true</result><AuthKey>1234</AuthKey><MemberNo>1234</MemberNo><PCBangNo>1234</PCBangNo>`))
	}
}
