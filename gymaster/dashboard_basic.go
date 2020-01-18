package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"io/ioutil"
	"net/http"
	"strconv"
)

type DashboardHomeData struct {
	News        NewsFeed
	OnlineUsers gytypes.UserOnlineList
}

type DashboardNotificationData struct {
	UnreadCount         int
	UnreadNotifications []NotificationChildDetails
	ReadCount           int
	ReadNotifications   []NotificationChildDetails
	OnlineUsers         gytypes.UserOnlineList
}

type DashboardScoreboardsData struct {
	Count          int
	ScoreboardList []gytypes.ScoreboardListData
	OnlineUsers    gytypes.UserOnlineList
}

type DashboardViewScoreboardData struct {
	Scoreboard gytypes.ScoreboardData
}

type DashboardProfileData struct {
	CountryList CountryListName
	OnlineUsers gytypes.UserOnlineList
}

type DashboardSettingsData struct {
	SyntaxThemeList []SyntaxThemeInfo
	SyntaxTest      string
	OnlineUsers     gytypes.UserOnlineList
}

type SyntaxThemeInfo struct {
	Name      string `json:"name"`
	ThemeName string `json:"themeName"`
}

func getSyntaxThemeList() ([]SyntaxThemeInfo, error) {
	var sl []SyntaxThemeInfo
	b, err := ioutil.ReadFile(gylib.ConcatByProgramLibDir("./templates/syntaxthemes.json"))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &sl)
	if err != nil {
		return nil, err
	}
	return sl, nil
}

func getSyntaxTestContent() (string, error) {
	b, err := ioutil.ReadFile(gylib.ConcatByProgramLibDir("./templates/syntax_test.c"))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func dashboardHomeGetEndpoint(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.FormValue("timeOut") == "yes" {
		appUsers.AddFlashMessage(w, r, "Waktu anda telah habis!", FlashError)
		http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard", 302)
		return
	}
	dhd := DashboardHomeData{
		News:        fetchNewsFeed(),
		OnlineUsers: appUsers.GetOnlineUsers(300),
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_home.html",
		"home", dhd, "")
}

func dashboardScoreboardsGetEndpoint(w http.ResponseWriter, r *http.Request) {
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard", 302)
		}
	}()
	list, err := appScoreboard.GetScoreboardListByUser(appUsers.GetLoggedUserInfo(r))
	if err != nil {
		return
	}
	dsd := DashboardScoreboardsData{
		Count:          len(list),
		ScoreboardList: list,
		OnlineUsers:    appUsers.GetOnlineUsers(300),
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_scoreboards.html",
		"scoreboard", dsd, "")
}

func dashboardViewScoreboardGetEndpoint(w http.ResponseWriter, r *http.Request) {
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/scoreboard", 302)
		}
	}()
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	sb, err := appScoreboard.GetScoreboardByUser(appUsers.GetLoggedUserInfo(r), id)
	if err != nil {
		return
	}
	dvsd := DashboardViewScoreboardData{
		Scoreboard: sb,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_scoreboardview.html",
		"scoreboard", dvsd, "")
}

func dashboardNotificationsEndpoint(w http.ResponseWriter, r *http.Request) {
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard", 302)
		}
	}()
	uid := appUsers.GetLoggedUserId(r)
	notifications, err := appNotifications.GetNotifications(uid)
	if err != nil {
		return
	}
	var unread, hasRead []NotificationChildDetails
	for _, nt := range notifications {
		if nt.HasRead {
			hasRead = append(hasRead, nt)
		} else {
			unread = append(unread, nt)
		}
	}
	dnd := DashboardNotificationData{
		UnreadCount:         len(unread),
		UnreadNotifications: unread,
		ReadCount:           len(hasRead),
		ReadNotifications:   hasRead,
		OnlineUsers:         appUsers.GetOnlineUsers(300),
	}
	// Read all unread notifications
	err = appNotifications.MarkNotificationsAllRead(uid)
	if err != nil {
		return
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_notifications.html",
		"notifications", dnd, "")
}

func dashboardProfileGetEndpoint(w http.ResponseWriter, r *http.Request) {
	cl, _ := GetCountryListName()
	dpd := DashboardProfileData{
		CountryList: cl,
		OnlineUsers: appUsers.GetOnlineUsers(300),
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_profileedit.html",
		"profile", dpd, "")
}

func dashboardProfilePostEndpoint(w http.ResponseWriter, r *http.Request) {
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/profile", 302)
		}
	}()
	user := appUsers.GetLoggedUserInfo(r)
	r.ParseForm()
	displayName := r.PostFormValue("display_name")
	gender := r.PostFormValue("gender")
	address := r.PostFormValue("address")
	institution := r.PostFormValue("institution")
	country := r.PostFormValue("country")
	avatarOption := r.PostFormValue("avatarOption")
	if country == "" {
		country = user.CountryId
	}
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	udm := NewUserDbModel(db)
	uid := user.Id
	ui, err := udm.GetUserById(uid)
	if err != nil {
		return
	}
	// Modify Profile here
	ui.DisplayName = displayName
	ui.Gender = gender
	ui.Address = address
	ui.Institution = institution
	ui.CountryId = country
	if avatarOption != "" {
		dropUserAvatarCache(ui.Avatar)
		avatarStr := getPersonalizedUserAvatar(ui.Id, avatarOption)
		ui.Avatar = avatarStr
	}
	err = udm.ModifyUserAccount(uid, ui)
	if err != nil {
		return
	}
	appUsers.RefreshUser(user.Id)
	appUsers.AddFlashMessage(w, r, "Success change your profile!", FlashSuccess)
	http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/profile", 302)
}

func dashboardSettingsGetEndpoint(w http.ResponseWriter, r *http.Request) {
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard", 302)
		}
	}()
	sl, err := getSyntaxThemeList()
	if err != nil {
		return
	}
	test, err := getSyntaxTestContent()
	if err != nil {
		return
	}
	dsd := DashboardSettingsData{
		SyntaxThemeList: sl,
		SyntaxTest:      test,
		OnlineUsers:     appUsers.GetOnlineUsers(300),
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_settings.html",
		"settings", dsd, "")
}

func dashboardSettingsPostEndpoint(w http.ResponseWriter, r *http.Request) {
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/settings", 302)
		}
	}()
	user := appUsers.GetLoggedUserInfo(r)
	r.ParseForm()
	email := r.PostFormValue("email")
	pass1 := r.PostFormValue("pass1")
	pass2 := r.PostFormValue("pass2")
	syntaxTheme := r.PostFormValue("syntax_theme")
	passHash := ""
	passSalt := gylib.GenerateRandomSalt() // Assumed it's regenerated
	if (pass1 == "") && (pass2 == "") {
		// Preserve old password if supposed to not changed
		passHash = user.Password
		passSalt = user.Salt
	} else {
		if pass1 != pass2 {
			err = errors.New("password yang akan diganti harus sama")
			return
		} else {
			passHash = calculateSaltedHash(pass1, passSalt)
		}
	}
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	udm := NewUserDbModel(db)
	uid := user.Id
	ui, err := udm.GetUserById(uid)
	if err != nil {
		return
	}
	// Modify here
	ui.Email = email
	ui.Password = passHash
	ui.Salt = passSalt
	ui.SyntaxTheme = syntaxTheme
	err = udm.ModifyUserAccount(uid, ui)
	if err != nil {
		return
	}
	appUsers.RefreshUser(user.Id)
	appUsers.AddFlashMessage(w, r, "Success updating your account settings!", FlashSuccess)
	http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/settings", 302)
}
