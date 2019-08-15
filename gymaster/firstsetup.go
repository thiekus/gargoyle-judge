package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

type FirstSetupData struct {
	DbDone bool
	DbHost string
	DbUser string
	DbPass string
	DbName string
	DbFile string
}

// In the case you was created DB schema successfully, but not user
var doneDbSetup = false

func firstSetupGetEndpoint(w http.ResponseWriter, r *http.Request) {
	if !appConfig.HasFirstSetup {
		r.ParseForm()
		done := r.FormValue("done") != ""
		if !done {
			fd := FirstSetupData{
				DbDone: doneDbSetup,
				DbHost: appConfig.DbHost,
				DbUser: appConfig.DbUsername,
				DbPass: appConfig.DbPassword,
				DbName: appConfig.DbName,
				DbFile: appConfig.DbFile,
			}
			CompileSinglePage(w, r, "first_setup.html", fd)
		} else {
			CompileSinglePage(w, r, "first_setup_done.html", nil)
			cfg := appConfig
			// Block this setup for avoid abuse in the future
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
	dbDriver := r.PostFormValue("dbdriver")
	dbHost := r.PostFormValue("dbhost")
	dbUser := r.PostFormValue("dbuser")
	dbPass := r.PostFormValue("dbpass")
	dbName := r.PostFormValue("dbname")
	dbFile := r.PostFormValue("dbfile")
	dbCreate := r.PostFormValue("dbcreate") != ""
	adminUser := r.PostFormValue("adminuser")
	adminEmail := r.PostFormValue("adminemail")
	adminPass1 := r.PostFormValue("adminpass1")
	adminPass2 := r.PostFormValue("adminpass2")
	adminCreate := r.PostFormValue("admincreate") != ""
	if !doneDbSetup {
		cfg := appConfig
		cfg.DbDriver = dbDriver
		cfg.DbHost = dbHost
		cfg.DbUsername = dbUser
		cfg.DbPassword = dbPass
		cfg.DbName = dbName
		cfg.DbFile = dbFile
		// Save and reload config
		saveConfigData(cfg)
	}
	appConfig = getConfigData()
	if dbCreate && !doneDbSetup {
		if err := CreateBlankDatabase(); err != nil {
			log := newLog()
			log.Error(err)
			appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
			http.Redirect(w, r, "gysetup", 302)
			return
		}
		doneDbSetup = true
	}
	if adminCreate {
		if adminUser == "" {
			appUsers.AddFlashMessage(w, r, "Username tidak boleh kosong!", FlashError)
			http.Redirect(w, r, "gysetup", 302)
			return
		}
		if adminEmail == "" {
			appUsers.AddFlashMessage(w, r, "Email tidak boleh kosong!", FlashError)
			http.Redirect(w, r, "gysetup", 302)
			return
		}
		if adminPass1 == "" {
			appUsers.AddFlashMessage(w, r, "Password tidak boleh kosong!", FlashError)
			http.Redirect(w, r, "gysetup", 302)
			return
		}
		if adminPass1 != adminPass2 {
			appUsers.AddFlashMessage(w, r, "Password konfirmasi tidak sama!", FlashError)
			http.Redirect(w, r, "gysetup", 302)
			return
		}
		db, err := OpenDatabase(false)
		if err != nil {
			log := newLog()
			log.Error(err)
			appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
			http.Redirect(w, r, "gysetup", 302)
			return
		}
		defer db.Close()
		query := `INSERT INTO gy_users (username, password, email, display_name, address, avatar, iguser, role, verified)
			VALUES (?, ?, ?, ?, "unknown", "m:username", "", 1, 1)`
		prep, err := db.Prepare(query)
		if err != nil {
			log := newLog()
			log.Error(err)
			appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
			http.Redirect(w, r, "gysetup", 302)
			return
		}
		defer prep.Close()
		passHash := fmt.Sprintf("%x", sha256.Sum256([]byte(adminPass1)))
		_, err = prep.Exec(adminUser, passHash, adminEmail, adminUser)
		if err != nil {
			log := newLog()
			log.Error(err)
			appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
			http.Redirect(w, r, "gysetup", 302)
			return
		}
	}
	http.Redirect(w, r, "gysetup?done=yes", 302)
}
