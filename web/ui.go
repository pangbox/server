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
	"crypto/md5"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

//go:embed assets/*
var assetFS embed.FS

//go:embed templates/*.html
var templateFS embed.FS

var templates = template.Must(template.ParseFS(templateFS, "templates/*.html"))

type RegisterPageParams struct {
	Errors []string
}

func (l *Handler) renderRegisterPage(w http.ResponseWriter, params RegisterPageParams) {
	if err := templates.ExecuteTemplate(w, "register", params); err != nil {
		log.Errorf("Error executing register template: %v", err)
	}
}

func (l *Handler) handleRegisterGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	l.renderRegisterPage(w, RegisterPageParams{})
}

func (l *Handler) handleRegisterPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	formdata, err := io.ReadAll(io.LimitReader(r.Body, maxFormSize))
	if err != nil {
		l.renderRegisterPage(w, RegisterPageParams{
			Errors: []string{err.Error()},
		})
		return
	}
	values, err := url.ParseQuery(string(formdata))
	if err != nil {
		l.renderRegisterPage(w, RegisterPageParams{
			Errors: []string{err.Error()},
		})
		return
	}
	formErrors := []string{}
	username := values.Get("username")
	password := values.Get("password")
	if len(username) < 3 {
		formErrors = append(formErrors, "Username too short.")
	}
	if len(username) > 22 {
		formErrors = append(formErrors, "Username too long.")
	}
	if len(password) < 5 {
		formErrors = append(formErrors, "Password too short.")
	}
	if len(formErrors) > 0 {
		l.renderRegisterPage(w, RegisterPageParams{
			Errors: formErrors,
		})
		return
	}

	// The client will MD5 the password before sending it.
	// TODO: only US
	passwordMD5 := md5.Sum([]byte(password))
	passwordMD5Hex := strings.ToUpper(hex.EncodeToString(passwordMD5[:]))

	_, err = l.accountsService.Register(r.Context(), username, passwordMD5Hex)
	if err != nil {
		l.renderRegisterPage(w, RegisterPageParams{
			Errors: []string{fmt.Sprintf("An error occurred: %v", err)},
		})
		return
	}

	if err := templates.ExecuteTemplate(w, "register_complete", nil); err != nil {
		log.Errorf("Error executing register template: %v", err)
	}
}
