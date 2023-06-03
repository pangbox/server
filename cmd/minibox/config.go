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

package main

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"

	"github.com/pangbox/server/minibox"
)

var (
	opts = minibox.Options{
		WebAddr:         ":8080",
		AdminAddr:       ":8081",
		QAAuthAddr:      ":8090",
		LoginAddr:       ":10101",
		GameAddr:        ":20202",
		MessageAddr:     ":30303",
		ServerIP:        "127.0.0.1",
		GameServerName:  "Pangbox",
		GameChannelName: "Snowblind",
		PangyaRegion:    "",
		PangyaDir:       ".",
		PangyaIFF:       "",
	}
	dbOpts = minibox.DataOptions{
		DatabaseURI: "sqlite://pangbox.sqlite3",
	}
	// Only used in GUI.
	language = ""
)

func init() {
	flag.StringVar(&opts.WebAddr, "web_addr", opts.WebAddr, "Address to listen on for webserver connections.")
	flag.StringVar(&opts.AdminAddr, "admin_addr", opts.AdminAddr, "Address to listen on for admin control panel.")
	flag.StringVar(&opts.QAAuthAddr, "qaauth_addr", opts.QAAuthAddr, "Address to listen on for QA authentication connections.")
	flag.StringVar(&opts.LoginAddr, "login_addr", opts.LoginAddr, "Address to listen on for login server connections.")
	flag.StringVar(&opts.GameAddr, "game_addr", opts.GameAddr, "Address to listen on for game server connections.")
	flag.StringVar(&opts.MessageAddr, "message_addr", opts.MessageAddr, "Address to listen on for message server connections.")
	flag.StringVar(&opts.ServerIP, "server_ip", opts.ServerIP, "IP address to advertise.")
	flag.StringVar(&opts.GameServerName, "game_server_name", opts.GameServerName, "Name of game server.")
	flag.StringVar(&opts.GameChannelName, "game_channel_name", opts.GameChannelName, "Name of game channel.")
	flag.StringVar(&opts.PangyaRegion, "pangya_region", opts.PangyaRegion, "Region of client, or auto-detect.")
	flag.StringVar(&opts.PangyaDir, "pangya_dir", opts.PangyaDir, "Directory of PangYa client.")
	flag.StringVar(&opts.PangyaIFF, "pangya_iff", opts.PangyaIFF, "OPTIONAL: Client IFF to load. Overrides the IFF found in the pak files if specified.")
	flag.StringVar(&dbOpts.DatabaseURI, "database", dbOpts.DatabaseURI, "Database URI.")
	flag.StringVar(&language, "lang", language, "Language to use in the UI, if enabled.")
}

type pangboxConfig struct {
	Database minibox.DataOptions `json:"Database"`
	Options  minibox.Options     `json:"Options"`
	Language string              `json:"Language,omitempty"`
}

// saveConfiguration saves the configuration to disk.
func saveConfiguration(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	return json.NewEncoder(f).Encode(pangboxConfig{
		Database: dbOpts,
		Options:  opts,
		Language: language,
	})
}

// loadConfiguration loads the configuration from disk.
func loadConfiguration(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	config := pangboxConfig{}
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return err
	}

	copyNotEmptyFields(&dbOpts, &config.Database)
	copyNotEmptyFields(&opts, &config.Options)
	if config.Language != "" {
		language = config.Language
	}
	return nil
}

func copyNotEmptyFields(dst, src any) {
	dstv := reflect.ValueOf(dst)
	srcv := reflect.ValueOf(src)
	for dstv.Type().Kind() == reflect.Ptr {
		dstv = dstv.Elem()
	}
	for srcv.Type().Kind() == reflect.Ptr {
		srcv = srcv.Elem()
	}
	if dstv.Type() != srcv.Type() {
		panic("bad usage")
	}
	for i := 0; i < dstv.NumField(); i++ {
		sf := srcv.Field(i)
		if reflect.New(sf.Type()).Elem().Equal(sf) {
			// field is equal to zero value
			continue
		}
		dstv.Field(i).Set(sf)
	}
}
