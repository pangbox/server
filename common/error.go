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

import "fmt"

// UnknownMessageError is returned when there is no known message ID in a given
// context.
type UnknownMessageError struct{ MessageID uint16 }

// Error implements the error interface.
func (e UnknownMessageError) Error() string {
	return fmt.Sprintf("unknown message %04x", e.MessageID)
}

// UnexpectedMessageError is returned when there an unexpected message is
// received.
type UnexpectedMessageError struct{ MessageID uint16 }

// Error implements the error interface.
func (e UnexpectedMessageError) Error() string {
	return fmt.Sprintf("unexpected message %04x", e.MessageID)
}
