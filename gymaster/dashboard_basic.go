package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"net/http"
)

type DashboardHomeData struct {
	News        NewsFeed
	OnlineUsers UserOnlineList
}

type DashboardProfileData struct {
	CountryList CountryListName
}

func dashboardHomeGetEndpoint(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.FormValue("timeOut") == "yes" {
		appUsers.AddFlashMessage(w, r, "Waktu anda telah habis!", FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard", 302)
		return
	}
	dhd := DashboardHomeData{
		News:        fetchNewsFeed(),
		OnlineUsers: appUsers.GetOnlineUsers(300),
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_home.html",
		"home", dhd, "")
}

func dashboardProfileGetEndpoint(w http.ResponseWriter, r *http.Request) {
	cl, _ := GetCountryListName("")
	dpd := DashboardProfileData{
		CountryList: cl,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_profile.html",
		"profile", dpd, "")
}

func dashboardProfilePostEndpoint(w http.ResponseWriter, r *http.Request) {
	log := newLog()
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
	udm, err := NewUserDbModel()
	if err != nil {
		log.Error(err)
		appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/settings", 302)
		return
	}
	defer udm.Close()
	uid := user.Id
	ui, err := udm.GetUserById(uid)
	if err != nil {
		log.Error(err)
		appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/settings", 302)
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
		log.Error(err)
		appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/settings", 302)
		return
	}
	appUsers.RefreshUser(user.Id)
	appUsers.AddFlashMessage(w, r, "Sukes mengupdate profil anda!", FlashInformation)
	http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/profile", 302)
}

func dashboardSettingsGetEndpoint(w http.ResponseWriter, r *http.Request) {
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_settings.html",
		"settings", nil, "")
}

func dashboardSettingsPostEndpoint(w http.ResponseWriter, r *http.Request) {
	log := newLog()
	user := appUsers.GetLoggedUserInfo(r)
	r.ParseForm()
	email := r.PostFormValue("email")
	pass1 := r.PostFormValue("pass1")
	pass2 := r.PostFormValue("pass2")
	passHash := ""
	passSalt := generateRandomSalt() // Assumed it's regenerated
	if (pass1 == "") && (pass2 == "") {
		// Preserve old password if supposed to not changed
		passHash = user.Password
		passSalt = user.Salt
	} else {
		if pass1 != pass2 {
			appUsers.AddFlashMessage(w, r, "Password yang akan diganti harus sama!", FlashError)
			http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/settings", 302)
			return
		} else {
			passHash = calculateSaltedHash(pass1, passSalt)
		}
	}
	udm, err := NewUserDbModel()
	if err != nil {
		log.Error(err)
		appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/settings", 302)
		return
	}
	defer udm.Close()
	uid := user.Id
	ui, err := udm.GetUserById(uid)
	if err != nil {
		log.Error(err)
		appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/settings", 302)
		return
	}
	// Modify here
	ui.Email = email
	ui.Password = passHash
	ui.Salt = passSalt
	err = udm.ModifyUserAccount(uid, ui)
	if err != nil {
		log.Error(err)
		appUsers.AddFlashMessage(w, r, err.Error(), FlashError)
		http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/settings", 302)
		return
	}
	appUsers.RefreshUser(user.Id)
	appUsers.AddFlashMessage(w, r, "Sukes mengupdate pengaturan akun anda!", FlashInformation)
	http.Redirect(w, r, getBaseUrlWithSlash(r)+"dashboard/settings", 302)
}