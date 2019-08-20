package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"fmt"
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

func (uc *UserController) GetLoggedUserId(r *http.Request) int {
	if ui := uc.GetLoggedUserInfo(r); ui != nil {
		return ui.Id
	} else {
		return 0
	}
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
	udm, err := NewUserDbModel()
	if err != nil {
		log.Errorf("[uid:%d] Refresh error: %s", userId, err)
		return err
	}
	defer udm.Close()
	ui, err := udm.GetUserById(userId)
	if err != nil {
		log.Errorf("[uid:%d] Refresh error: %s", userId, err)
		return err
	}
	uc.umap[userId] = ui
	return err
}

func (uc *UserController) UserLogin(username string, password string) (string, error) {
	log := newLog()
	log.Printf("User %s trying to login...", username)
	udm, err := NewUserDbModel()
	if err != nil {
		log.Errorf("[%s] Login error: %s", username, err)
		return "", err
	}
	defer udm.Close()
	ui, err := udm.GetUserByLogin(username, password)
	if err != nil {
		log.Errorf("[%s] Login error: %s", username, err)
		return "", err
	}
	var token string
	// Avoid token collisions
	for {
		token = generateRandomToken()
		if _, tokenExists := uc.tmap[token]; !tokenExists {
			break
		}
	}
	uid := ui.Id
	ui.Token = token
	uc.umap[uid] = ui
	uc.tmap[token] = uid
	log.Printf("User %s logged in with token %s", username, token)
	return token, nil
}

func (uc *UserController) UserLoginFromWebsite(w http.ResponseWriter, r *http.Request, username string, password string) error {
	token, err := uc.UserLogin(username, password)
	if err != nil {
		return err
	}
	// Set session
	uc.SetLoggedUserInfo(w, r, token)
	return nil
}

func (uc *UserController) UserRemoveFromList(token string) {
	delete(uc.umap, uc.tmap[token])
	delete(uc.tmap, token)
}

func (uc *UserController) UserLogout(token string) {
	uc.UserRemoveFromList(token)
}

func (uc *UserController) UserLogoutFromWebsite(w http.ResponseWriter, r *http.Request) {
	token := uc.GetLoggedUserToken(r)
	uc.UserLogout(token)
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
