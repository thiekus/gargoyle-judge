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
	"net/http"
	"strconv"
)

type DashboardManageUsersData struct {
	UserCount int
	Users     []gytypes.UserInfo
}

type DashboardUserAddData struct {
	Roles       []gytypes.UserRoleAccess
	CountryList CountryListName
}

type DashboardUserEditData struct {
	Roles       []gytypes.UserRoleAccess
	UserInfo    gytypes.UserInfo
	CountryList CountryListName
}

func dashboardManageUsersGetEndpoint(w http.ResponseWriter, r *http.Request) {
	ui := appUsers.GetLoggedUserInfo(r)
	if !ui.Roles.SysAdmin {
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
	udm := NewUserDbModel(db)
	ul, err := udm.GetUserList()
	if err != nil {
		return
	}
	udd := DashboardManageUsersData{
		UserCount: len(ul),
		Users:     ul,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_manageusers.html",
		"manageusers", udd, "")
}

func dashboardUserAddGetEndpoint(w http.ResponseWriter, r *http.Request) {
	ui := appUsers.GetLoggedUserInfo(r)
	if !ui.Roles.SysAdmin {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/manageUsers", 302)
		}
	}()
	cl, err := GetCountryListName()
	if err != nil {
		return
	}
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	udm := NewUserDbModel(db)
	roles, err := udm.GetAccessRoleList()
	if err != nil {
		return
	}
	uad := DashboardUserAddData{
		Roles:       roles,
		CountryList: cl,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_useradd.html",
		"manageusers", uad, "")
}

func dashboardUserAddPostEndpoint(w http.ResponseWriter, r *http.Request) {
	ui2 := appUsers.GetLoggedUserInfo(r)
	if !ui2.Roles.SysAdmin {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/userAdd", 302)
		}
	}()
	r.ParseForm()
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	accessRoleStr := r.PostFormValue("access_role")
	accessRole, _ := strconv.Atoi(accessRoleStr)
	displayName := r.PostFormValue("display_name")
	gender := r.PostFormValue("gender")
	address := r.PostFormValue("address")
	institution := r.PostFormValue("institution")
	country := r.PostFormValue("country")
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	udm := NewUserDbModel(db)
	ui := gytypes.UserInfo{
		Email:       email,
		DisplayName: displayName,
		Gender:      gender,
		Address:     address,
		Institution: institution,
		CountryId:   country,
	}
	err = udm.CreateUserAccount(username, password, accessRole, ui)
	if err != nil {
		return
	}
	appUsers.AddFlashMessage(w, r, "Success adding new user!", FlashSuccess)
	http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/manageUsers", 302)
}

func dashboardUserEditGetEndpoint(w http.ResponseWriter, r *http.Request) {
	ui2 := appUsers.GetLoggedUserInfo(r)
	if !ui2.Roles.SysAdmin {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/manageUsers", 302)
		}
	}()
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	cl, err := GetCountryListName()
	if err != nil {
		return
	}
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	udm := NewUserDbModel(db)
	roles, err := udm.GetAccessRoleList()
	if err != nil {
		return
	}
	ui, err := udm.GetUserById(id)
	if err != nil {
		return
	}
	ued := DashboardUserEditData{
		Roles:       roles,
		UserInfo:    ui,
		CountryList: cl,
	}
	CompileDashboardPage(w, r, "dashboard_base.html", "dashboard_useredit.html",
		"manageusers", ued, "")
}

func dashboardUserEditPostEndpoint(w http.ResponseWriter, r *http.Request) {
	ui2 := appUsers.GetLoggedUserInfo(r)
	if !ui2.Roles.SysAdmin {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/userEdit", 302)
		}
	}()
	r.ParseForm()
	idStr := r.PostFormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		err = errors.New("invalid user id")
	}
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	accessRoleStr := r.PostFormValue("access_role")
	accessRole, _ := strconv.Atoi(accessRoleStr)
	displayName := r.PostFormValue("display_name")
	gender := r.PostFormValue("gender")
	address := r.PostFormValue("address")
	institution := r.PostFormValue("institution")
	country := r.PostFormValue("country")
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	udm := NewUserDbModel(db)
	// Get current configuration
	ui, err := udm.GetUserById(id)
	if err != nil {
		return
	}
	ui.Username = username
	ui.Email = email
	// Change password only if desired
	if password != "" {
		salt := gylib.GenerateRandomSalt()
		hash := calculateSaltedHash(password, salt)
		ui.Password = hash
		ui.Salt = salt
	}
	if accessRole != 0 {
		ui.RoleId = accessRole
	}
	ui.DisplayName = displayName
	ui.Gender = gender
	ui.Address = address
	ui.Institution = institution
	ui.CountryId = country
	err = udm.ModifyUserAccount(id, ui)
	if err != nil {
		return
	}
	appUsers.RefreshUser(id)
	appUsers.AddFlashMessage(w, r, "Success updating account settings!", FlashSuccess)
	http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/manageUsers", 302)
}

func dashboardUserDeleteGetEndpoint(w http.ResponseWriter, r *http.Request) {
	ui := appUsers.GetLoggedUserInfo(r)
	if !ui.Roles.SysAdmin {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	log := gylib.GetStdLog()
	var err error = nil
	defer func() {
		if err != nil {
			log.Error(err)
			appUsers.AddFlashMessage(w, r, "Error: "+err.Error(), FlashError)
			http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/manageUsers", 302)
		}
	}()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return
	}
	db, err := OpenDatabase()
	if err != nil {
		return
	}
	defer db.Close()
	udm := NewUserDbModel(db)
	err = udm.DeleteUserById(id)
	if err != nil {
		return
	}
	appUsers.AddFlashMessage(w, r, "Success deleting account!", FlashSuccess)
	http.Redirect(w, r, gylib.GetBaseUrlWithSlash(r)+"dashboard/manageUsers", 302)
}
