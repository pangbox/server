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

	"github.com/pangbox/server/database/accounts"
	"github.com/pangbox/server/gen/proto/go/topologypb/topologypbconnect"
	"github.com/pangbox/server/login"
	log "github.com/sirupsen/logrus"
)

type LoginOptions struct {
	Addr            string
	TopologyClient  topologypbconnect.TopologyServiceClient
	AccountsService *accounts.Service
}

type LoginServer struct {
	service *Service
}

func NewLoginServer(ctx context.Context) *LoginServer {
	login := new(LoginServer)
	login.service = NewService(ctx)
	return login
}

func (l *LoginServer) Configure(opts LoginOptions) error {
	spawn := func(ctx context.Context, service *Service) {
		loginServer := login.New(login.Options{
			TopologyClient:  opts.TopologyClient,
			AccountsService: opts.AccountsService,
		})

		service.SetShutdownFunc(func(shutdownCtx context.Context) error {
			return loginServer.Shutdown(shutdownCtx)
		})

		if ctx.Err() != nil {
			log.Errorf("LoginServer cancelled before server could start: %v", ctx.Err())
			return
		}

		err := loginServer.Listen(ctx, opts.Addr)
		if err != nil {
			log.Errorf("Error serving LoginServer: %v", err)
		}
	}

	return l.service.Configure(spawn)
}

func (l *LoginServer) Running() bool {
	return l.service.Running()
}

func (l *LoginServer) Start() error {
	return l.service.Start()
}

func (l *LoginServer) Stop() error {
	return l.service.Stop()
}
