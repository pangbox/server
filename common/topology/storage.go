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
	"encoding/binary"
	"errors"

	"github.com/bufbuild/connect-go"
	"github.com/pangbox/server/gen/proto/go/topologypb"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/protobuf/proto"
)

var (
	errServerNotFound = connect.NewError(connect.CodeNotFound, errors.New("server with ID not found"))
)

// Storage implementations implement a storage layer for the topology.
type Storage interface {
	Get(id uint16) (*topologypb.ServerEntry, error)
	Put(id uint16, server *topologypb.ServerEntry) error
	List() ([]*topologypb.ServerEntry, error)
}

// MemoryStorage implements topology storage on top of memory. MemoryStorage
// defensively deep-copies all protobufs to prevent any pointers from being
// shared amongst users and the storage layers, similar to a database engine
// that would require marshalling and unmarshalling.
type MemoryStorage struct {
	servers   []*topologypb.ServerEntry
	serverMap map[uint16]*topologypb.ServerEntry
}

// NewMemoryStorage creates a new memory store with the provided servers.
func NewMemoryStorage(entries []*topologypb.ServerEntry) *MemoryStorage {
	store := &MemoryStorage{
		servers:   []*topologypb.ServerEntry{},
		serverMap: map[uint16]*topologypb.ServerEntry{},
	}

	for _, entry := range entries {
		_ = store.Put(uint16(entry.Server.Id), entry)
	}

	return store
}

// Get implements topology.Storage.
func (s *MemoryStorage) Get(id uint16) (*topologypb.ServerEntry, error) {
	if entry, ok := s.serverMap[id]; ok {
		return proto.Clone(entry).(*topologypb.ServerEntry), nil
	}
	return nil, errServerNotFound
}

// Put implements topology.Storage.
func (s *MemoryStorage) Put(id uint16, entry *topologypb.ServerEntry) error {
	s.serverMap[id] = proto.Clone(entry).(*topologypb.ServerEntry)
	for i := range s.servers {
		if uint16(s.servers[i].Server.Id) == id {
			s.servers[i] = proto.Clone(entry).(*topologypb.ServerEntry)
			return nil
		}
	}
	s.servers = append(s.servers, proto.Clone(entry).(*topologypb.ServerEntry))
	return nil
}

// List implements topology.Storage.
func (s *MemoryStorage) List() ([]*topologypb.ServerEntry, error) {
	result := make([]*topologypb.ServerEntry, len(s.servers))
	for i, server := range s.servers {
		result[i] = proto.Clone(server).(*topologypb.ServerEntry)
	}
	return result, nil
}

// LevelDBStorage implements topology storage on top of LevelDB.
type LevelDBStorage struct {
	db *leveldb.DB
}

// NewLevelDBStorage creates a new LevelDB-backed storage engine.
func NewLevelDBStorage(db *leveldb.DB) *LevelDBStorage {
	return &LevelDBStorage{db}
}

func (s *LevelDBStorage) toKey(id uint16) []byte {
	k := [2]byte{}
	binary.BigEndian.PutUint16(k[:], id)
	return k[:]
}

func (s *LevelDBStorage) translateErr(err error) error {
	switch err {
	case leveldb.ErrNotFound:
		return errServerNotFound
	default:
		return err
	}
}

// Get implements topology.Storage.
func (s *LevelDBStorage) Get(id uint16) (*topologypb.ServerEntry, error) {
	val, err := s.db.Get(s.toKey(id), nil)
	if err != nil {
		return nil, s.translateErr(err)
	}
	result := &topologypb.ServerEntry{}
	if err := proto.Unmarshal(val, result); err != nil {
		return nil, err
	}
	return result, nil
}

// Put implements topology.Storage.
func (s *LevelDBStorage) Put(id uint16, server *topologypb.ServerEntry) error {
	val, err := proto.Marshal(server)
	if err != nil {
		return err
	}

	return s.db.Put(s.toKey(id), val, nil)
}

// List implements topology.Storage.
func (s *LevelDBStorage) List() ([]*topologypb.ServerEntry, error) {
	result := []*topologypb.ServerEntry{}
	iter := s.db.NewIterator(nil, nil)
	defer iter.Release()
	for iter.Next() {
		entry := &topologypb.ServerEntry{}
		if err := proto.Unmarshal(iter.Value(), entry); err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return result, nil
}
