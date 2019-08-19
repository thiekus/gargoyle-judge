package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"strings"
)

type UsersMap map[int]UserInfo
type UsersTokenMap map[string]int

const UserSessionName = "GargoyleUserSession"
const FlashSessionName = "GargoyleFlashMessage"

const (
	FlashInformation = 0
	FlashWarning     = 1
	FlashError       = 2
)

type UserController struct {
	sessionStore *sessions.CookieStore
	umap         UsersMap
	tmap         UsersTokenMap
}

func MakeUserController() UserController {
	luser := UserController{
		sessionStore: sessions.NewCookieStore([]byte(appConfig.SessionKey)),
		umap:         make(UsersMap),
		tmap:         make(UsersTokenMap),
	}
	return luser
}

func (uc *UserController) GetUserById(id int) *UserInfo {
	if info, ok := uc.umap[id]; ok {
		return &info
	} else {
		return nil
	}
}

func (uc *UserController) GetUserByToken(token string) *UserInfo {
	if uid, ok := uc.tmap[token]; ok {
		return uc.GetUserById(uid)
	} else {
		return nil
	}
}

func (uc *UserController) GetUserSession(r *http.Request) (*sessions.Session, error) {
	return uc.sessionStore.Get(r, UserSessionName)
}

func (uc *UserController) GetMessageSession(r *http.Request) (*sessions.Session, error) {
	return uc.sessionStore.Get(r, FlashSessionName)
}

func (uc *UserController) GetLoggedUserToken(r *http.Request) string {
	if sess, err := uc.GetUserSession(r); err == nil {
		token := sess.Values["UserToken"]
		if token != nil {
			return token.(string)
		}
	}
	return ""
}

func (uc *UserController) GetLoggedUserInfo(r *http.Request) *UserInfo {
	if token := uc.GetLoggedUserToken(r); token != "" {
		return appUsers.GetUserByToken(token)
	}
	return nil
}

func (uc *UserController) SetLoggedUserInfo(w http.ResponseWriter, r *http.Request, token string) {
	if sess, err := uc.GetUserSession(r); err == nil {
		sess.Values["UserToken"] = token
		sess.Save(r, w)
	}
}

func (uc *UserController) RefreshUser(userId int) error {
	log := newLog()
	log.Printf("Refreshing user id no %d", userId)
	db, err := OpenDatabase(false)
	if err != nil {
		return err
	}
	defer db.Close()
	stmt, err := db.Prepare(
		`SELECT id, username, password, email, iguser, display_name, address, avatar, role
        FROM %TABLEPREFIX%users WHERE id = ?`)
	if err != nil {
		log.Errorf("[uid:%s] Refresh error: %s", userId, err)
		return err
	}
	defer stmt.Close()
	var ui UserInfo
	var uid int
	var roleId int
	err = stmt.QueryRow(userId).Scan(
		&uid,
		&ui.Username,
		&ui.Password,
		&ui.Email,
		&ui.IgUsername,
		&ui.DisplayName,
		&ui.Address,
		&ui.Avatar,
		&roleId,
	)
	if err != nil {
		log.Errorf("[uid:%s] Refresh error: %s", userId, err)
		return errors.New("username or password either invalid or not exists")
	}
	ui.Id = uid
	st2, err := db.Prepare("SELECT rolename, access_user, access_jury, access_root FROM %TABLEPREFIX%roles WHERE id = ?")
	if err != nil {
		log.Errorf("[uid:%s] Refresh error: %s", userId, err)
		return err
	}
	defer st2.Close()
	var acsUser int
	var acsJury int
	var acsRoot int
	err = st2.QueryRow(roleId).Scan(
		&ui.Roles.RoleName,
		&acsUser,
		&acsJury,
		&acsRoot,
	)
	if err != nil {
		log.Errorf("[uid:%s] Refresh error: %s", userId, err)
		return errors.New("invalid role id")
	}
	ui.Roles.Contestant = acsUser > 0
	ui.Roles.Operator = acsJury > 0
	ui.Roles.SysAdmin = acsRoot > 0
	uc.umap[userId] = ui
	return err
}

