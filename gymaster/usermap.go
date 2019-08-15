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

type LoggedUsers struct {
	sessionStore *sessions.CookieStore
	umap         UsersMap
	tmap         UsersTokenMap
}

func MakeLoggedUsersMap() LoggedUsers {
	luser := LoggedUsers{
		sessionStore: sessions.NewCookieStore([]byte(appConfig.SessionKey)),
		umap:         make(UsersMap),
		tmap:         make(UsersTokenMap),
	}
	return luser
}

func (lu *LoggedUsers) GetUserById(id int) *UserInfo {
	if info, ok := lu.umap[id]; ok {
		return &info
	} else {
		return nil
	}
}

func (lu *LoggedUsers) GetUserByToken(token string) *UserInfo {
	if uid, ok := lu.tmap[token]; ok {
		return lu.GetUserById(uid)
	} else {
		return nil
	}
}

func (lu *LoggedUsers) GetUserSession(r *http.Request) (*sessions.Session, error) {
	return lu.sessionStore.Get(r, UserSessionName)
}

func (lu *LoggedUsers) GetMessageSession(r *http.Request) (*sessions.Session, error) {
	return lu.sessionStore.Get(r, FlashSessionName)
}

func (lu *LoggedUsers) GetLoggedUserToken(r *http.Request) string {
	if sess, err := lu.GetUserSession(r); err == nil {
		token := sess.Values["UserToken"]
		if token != nil {
			return token.(string)
		}
	}
	return ""
}

func (lu *LoggedUsers) GetLoggedUserInfo(r *http.Request) *UserInfo {
	if token := lu.GetLoggedUserToken(r); token != "" {
		return appUsers.GetUserByToken(token)
	}
	return nil
}

func (lu *LoggedUsers) SetLoggedUserInfo(w http.ResponseWriter, r *http.Request, token string) {
	if sess, err := lu.GetUserSession(r); err == nil {
		sess.Values["UserToken"] = token
		sess.Save(r, w)
	}
}

func (lu *LoggedUsers) RefreshUser(userId int) error {
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
	lu.umap[userId] = ui
	return err
}

func (lu *LoggedUsers) UserLogin(w http.ResponseWriter, r *http.Request, username string, password string) error {
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
		if _, tokenExists := lu.tmap[token]; !tokenExists {
			break
		}
	}
	ui.Token = token
	lu.umap[uid] = ui
	lu.tmap[token] = uid
	// Set session
	lu.SetLoggedUserInfo(w, r, token)
	log.Printf("User %s logged in with token %s", username, token)
	return nil
}

func (lu *LoggedUsers) UserRemoveFromList(token string) {
	delete(lu.umap, lu.tmap[token])
	delete(lu.tmap, token)
}

func (lu *LoggedUsers) UserLogout(w http.ResponseWriter, r *http.Request) {
	token := lu.GetLoggedUserToken(r)
	lu.UserRemoveFromList(token)
	lu.SetLoggedUserInfo(w, r, "")
}

func (lu *LoggedUsers) GetFlashMessage(w http.ResponseWriter, r *http.Request) (string, string) {
	if sess, err := lu.GetMessageSession(r); err == nil {
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

func (lu *LoggedUsers) AddFlashMessage(w http.ResponseWriter, r *http.Request, message string, ftype int) {
	if sess, err := lu.GetMessageSession(r); err == nil {
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
