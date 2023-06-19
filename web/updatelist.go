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

package web

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pangbox/pangfiles/crypto/pyxtea"
	"github.com/pangbox/pangfiles/encoding/litexml"
	"github.com/pangbox/pangfiles/updatelist"
	log "github.com/sirupsen/logrus"
)

func (l *Handler) handleUpdateList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := l.updateHandler.updateList(w); err != nil {
		log.Printf("Error writing updateList: %v", err)

		w.WriteHeader(500)
		w.Write([]byte("server error"))
	}
}

type updateListCacheEntry struct {
	modTime time.Time
	fSize   int64
	fInfo   updatelist.FileInfo
}

type updateHandler struct {
	key pyxtea.Key

	dir string

	// +checklocks:mutex
	cache map[string]updateListCacheEntry

	mutex sync.RWMutex
}

func newUpdateListHandler(key pyxtea.Key, dir string) *updateHandler {
	ul := &updateHandler{
		key:   key,
		dir:   dir,
		cache: map[string]updateListCacheEntry{},
	}

	// Warm the updatelist.
	go ul.updateList(io.Discard)

	return ul
}

func (s *updateHandler) calcEntry(wg *sync.WaitGroup, entry *updatelist.FileInfo, f os.FileInfo) {
	defer wg.Done()
	var err error

	name := f.Name()
	*entry, err = updatelist.MakeFileInfo(s.dir, "", f, f.Size())

	if err != nil {
		log.Printf("Error calculating entry for %s: %s", name, err)
		entry.Filename = name
	} else {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		s.cache[name] = updateListCacheEntry{
			modTime: f.ModTime(),
			fSize:   f.Size(),
			fInfo:   *entry,
		}
	}
}

func (s *updateHandler) updateList(rw io.Writer) error {
	start := time.Now()

	files, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}

	doc := updatelist.Document{}
	doc.Info.Version = "1.0"
	doc.Info.Encoding = "euc-kr"
	doc.Info.Standalone = "yes"
	doc.PatchVer = "FakeVer"
	doc.PatchNum = 9999
	doc.UpdateListVer = "20090331"

	hit, miss := 0, 0

	var wg sync.WaitGroup
	doc.UpdateFiles.Files = make([]updatelist.FileInfo, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if !strings.HasSuffix(name, ".pak") {
			continue
		}

		s.mutex.RLock()
		cache, ok := s.cache[name]
		s.mutex.RUnlock()

		info, err := f.Info()
		if err != nil {
			panic(err)
		}
		if ok && cache.modTime == info.ModTime() && cache.fSize == info.Size() {
			// Cache hit
			hit++
			doc.UpdateFiles.Files = append(doc.UpdateFiles.Files, cache.fInfo)
			doc.UpdateFiles.Count++
		} else {
			// Cache miss, calculate concurrently.
			miss++
			doc.UpdateFiles.Files = append(doc.UpdateFiles.Files, updatelist.FileInfo{})
			doc.UpdateFiles.Count++
			entry := &doc.UpdateFiles.Files[len(doc.UpdateFiles.Files)-1]
			wg.Add(1)
			go s.calcEntry(&wg, entry, info)
		}
	}
	if doc.UpdateFiles.Count == 0 {
		log.Errorf("Did not find pak files; did you set -pangya_dir?")
	}

	wg.Wait()

	data, err := litexml.Marshal(doc)
	if err != nil {
		return err
	}

	if err := pyxtea.EncipherStreamPadNull(s.key, bytes.NewReader(data), rw); err != nil {
		return err
	}

	log.Printf("Updatelist calculated in %s (cache hits: %d, misses: %d)", time.Since(start), hit, miss)
	return nil
}