func (uc *UserController) UserLogin(w http.ResponseWriter, r *http.Request, username string, password string) error {
	log := newLog()
	log.Printf("User %s trying to login...", username)
	db, err := OpenDatabase(false)
	if err != nil {
		return err
	}
	defer db.Close()
	stmt, err := db.Prepare(
		`SELECT id, username, password, email, iguser, display_name, address, avatar, role
        FROM %TABLEPREFIX%users WHERE username = ? AND password = ?`)
	if err != nil {
		log.Errorf("[%s] Login error: %s", username, err)
		return err
	}
	defer stmt.Close()
	passHash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	var ui UserInfo
	var uid int
	var roleId int
	err = stmt.QueryRow(username, passHash).Scan(
		&uid,
		&ui.Username,
		&ui.Password,
		&ui.Email,
		&ui.IgUsername,
		&ui.DisplayName,
		&ui.Address,
		&ui.Avatar,
		&roleId,
	)
	if err != nil {
		log.Errorf("[%s] Login error: %s", username, err)
		return errors.New("username or password either invalid or not exists")
	}
	ui.Id = uid
	st2, err := db.Prepare("SELECT rolename, access_user, access_jury, access_root FROM %TABLEPREFIX%roles WHERE id = ?")
	if err != nil {
		log.Errorf("[%s] Login error: %s", username, err)
		return err
	}
	defer st2.Close()
	var acsUser int
	var acsJury int
	var acsRoot int
	err = st2.QueryRow(roleId).Scan(
		&ui.Roles.RoleName,
		&acsUser,
		&acsJury,
		&acsRoot,
	)
	if err != nil {
		log.Errorf("[%s] Login error: %s", username, err)
		return errors.New("invalid role id")
	}
	ui.Roles.Contestant = acsUser > 0
	ui.Roles.Operator = acsJury > 0
	ui.Roles.SysAdmin = acsRoot > 0
	var token string
	// Avoid token collisions
	for {
		token = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%x", securecookie.GenerateRandomKey(32)))))
		if _, tokenExists := uc.tmap[token]; !tokenExists {
			break
		}
	}
	ui.Token = token
	uc.umap[uid] = ui
	uc.tmap[token] = uid
	// Set session
	uc.SetLoggedUserInfo(w, r, token)
	log.Printf("User %s logged in with token %s", username, token)
	return nil
}

func (uc *UserController) UserRemoveFromList(token string) {
	delete(uc.umap, uc.tmap[token])
	delete(uc.tmap, token)
}

func (uc *UserController) UserLogout(w http.ResponseWriter, r *http.Request) {
	token := uc.GetLoggedUserToken(r)
	uc.UserRemoveFromList(token)
	uc.SetLoggedUserInfo(w, r, "")
}

func (uc *UserController) GetFlashMessage(w http.ResponseWriter, r *http.Request) (string, string) {
	if sess, err := uc.GetMessageSession(r); err == nil {
		if flash := sess.Flashes(); len(flash) > 0 {
			flashArg := strings.SplitN(flash[len(flash)-1].(string), ";", 2)
			flashType := flashArg[0]
			flashMsg := flashArg[1]
			sess.Save(r, w)
			return flashMsg, flashType
		}
	}
	return "", ""
}

func (uc *UserController) AddFlashMessage(w http.ResponseWriter, r *http.Request, message string, ftype int) {
	if sess, err := uc.GetMessageSession(r); err == nil {
		flashType := "info"
		switch ftype {
		case FlashInformation:
			flashType = "info"
		case FlashWarning:
			flashType = "warning"
		case FlashError:
			flashType = "error"
		}
		flashData := fmt.Sprintf("%s;%s", flashType, message)
		sess.AddFlash(flashData)
		sess.Save(r, w)
	}
}
