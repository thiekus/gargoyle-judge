package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"net/http"
)

type DashboardManageContestData struct {
	Count    int
	Contests []gytypes.ContestData
}

type DashboardContestAddData struct {
	Languages gytypes.LanguageProgramMap
}

func dashboardManageContestsGetEndpoint(w http.ResponseWriter, r *http.Request) {
	ui := appUsers.GetLoggedUserInfo(r)
	if !ui.Roles.Jury {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard", 302)
		}
	}()
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	cdm := NewContestDbModel(db)
	cl, err := cdm.GetContestList()
	if err != nil {
		return
	}
	cgd := DashboardManageContestData{
		Count:    len(cl),
		Contests: cl,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_managecontests.html",
		"managecontests", cgd, "")
}

func dashboardContestAddGetEndpoint(w http.ResponseWriter, r *http.Request) {
	ui := appUsers.GetLoggedUserInfo(r)
	if !ui.Roles.Jury {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/manageContests", 302)
		}
	}()
	langs, err := appLangPrograms.GetLanguageMap()
	if err != nil {
		return
	}
	cad := DashboardContestAddData{
		Languages: langs,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_contestadd.html",
		"managecontests", cad, "")
}
