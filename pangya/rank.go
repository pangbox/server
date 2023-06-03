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

// Rank represents an in-game rank.
type Rank byte

// These are the known possible in-game rank values.
const (
	RookieF         Rank = 0x00
	RookieE         Rank = 0x01
	RookieD         Rank = 0x02
	RookieC         Rank = 0x03
	RookieB         Rank = 0x04
	RookieA         Rank = 0x05
	BeginnerE       Rank = 0x06
	BeginnerD       Rank = 0x07
	BeginnerC       Rank = 0x08
	BeginnerB       Rank = 0x09
	BeginnerA       Rank = 0x0A
	JuniorE         Rank = 0x0B
	JuniorD         Rank = 0x0C
	JuniorC         Rank = 0x0D
	JuniorB         Rank = 0x0E
	JuniorA         Rank = 0x0F
	SeniorE         Rank = 0x10
	SeniorD         Rank = 0x11
	SeniorC         Rank = 0x12
	SeniorB         Rank = 0x13
	SeniorA         Rank = 0x14
	AmateurE        Rank = 0x15
	AmateurD        Rank = 0x16
	AmateurC        Rank = 0x17
	AmateurB        Rank = 0x18
	AmateurA        Rank = 0x19
	SemiProE        Rank = 0x1A
	SemiProD        Rank = 0x1B
	SemiProC        Rank = 0x1C
	SemiProB        Rank = 0x1D
	SemiProA        Rank = 0x1E
	ProE            Rank = 0x1F
	ProD            Rank = 0x20
	ProC            Rank = 0x21
	ProB            Rank = 0x22
	ProA            Rank = 0x23
	NationalProE    Rank = 0x24
	NationalProD    Rank = 0x25
	NationalProC    Rank = 0x26
	NationalProB    Rank = 0x27
	NationalProA    Rank = 0x28
	WorldProE       Rank = 0x29
	WorldProD       Rank = 0x2A
	WorldProC       Rank = 0x2B
	WorldProB       Rank = 0x2C
	WorldProA       Rank = 0x2D
	MasterE         Rank = 0x2E
	MasterD         Rank = 0x2F
	MasterC         Rank = 0x30
	MasterB         Rank = 0x31
	MasterA         Rank = 0x32
	TopMasterE      Rank = 0x33
	TopMasterD      Rank = 0x34
	TopMasterC      Rank = 0x35
	TopMasterB      Rank = 0x36
	TopMasterA      Rank = 0x37
	WorldMasterE    Rank = 0x38
	WorldMasterD    Rank = 0x39
	WorldMasterC    Rank = 0x3A
	WorldMasterB    Rank = 0x3B
	WorldMasterA    Rank = 0x3C
	LegendE         Rank = 0x3D
	LegendD         Rank = 0x3E
	LegendC         Rank = 0x3F
	LegendB         Rank = 0x40
	LegendA         Rank = 0x41
	InfinityLegendE Rank = 0x42
	InfinityLegendD Rank = 0x43
	InfinityLegendC Rank = 0x44
	InfinityLegendB Rank = 0x45
	InfinityLegendA Rank = 0x46
)
