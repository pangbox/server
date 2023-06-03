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

package topology

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/pangbox/server/gen/proto/go/topologypb/topologypbconnect"
	"golang.org/x/net/http2"
)

const DefaultDialTimeout = 10 * time.Second

type ClientOptions struct {
	// BaseURL specifies the base URL to send requests to. Use the h2c:// scheme to use H2C.
	BaseURL string

	// DialTimeout specifies the maximum amount of time to dial the topology server in a request.
	// If unspecified, this will be [DefaultDialTimeout].
	DialTimeout time.Duration
}

func NewClient(options ClientOptions) (topologypbconnect.TopologyServiceClient, error) {
	baseURL, err := url.Parse(options.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing topology server URL: %w", err)
	}

	if options.DialTimeout == 0 {
		options.DialTimeout = DefaultDialTimeout
	}

	transport := &http2.Transport{}
	netDialer := net.Dialer{Timeout: options.DialTimeout}
	if baseURL.Scheme == "h2c" {
		baseURL.Scheme = "https"
		transport.DialTLSContext = func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return netDialer.DialContext(ctx, network, addr)
		}
	} else {
		transport.DialTLSContext = func(ctx context.Context, network, addr string, tlsConfig *tls.Config) (net.Conn, error) {
			tlsDialer := tls.Dialer{
				NetDialer: &netDialer,
				Config:    tlsConfig,
			}
			return tlsDialer.DialContext(ctx, network, addr)
		}
	}
	client := &http.Client{
		Transport: transport,
	}
	return topologypbconnect.NewTopologyServiceClient(client, baseURL.String()), nil
}
