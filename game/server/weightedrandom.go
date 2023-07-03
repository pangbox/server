// Copyright (C) 2023, John Chadwick <john@jchw.io>, JMC47
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
// SPDX-FileCopyrightText: Copyright (c) 2023 John Chadwick, JMC47
// SPDX-License-Identifier: ISC

package gameserver

import (
	"errors"
	"math/rand"
	"sort"
)

type WeightedRand struct {
	cumulativeWeights []int64
	values            []uint32
	totalWeight       int64
}

func NewWeightedRand() *WeightedRand {
	return &WeightedRand{}
}

func (w *WeightedRand) Add(value uint32, weight int64) error {
	// Detect if the total weight is going to overflow
	if w.totalWeight+weight < w.totalWeight {
		return errors.New("total weight will overflow")
	}

	// Update total weight
	w.totalWeight += weight

	// Append the value
	w.values = append(w.values, value)

	// Compute cumulative weight
	var cumulativeWeight int64
	if len(w.cumulativeWeights) > 0 {
		cumulativeWeight = w.cumulativeWeights[len(w.cumulativeWeights)-1] + weight
	} else {
		cumulativeWeight = weight
	}
	w.cumulativeWeights = append(w.cumulativeWeights, cumulativeWeight)

	return nil
}

func (w *WeightedRand) Choose() uint32 {
	// Generate a random number in the range [0, totalWeight)
	r := rand.Int63n(w.totalWeight)

	// Use binary search to find the index where our random number fits in
	index := sort.Search(len(w.cumulativeWeights), func(i int) bool { return w.cumulativeWeights[i] > r })

	// Return the corresponding value
	return w.values[index]
}
