package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import "net/http"


type FirstSetupData struct {
	DbHost string
	DbUser string
	DbPass string
	DbName string
}

func firstSetupGetEndpoint(w http.ResponseWriter, r *http.Request) {
	if !appConfig.HasFirstSetup {
		r.ParseForm()
		done := r.FormValue("done") != ""
		if !done {
			fd := FirstSetupData{
				DbHost: appConfig.DbHost,
				DbUser: appConfig.DbUsername,
				DbPass: appConfig.DbPassword,
				DbName: appConfig.DbName,
			}
			CompileSinglePage(w, r, "first_setup.html", fd)
		} else {
			CompileSinglePage(w, r, "first_setup_done.html", nil)
			cfg := appConfig
			cfg.HasFirstSetup = true
			saveConfigData(cfg)
			appConfig = getConfigData()
		}
	} else {
		http.Error(w, "403 Forbidden", 403)
	}
}

func firstSetupPostEndpoint(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	dbHost := r.PostFormValue("dbhost")
	dbUser := r.PostFormValue("dbuser")
	dbPass := r.PostFormValue("dbpass")
	dbName := r.PostFormValue("dbname")
	dbCreate := r.PostFormValue("dbcreate") != ""
	cfg := appConfig
	cfg.DbHost = dbHost
	cfg.DbUsername = dbUser
	cfg.DbPassword = dbPass
	cfg.DbName = dbName
	// Save and reload config
	saveConfigData(cfg)
	appConfig = getConfigData()
	if dbCreate {
		if err := CreateBlankDatabase(); err != nil {
			log := newLog()
			log.Error(err)
			appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
			http.Redirect(w, r, "gsetup", 302)
			return
		}
	}
	http.Redirect(w, r, "gsetup?done=yes", 302)
}
