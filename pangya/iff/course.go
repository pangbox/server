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

package iff

import "github.com/pangbox/server/pangya"

type Course struct {
	Active           bool
	_                [3]byte
	ID               uint32
	Name             string `struct:"[40]byte"`
	Level            byte
	Icon             string `struct:"[40]byte"`
	_                [3]byte
	Price            uint32
	DiscountPrice    uint32
	Condition        uint32
	ShopFlag         byte
	MoneyFlag        byte
	TimeFlag         byte
	TimeByte         byte
	Point            uint32
	Unknown          [0x1C]byte
	StartTime        pangya.SystemTime
	EndTime          pangya.SystemTime
	ShortName        string `struct:"[40]byte"`
	LocalizedName    string `struct:"[40]byte"`
	CourseFlag       byte
	PropertyFileName string `struct:"[40]byte"`
	Unknown2         uint32
	CourseSequence   string `struct:"[40]byte"`
}
