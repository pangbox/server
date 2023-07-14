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

package common

import (
	"fmt"
	"reflect"
)

type Message interface{}

type MessageTable[T Message] struct {
	IDToMessage map[uint16]T
	MessageToID map[reflect.Type]uint16
}

type AnyMessageTable interface {
	ID(msg Message) (uint16, error)
	Build(id uint16) (Message, error)
}

func NewMessageTable[T Message](index map[uint16]T) MessageTable[T] {
	table := MessageTable[T]{
		IDToMessage: index,
		MessageToID: make(map[reflect.Type]uint16),
	}
	for id, msg := range index {
		typ := reflect.TypeOf(msg)
		if otherId, ok := table.MessageToID[typ]; ok {
			panic(fmt.Errorf("conflict: multiple IDs for message %T: 0x%04x and 0x%04x", msg, id, otherId))
		}
		table.MessageToID[typ] = id
	}
	return table
}

func (table MessageTable[T]) Any() AnyMessageTable {
	anytable := &MessageTable[Message]{
		IDToMessage: make(map[uint16]Message),
		MessageToID: make(map[reflect.Type]uint16),
	}
	for id, msg := range table.IDToMessage {
		anytable.IDToMessage[id] = msg
		anytable.MessageToID[reflect.TypeOf(msg)] = id
	}
	return anytable
}

func (table MessageTable[T]) ID(msg T) (uint16, error) {
	id, ok := table.MessageToID[reflect.TypeOf(msg)]
	if !ok {
		return id, UnknownMessageError{MessageID: id}
	}
	return id, nil
}

func (table MessageTable[T]) Build(id uint16) (T, error) {
	message, ok := table.IDToMessage[id]
	if !ok {
		return message, UnknownMessageError{MessageID: id}
	}
	return reflect.New(reflect.TypeOf(message).Elem()).Interface().(T), nil
}
