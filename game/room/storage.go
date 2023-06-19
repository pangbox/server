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

package room

import (
	"container/heap"
	"context"
	"errors"
	"sync"

	"github.com/davecgh/go-spew/spew"
	gamemodel "github.com/pangbox/server/game/model"
)

// The primary job of this code is to manage the allocation of rooms in a
// channel.

type RoomEntry struct {
	// Room instance
	room Room

	// A copy of the state that gets synced to us
	state gamemodel.RoomState

	// The heap index
	heapIndex int
}

// RoomHeap implements a heap.Interface for the room list.
// It provides a min heap such that the lowest-numbered inactive room will be
// at the lowest index.
type RoomHeap []*RoomEntry

type Storage struct {
	// Rooms sorted by room number. Guaranteed to be contiguous.
	roomsByNumber []*RoomEntry
	resizeMu      sync.RWMutex

	// Room min-heap over room number. Used to find inactive room slots to
	// use when creating new rooms.
	roomHeap RoomHeap
}

// NewRoom returns a new room.
func (m *Storage) NewRoom(ctx context.Context) *Room {
	if len(m.roomHeap) == 0 {
		return m.allocRoom(ctx)
	}
	lowestRoom := m.roomHeap[0]
	if lowestRoom.state.Active {
		return m.allocRoom(ctx)
	}
	lowestRoom.state.Active = true
	heap.Fix(&m.roomHeap, lowestRoom.heapIndex)
	return &lowestRoom.room
}

// UpdateRoom updates the room state for a given room.
func (m *Storage) UpdateRoom(ctx context.Context, state gamemodel.RoomState) error {
	n := int(state.RoomNumber)

	m.resizeMu.RLock()
	roomsByNumber := m.roomsByNumber
	m.resizeMu.RUnlock()

	if n >= len(roomsByNumber) {
		return errors.New("room number too large")
	}
	entry := roomsByNumber[n]
	entry.state = state

	heap.Fix(&m.roomHeap, entry.heapIndex)

	// TODO: we should try to only cull when it's not too busy
	if !state.Active {
		m.cull()
	}

	return nil
}

// GetRoom gets a room by ID. This can be called from any thread.
func (m *Storage) GetRoom(ctx context.Context, roomNumber int16) *Room {
	m.resizeMu.RLock()
	roomsByNumber := m.roomsByNumber
	m.resizeMu.RUnlock()

	if int(roomNumber) < len(roomsByNumber) {
		return &roomsByNumber[roomNumber].room
	}

	return nil
}

// GetRoomList returns a full list of rooms.
func (m *Storage) GetRoomList() []*Room {
	results := []*Room{}

	m.resizeMu.RLock()
	roomsByNumber := m.roomsByNumber
	m.resizeMu.RUnlock()

	for _, room := range roomsByNumber {
		if room.state.Active {
			results = append(results, &room.room)
			spew.Dump(room.state)
		}
	}

	return results
}

func (m *Storage) allocRoom(ctx context.Context) *Room {
	entry := new(RoomEntry)

	m.resizeMu.Lock()
	n := len(m.roomsByNumber)
	m.roomsByNumber = append(m.roomsByNumber, entry)
	m.resizeMu.Unlock()

	entry.state.RoomNumber = int16(n)
	entry.state.Active = true
	entry.room.state.RoomNumber = entry.state.RoomNumber
	heap.Push(&m.roomHeap, entry)
	return &entry.room
}

// cull removes any extraneous rooms from the list/heap
func (m *Storage) cull() {
	// We don't really want to cause too much contention.
	// If things are busy, this process can just wait.
	/*
		if !m.resizeMu.TryLock() {
			return
		}
		defer m.resizeMu.Unlock()

		i := len(m.roomsByNumber) - 1
		for ; i >= 0; i-- {
			if m.roomsByNumber[i].state.Active {
				break
			}
		}

		for j, l := i+1, len(m.roomsByNumber); j < l; j++ {
			heap.Remove(&m.roomHeap, m.roomsByNumber[j].heapIndex)
			m.roomsByNumber[j] = nil
		}

		m.roomsByNumber = m.roomsByNumber[:i+1]
	*/
}

func (h RoomHeap) Len() int { return len(h) }

func (h RoomHeap) Less(i, j int) bool {
	if h[i].state.Active != h[j].state.Active {
		ia := 0
		if h[i].state.Active {
			ia = 1
		}
		ja := 0
		if h[j].state.Active {
			ja = 1
		}
		return ia < ja
	}
	return h[i].state.RoomNumber < h[j].state.RoomNumber
}

func (h RoomHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].heapIndex = i
	h[j].heapIndex = j
}

func (h *RoomHeap) Push(x any) {
	n := len(*h)
	room := x.(*RoomEntry)
	room.heapIndex = n
	*h = append(*h, room)
}

func (h *RoomHeap) Pop() any {
	old := *h
	n := len(old)
	room := old[n-1]
	old[n-1] = nil
	room.heapIndex = -1
	*h = old[0 : n-1]
	return room
}
