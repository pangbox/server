// Copyright (c) 2012 The polyglot Authors.
// Copyright (c) 2023 John Chadwick
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. The names of the authors may not be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHORS ``AS IS'' AND ANY EXPRESS OR
// IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
// THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// SPDX-License-Identifier: BSD-3-Clause
//
// Based on https://github.com/lxn/polyglot.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pangbox/server/cmd/minibox/lang"
	"github.com/rs/zerolog"
	"golang.org/x/exp/slices"
)

var (
	outPath   = flag.String("out", "", "The directory to output locale files into.")
	srcPath   = flag.String("src", "", "The directory path where to recursively search for Go files.")
	locales   = flag.String("locales", "", `Comma-separated list of locales, for which to generate or update tr files. e.g.: "de_AT,de_DE,de,es,fr,it".`)
	defLocale = flag.String("default", "en", "Locale that should be treated as the default locale.")
)

func sourceKey(source string, context []string) string {
	if len(context) == 0 {
		return source
	}

	return fmt.Sprintf("__%s__%s__", source, strings.Join(context, "__"))
}

type visitor struct {
	fileSet           *token.FileSet
	sourceKey2Message map[string]*lang.Message
}

func (v visitor) Visit(node ast.Node) (w ast.Visitor) {
	if callExpr, ok := node.(*ast.CallExpr); ok {
		if ident, ok := callExpr.Fun.(*ast.Ident); !ok || ident.Name != "tr" {
			return v
		}

		if len(callExpr.Args) > 0 {
			if basicLit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
				source := string(basicLit.Value[1 : len(basicLit.Value)-1])
				var context []string
				for _, arg := range callExpr.Args[1:] {
					if basicLit, ok := arg.(*ast.BasicLit); ok {
						c := string(basicLit.Value[1 : len(basicLit.Value)-1])
						context = append(context, c)
					}
				}
				srcKey := sourceKey(source, context)
				v.sourceKey2Message[srcKey] = &lang.Message{Source: source, Context: context}
			}
		}
	}

	return v
}

func (v visitor) scanDir(log zerolog.Logger, dirPath string) {
	dir, err := os.Open(dirPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", dirPath).Msg("opening directory")
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		log.Fatal().Err(err).Msg("reading directory")
	}

	for _, name := range names {
		fullPath := path.Join(dirPath, name)

		fi, err := os.Stat(fullPath)
		if err != nil {
			log.Fatal().Err(err).Str("path", fullPath).Msg("statting path")
		}

		if fi.IsDir() {
			v.scanDir(log, fullPath)
		} else if !fi.IsDir() && strings.HasSuffix(fullPath, ".go") {
			astFile, err := parser.ParseFile(v.fileSet, fullPath, nil, 0)
			if err != nil {
				log.Fatal().Err(err).Str("path", fullPath).Msg("parsing file")
			}

			ast.Walk(v, astFile)
		}
	}
}

func readTranslation(log zerolog.Logger, filePath string) map[string]*lang.Message {
	log = log.With().Str("path", filePath).Logger()
	sk2m := make(map[string]*lang.Message)

	if fi, _ := os.Stat(filePath); fi == nil {
		return sk2m
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("opening file")
	}
	defer file.Close()

	var trf lang.Translation
	if err := json.NewDecoder(file).Decode(&trf); err != nil {
		log.Fatal().Err(err).Msg("parsing json from file")
	}

	for _, msg := range trf.Messages {
		sk2m[sourceKey(msg.Source, msg.Context)] = msg
	}

	return sk2m
}

func writeTranslation(log zerolog.Logger, filePath string, trf lang.Translation) {
	log = log.With().Str("path", filePath).Logger()
	slices.SortFunc(trf.Messages, func(a, b *lang.Message) bool {
		if a.Source == b.Source {
			return strings.Join(a.Context, "") < strings.Join(b.Context, "")
		}
		return a.Source < b.Source
	})

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("creating file")
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "\t")
	if err := enc.Encode(trf); err != nil {
		log.Fatal().Err(err).Msg("writing json to file")
	}
}

func writeUpdatedTranslation(log zerolog.Logger, filePath string, sourceKey2Message, oldSourceKey2Message map[string]*lang.Message, loc string) {
	var trf lang.Translation
	for _, msg := range sourceKey2Message {
		msgCopy := *msg

		if oldMsg, ok := oldSourceKey2Message[sourceKey(msg.Source, msg.Context)]; ok {
			msgCopy.Translation = oldMsg.Translation
		}

		trf.Messages = append(trf.Messages, &msgCopy)
	}

	writeTranslation(log, filePath, trf)
}

func writeDefaultTranslation(log zerolog.Logger, filePath string, sourceKey2Message map[string]*lang.Message, loc string) {
	var trf lang.Translation
	for _, msg := range sourceKey2Message {
		msgCopy := *msg
		msgCopy.Translation = msg.Source
		trf.Messages = append(trf.Messages, &msgCopy)
	}

	writeTranslation(log, filePath, trf)
}

func main() {
	flag.Parse()

	log := zerolog.New(os.Stderr)

	if *srcPath == "" || *locales == "" {
		flag.Usage()
		os.Exit(1)
	}

	v := visitor{
		fileSet:           token.NewFileSet(),
		sourceKey2Message: make(map[string]*lang.Message),
	}

	v.scanDir(log, *srcPath)

	locs := strings.Split(*locales, ",")
	for _, loc := range locs {
		loc = strings.TrimSpace(loc)

		filePath := filepath.Join(*outPath, fmt.Sprintf("%s.json", loc))

		if loc == *defLocale {
			writeDefaultTranslation(log, filePath, v.sourceKey2Message, loc)
		} else {
			writeUpdatedTranslation(log, filePath, v.sourceKey2Message, readTranslation(log, filePath), loc)
		}
	}
}
