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

package pangya

type SystemTime struct {
	/* 0x00 */ Year, Month, DayOfWeek, Day uint16
	/* 0x08 */ Hour, Minute, Second, Milliseconds uint16
	/* 0x10 */
}

func (s SystemTime) IsZero() bool {
	return (s.Year == 0 && s.Month == 0 && s.DayOfWeek == 0 && s.Day == 0 &&
		s.Hour == 0 && s.Minute == 0 && s.Second == 0 && s.Milliseconds == 0)
}

func (s SystemTime) IsValid() bool {
	if s.Year < 1601 || s.Year > 30827 {
		return false
	}
	if s.Month < 1 || s.Month > 12 {
		return false
	}
	if s.DayOfWeek > 6 {
		return false
	}
	if s.Day < 1 || s.Day > 31 {
		return false
	}
	if s.Hour > 23 {
		return false
	}
	if s.Minute > 59 {
		return false
	}
	if s.Second > 59 {
		return false
	}
	if s.Milliseconds > 999 {
		return false
	}
	return true
}
