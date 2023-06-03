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

package pubsub

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var (
	_ = PubSub(&PostgresPubSub{})
)

// GetPostgresListenerFunc is a type of function that returns a new
// pq.Listener ready to start listening to pubsub channels. You must provide
// such a function to PostgresPubSub.
type GetPostgresListenerFunc func() (*pq.Listener, error)

// PostgresStream is a handler for a single PostgreSQL subscription.
type PostgresStream struct {
	listener *pq.Listener
	logger   *log.Entry
	stream   chan Message
}

func newPostgresStream(listener *pq.Listener, channel string, log *log.Entry) (*PostgresStream, error) {
	stream := &PostgresStream{
		listener: listener,
		logger:   log.WithField("channel", channel),
		stream:   make(chan Message),
	}

	err := stream.listen(channel)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

// listen begins listening on a given channel.
func (s *PostgresStream) listen(channel string) error {
	err := s.listener.Listen(channel)
	if err != nil {
		return err
	}

	go func() {
		defer close(s.stream)

		for {
			select {
			case data, ok := <-s.listener.Notify:
				if !ok {
					return
				}
				message := Message{}
				err := json.Unmarshal([]byte(data.Extra), &message)
				if err != nil {
					s.logger.Error(err)
					break
				}
				s.stream <- message
			case <-time.After(30 * time.Second):
				go s.listener.Ping()
			}

		}
	}()

	return nil
}

// Channel returns a channel of streaming messages.
func (s *PostgresStream) Channel() <-chan Message {
	return s.stream
}

// Close ends the pubsub stream.
func (s *PostgresStream) Close() error {
	return s.listener.Close()
}

// PostgresPubSub is an implementation of pubsub using PostgreSQL
// notifications.
type PostgresPubSub struct {
	getListener GetPostgresListenerFunc
	chanPrefix  string
	logger      *log.Entry
}

// NewPostgresPubSub creates a new PostgreSQL publish/subscribe engine.
func NewPostgresPubSub(getListener GetPostgresListenerFunc, chanPrefix string, log *log.Entry) (*PostgresPubSub, error) {
	return &PostgresPubSub{
		getListener: getListener,
		chanPrefix:  chanPrefix,
		logger:      log.WithField("pubsub", "postgres"),
	}, nil
}

// Publish publishes a message on a given channel.
func (p *PostgresPubSub) Publish(channel string, message Message) error {
	return nil
}

// Subscribe listens for messages on a given channel.
func (p *PostgresPubSub) Subscribe(channel string) (Stream, error) {
	listener, err := p.getListener()
	if err != nil {
		return nil, err
	}

	stream, err := newPostgresStream(listener, p.chanPrefix+channel, p.logger)
	if err != nil {
		listener.Close()
		return nil, err
	}

	return stream, nil
}
