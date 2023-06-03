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
	"context"
	"net/http"
	"strings"

	"github.com/pangbox/pangfiles/crypto/pyxtea"
	"github.com/pangbox/pangfiles/pak"
	"github.com/pangbox/server/common/pycrypto"
	"github.com/pangbox/server/database/accounts"
	"github.com/pangbox/server/web"
	log "github.com/sirupsen/logrus"
)

type WebOptions struct {
	Addr            string
	PangyaKey       pyxtea.Key
	PangyaDir       string
	AccountsService *accounts.Service
}

type WebServer struct {
	service *Service
}

func NewWeb(ctx context.Context) *WebServer {
	web := new(WebServer)
	web.service = NewService(ctx)
	return web
}

func (w *WebServer) Configure(opts WebOptions) error {
	spawn := func(ctx context.Context, service *Service) {
		webServer := http.Server{Addr: opts.Addr, Handler: web.New(web.Options{
			ServePangYaData: true,
			UpdateList: &web.UpdateListOptions{
				Key: opts.PangyaKey,
				Dir: opts.PangyaDir,
			},
			AccountsService: opts.AccountsService,
		})}

		service.SetShutdownFunc(func(shutdownCtx context.Context) error {
			return webServer.Shutdown(shutdownCtx)
		})

		if ctx.Err() != nil {
			log.Errorf("Web server cancelled before server could start: %v", ctx.Err())
			return
		}

		err := webServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("Error serving web server: %v", err)
		}
	}

	return w.service.Configure(spawn)
}

func (w *WebServer) Running() bool {
	return w.service.Running()
}

func (w *WebServer) Start() error {
	return w.service.Start()
}

func (w *WebServer) Stop() error {
	return w.service.Stop()
}

func getPakKey(log *log.Entry, region string, patterns []string) (pyxtea.Key, error) {
	if region == "" {
		log.Println("Auto-detecting pak region (use -region to improve startup delay.)")
		key, err := pak.DetectRegion(patterns, pycrypto.Keys)
		if err != nil {
			return pyxtea.Key{}, err
		}
		log.Printf("Detected pak region as %s.", strings.ToUpper(pycrypto.GetKeyRegion(key)))
		return key, nil
	}
	return pycrypto.GetRegionKey(region)
}
