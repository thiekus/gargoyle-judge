package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

type LoginFormData struct {
	Username string
	Password string
	Target   string
}

func loginGetEndpoint(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	lfd := LoginFormData{Target: r.FormValue("target")}
	CompileSinglePage(w, r, "login.html", lfd)
}

func loginPostEndpoint(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	err := appUsers.UserLogin(w, r, username, password)
	target := r.PostFormValue("target")
	if err != nil {
		appUsers.AddFlashMessage(w, r, fmt.Sprintf("Error: %s", err), FlashError)
		if target != "" {
			http.Redirect(w, r, getBaseUrl(r)+"login?target="+target, 302)
		} else {
			http.Redirect(w, r, getBaseUrl(r)+"login", 302)
		}
	} else {
		// Login success
		if target != "" {
			targetDec, _ := base64.StdEncoding.DecodeString(target)
			http.Redirect(w, r, getBaseUrl(r)+string(targetDec), 302)
		} else {
			http.Redirect(w, r, getBaseUrl(r)+"dashboard", 302)
		}
	}
}

func logoutGetEndpoint(w http.ResponseWriter, r *http.Request) {
	appUsers.UserLogout(w, r)
	http.Redirect(w, r, getBaseUrl(r), 302)
}

func forgotPassGetEndpoint(w http.ResponseWriter, r *http.Request) {
	CompileSinglePage(w, r, "passreset_request.html", nil)
}
