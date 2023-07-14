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
	"github.com/pangbox/server/message"
	"github.com/rs/zerolog"
)

type MessageOptions struct {
	Logger          zerolog.Logger
	Addr            string
	TopologyClient  topologypbconnect.TopologyServiceClient
	AccountsService *accounts.Service
}

type MessageServer struct {
	service *Service
}

func NewMessageServer(ctx context.Context) *MessageServer {
	message := new(MessageServer)
	message.service = NewService(ctx)
	return message
}

func (m *MessageServer) Configure(opts MessageOptions) error {
	log := opts.Logger
	spawn := func(ctx context.Context, service *Service) {
		messageServer := message.New(message.Options{
			Logger:          opts.Logger,
			TopologyClient:  opts.TopologyClient,
			AccountsService: opts.AccountsService,
		})

		service.SetShutdownFunc(func(shutdownCtx context.Context) error {
			return messageServer.Shutdown(shutdownCtx)
		})

		if ctx.Err() != nil {
			log.Error().Err(ctx.Err()).Msg("cancelled before message server could start")
			return
		}

		err := messageServer.Listen(ctx, opts.Addr)
		if err != nil {
			log.Error().Err(err).Msg("error serving message server")
		}
	}

	return m.service.Configure(spawn)
}

func (m *MessageServer) Running() bool {
	return m.service.Running()
}

func (m *MessageServer) Start() error {
	return m.service.Start()
}

func (m *MessageServer) Stop() error {
	return m.service.Stop()
}
