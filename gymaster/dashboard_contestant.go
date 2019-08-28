package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type DashboardProblemData struct {
	ProblemData
	RemainingTime int
}

type DashboardProblemSetData struct {
	ProblemSet
	RemainingTime int
}

type DashboardUserSubmissionsData struct {
	Count       int
	Submissions []SubmissionInfo
}

func dashboardContestsGetEndpoint(w http.ResponseWriter, r *http.Request) {
	cdm, err := NewContestDbModel()
	if err != nil {
		http.Error(w, "500 Internal Server Error: "+err.Error(), 500)
		return
	}
	defer cdm.Close()
	cl, err := cdm.GetContestListOfUserId(appUsers.GetLoggedUserId(r))
	if err != nil {
		log := newLog()
		log.Error(err)
		appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/contests", 302)
		return
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_contestgate.html",
		"contests", cl, "")
}

func dashboardProblemSetGetEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	// Check access
	ca, err := appContestAccess.GetAccessInfoOfUser(appUsers.GetLoggedUserId(r), id)
	if err != nil {
		log := newLog()
		log.Error(err)
		appUsers.AddFlashMessage(w, r, "Access denied: "+err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/contests", 302)
		return
	}
	err = appContestAccess.CheckAccessInfo(ca)
	if err != nil {
		log := newLog()
		log.Error(err)
		appUsers.AddFlashMessage(w, r, "Access denied: "+err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/contests", 302)
		return
	}
	cdm, err := NewContestDbModel()
	if err != nil {
		http.Error(w, "500 Internal Server Error: "+err.Error(), 500)
		return
	}
	defer cdm.Close()
	cd, err := cdm.GetContestDetails(id)
	if err != nil {
		log := newLog()
		log.Error(err)
		appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/contests", 302)
		return
	}
	qs, err := cdm.GetProblemSet(id)
	ps := ProblemSet{
		Contest:  cd,
		Problems: qs,
		Count:    len(qs),
	}
	dps := DashboardProblemSetData{
		ProblemSet:    ps,
		RemainingTime: ca.RemainTime,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_problemset.html",
		"contests", dps, cd.Title)
}

func dashboardProblemGetEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	// Check access
	ca, err := appContestAccess.GetAccessInfoOfUser(appUsers.GetLoggedUserId(r), id)
	if err != nil {
		log := newLog()
		log.Error(err)
		appUsers.AddFlashMessage(w, r, "Access denied: "+err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/contests", 302)
		return
	}
	err = appContestAccess.CheckAccessInfo(ca)
	if err != nil {
		log := newLog()
		log.Error(err)
		appUsers.AddFlashMessage(w, r, "Access denied: "+err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/contests", 302)
		return
	}
	cdm, err := NewContestDbModel()
	if err != nil {
		http.Error(w, "500 Internal Server Error: "+err.Error(), 500)
		return
	}
	defer cdm.Close()
	qd, err := cdm.GetProblemById(id)
	if err != nil {
		log := newLog()
		log.Error(err)
		appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/contests", 302)
		return
	}
	pd := DashboardProblemData{
		ProblemData:   qd,
		RemainingTime: ca.RemainTime,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_problemview.html",
		"contests", pd, qd.Name)
}

func dashboardUserSubmissionsGetEndpoint(w http.ResponseWriter, r *http.Request) {
	sdm, err := NewSubmissionDbModel()
	if err != nil {
		log := newLog()
		log.Error(err)
		appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/userSubmissions", 302)
		return
	}
	defer sdm.Close()
	subList, err := sdm.GetSubmissionList(appUsers.GetLoggedUserId(r), 0)
	if err != nil {
		log := newLog()
		log.Error(err)
		appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/userSubmissions", 302)
		return
	}
	sd := DashboardUserSubmissionsData{
		Count:       len(subList),
		Submissions: subList,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_submissions.html",
		"usersubmissions", sd, "")
}
