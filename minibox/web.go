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

	"github.com/pangbox/pangfiles/crypto/pyxtea"
	"github.com/pangbox/pangfiles/pak"
	"github.com/pangbox/server/common/pycrypto"
	"github.com/pangbox/server/database/accounts"
	"github.com/pangbox/server/web"
	"github.com/rs/zerolog"
)

type WebOptions struct {
	Logger          zerolog.Logger
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
	log := opts.Logger
	spawn := func(ctx context.Context, service *Service) {
		webServer := http.Server{Addr: opts.Addr, Handler: web.New(web.Options{
			Logger:          log,
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
			log.Error().Err(ctx.Err()).Msg("cancelled before web server could start")
			return
		}

		err := webServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("error serving web server")
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

func getPakKey(log zerolog.Logger, region string, patterns []string) (pyxtea.Key, error) {
	if region == "" {
		log.Info().Msg("auto-detecting pak region (use -region to improve startup delay)")
		key, err := pak.DetectRegion(patterns, pycrypto.Keys)
		if err != nil {
			return pyxtea.Key{}, err
		}
		log.Info().Str("region", pycrypto.GetKeyRegion(key)).Msg("auto-detected pak region")
		return key, nil
	}
	return pycrypto.GetRegionKey(region)
}
