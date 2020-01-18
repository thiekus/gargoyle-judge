package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"io/ioutil"
	"net/http"
	"strconv"
)

type DashboardContestGateData struct {
	Count    int
	Contests []gytypes.ContestData
}

type DashboardProblemData struct {
	gytypes.ProblemData
	Problems      []gytypes.ProblemData
	ProblemCount  int
	AllowedLang   []gytypes.LanguageProgramData
	RemainingTime int
}

type DashboardProblemSetData struct {
	gytypes.ProblemSet
	RemainingTime int
}

type DashboardUserSubmissionsData struct {
	Count       int
	Submissions []gytypes.SubmissionData
}

type DashboardUserViewSubmissionData struct {
	Submission gytypes.SubmissionData
	TestResult []gytypes.TestResultData
	TestCount  int
}

func dashboardContestsGetEndpoint(w http.ResponseWriter, r *http.Request) {
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/contests", 302)
		}
	}()
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	cdm := NewContestDbModel(db)
	cl, err := cdm.GetContestListOfUserId(appUsers.GetLoggedUserId(r))
	if err != nil {
		return
	}
	cgd := DashboardContestGateData{
		Count:    len(cl),
		Contests: cl,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_contestgate.html",
		"contests", cgd, "")
}

func dashboardProblemSetGetEndpoint(w http.ResponseWriter, r *http.Request) {
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/contests", 302)
		}
	}()
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	// Check access
	ca, err := appContestAccess.GetAccessInfoOfUser(appUsers.GetLoggedUserId(r), id)
	if err != nil {
		return
	}
	err = appContestAccess.CheckAccessInfo(ca)
	if err != nil {
		return
	}
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	cdm := NewContestDbModel(db)
	cd, err := cdm.GetContestDetails(id)
	if err != nil {
		return
	}
	qs, err := cdm.GetProblemSet(id)
	if err != nil {
		return
	}
	ps := gytypes.ProblemSet{
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
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/contests", 302)
		}
	}()
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	cdm := NewContestDbModel(db)
	qd, err := cdm.GetProblemById(id)
	if err != nil {
		return
	}
	ps, err := cdm.GetProblemSet(qd.ContestId)
	if err != nil {
		return
	}
	// Check access
	ca, err := appContestAccess.GetAccessInfoOfUser(appUsers.GetLoggedUserId(r), qd.ContestId)
	if err != nil {
		return
	}
	err = appContestAccess.CheckAccessInfo(ca)
	if err != nil {
		return
	}
	lang, err := appLangPrograms.GetLanguageOfContest(qd.AllowedLang)
	if err != nil {
		return
	}
	pd := DashboardProblemData{
		ProblemData:   qd,
		Problems:      ps,
		ProblemCount:  len(ps),
		AllowedLang:   lang,
		RemainingTime: ca.RemainTime,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_problemview.html",
		"contests", pd, qd.Name)
}

func dashboardProblemPostEndpoint(w http.ResponseWriter, r *http.Request) {
	log := gylib.GetStdLog()
	var err error = nil
	errRedirect := gylib.GetBaseUrlWithSlash(r) + "dashboard/contests"
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, errRedirect, 302)
		}
	}()
	r.ParseMultipartForm(500 * 1024)
	inputType := r.PostFormValue("inputType")
	contestId, err := strconv.Atoi(r.PostFormValue("problemSetId"))
	if err != nil {
		return
	}
	userId := appUsers.GetLoggedUserId(r)
	problemId, err := strconv.Atoi(r.PostFormValue("problemId"))
	if err != nil {
		errRedirect = gylib.GetBaseUrlWithSlash(r) + "dashboard/problemSet/" + strconv.Itoa(contestId)
		return
	}
	langId, err := strconv.Atoi(r.PostFormValue("lang"))
	if err != nil {
		errRedirect = gylib.GetBaseUrlWithSlash(r) + "dashboard/problem/" + strconv.Itoa(problemId)
		return
	}
	var code string
	switch inputType {
	case "file":
		// TODO: Submission via file upload doesn't work with QuickForm yet
		f, _, err := r.FormFile("sourceCodeFile")
		if err != nil {
			return
		}
		defer f.Close()
		b, err := ioutil.ReadAll(f)
		if err != nil {
			return
		}
		code = string(b)
	case "text":
		code = r.PostFormValue("sourceCodeText")
	default:
		err = errors.New("invalid type name")
		errRedirect = gylib.GetBaseUrlWithSlash(r) + "dashboard/problem/" + strconv.Itoa(problemId)
		return
	}
	sub, err := NewSubmissionProcessor(&appSlaves, problemId, userId, langId, code)
	if err != nil {
		errRedirect = gylib.GetBaseUrlWithSlash(r) + "dashboard/problem/" + strconv.Itoa(problemId)
		return
	}
	if err = sub.DoProcess(); err == nil {
		appUsers.AddFlashMessage(w, r, "Berhasil memasukan submisi!", FlashSuccess)
		http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/problemSet/"+strconv.Itoa(contestId), 302)
	} else {
		errRedirect = gylib.GetBaseUrlWithSlash(r) + "dashboard/problem/" + strconv.Itoa(problemId)
	}
}

func dashboardUserSubmissionsGetEndpoint(w http.ResponseWriter, r *http.Request) {
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/userSubmissions", 302)
		}
	}()
	r.ParseForm()
	problemId, err := strconv.Atoi(r.FormValue("problem"))
	if err != nil {
		problemId = 0
	}
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	sdm := NewSubmissionDbModel(db)
	subList, err := sdm.GetSubmissionList(appUsers.GetLoggedUserId(r), problemId)
	if err != nil {
		return
	}
	sd := DashboardUserSubmissionsData{
		Count:       len(subList),
		Submissions: subList,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_submissions.html",
		"usersubmissions", sd, "")
}

func dashboardUserViewSubmissionGetEndpoint(w http.ResponseWriter, r *http.Request) {
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/userSubmissions", 302)
		}
	}()
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	sdm := NewSubmissionDbModel(db)
	sub, err := sdm.GetSubmission(id)
	if err != nil {
		return
	}
	// Prevent someone cheating others
	uid := appUsers.GetLoggedUserId(r)
	if sub.UserId != uid {
		err = errors.New("illegal operation")
		return
	}
	// Get test results
	tests, err := sdm.GetTestResultOfSubmission(sub.Id)
	if sub.UserId != uid {
		return
	}
	sd := DashboardUserViewSubmissionData{
		Submission: sub,
		TestResult: tests,
		TestCount:  len(tests),
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_submissionview.html",
		"usersubmissions", sd, "")
}
