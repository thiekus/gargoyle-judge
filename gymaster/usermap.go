package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/sessions"
	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"net/http"
	"strings"
	"sync"
	"time"
)

//type UsersMap map[int]UserInfo
type UsersMap struct {
	sync.Map
}

//type UsersTokenMap map[string]int
type UsersTokenMap struct {
	sync.Map
}

const UserSessionName = "GargoyleUserSession"
const FlashSessionName = "GargoyleFlashMessage"

const (
	FlashInformation = 0
	FlashWarning     = 1
	FlashError       = 2
	FlashSuccess     = 3
)

type UserController struct {
	sessionStore *sessions.CookieStore
	umap         UsersMap
	tmap         UsersTokenMap
}

func MakeUserController() UserController {
	luser := UserController{
		sessionStore: sessions.NewCookieStore([]byte(appConfig.SessionKey)),
		umap:         UsersMap{},
		tmap:         UsersTokenMap{},
	}
	// Prevent users kicked out after short maintenance
	if db, err := OpenDatabase(); err == nil {
		defer db.Close()
		udm := NewUserDbModel(db)
		if tl, err := udm.GetTokenList(); err == nil {
			storedCount := 0
			for _, v := range tl {
				if ui, err := udm.GetUserById(v.UserId); err == nil {
					luser.storeUserMap(v.UserId, ui)
					luser.storeTokenMap(v.Token, v.UserId)
					storedCount++
				}
			}
			if storedCount > 0 {
				gylib.GetStdLog().Printf("%d user token retrieved", storedCount)
			}
		}
	}
	return luser
}

func (uc *UserController) deleteUserMap(key int) {
	uc.umap.Delete(key)
}

func (uc *UserController) loadUserMap(key int) (gytypes.UserInfo, bool) {
	ui, exists := uc.umap.Load(key)
	uiResult := gytypes.UserInfo{}
	if ui != nil {
		uiResult = ui.(gytypes.UserInfo)
	}
	return uiResult, exists
}

func (uc *UserController) rangeUserMap(f func(key int, value gytypes.UserInfo) bool) {
	uc.umap.Range(func(key, value interface{}) bool {
		return f(key.(int), value.(gytypes.UserInfo))
	})
}

func (uc *UserController) storeUserMap(key int, ui gytypes.UserInfo) {
	uc.umap.Store(key, ui)
}

func (uc *UserController) deleteTokenMap(key string) {
	uc.tmap.Delete(key)
}

func (uc *UserController) loadTokenMap(key string) (int, bool) {
	ti, exists := uc.tmap.Load(key)
	tiResult := 0
	if ti != nil {
		tiResult = ti.(int)
	}
	return tiResult, exists
}

func (uc *UserController) rangeTokenMap(f func(key string, value int) bool) {
	uc.tmap.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(int))
	})
}

func (uc *UserController) storeTokenMap(key string, val int) {
	uc.tmap.Store(key, val)
}

func (uc *UserController) CleanCookies(w http.ResponseWriter, r *http.Request) {
	// GargoyleUserSession
	gus := http.Cookie{
		Name:     UserSessionName,
		Value:    "",
		Domain:   r.URL.Hostname(),
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, &gus)
	// GargoyleFlashMessage
	gfm := http.Cookie{
		Name:     FlashSessionName,
		Value:    "",
		Domain:   r.URL.Hostname(),
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, &gfm)
}

func (uc *UserController) GetUserById(id int) *gytypes.UserInfo {
	if info, ok := uc.loadUserMap(id); ok {
		return &info
	} else {
		return nil
	}
}

func (uc *UserController) GetUserByToken(token string) *gytypes.UserInfo {
	if uid, ok := uc.loadTokenMap(token); ok {
		ui := uc.GetUserById(uid)
		// Assume is real owner who access the account
		if ui != nil {
			userInfo, _ := uc.loadUserMap(uid)
			userInfo.RefreshLastAccess()
			uc.storeUserMap(uid, userInfo)
			ui = uc.GetUserById(uid)
		}
		return ui
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

func (uc *UserController) GetLoggedUserInfo(r *http.Request) *gytypes.UserInfo {
	if token := uc.GetLoggedUserToken(r); token != "" {
		return appUsers.GetUserByToken(token)
	}
	return nil
}

func (uc *UserController) GetOnlineUsers(maxLastTime int) gytypes.UserOnlineList {
	var onlineList []gytypes.UserOnline
	timeNow := time.Now().Unix()
	uc.rangeUserMap(func(uid int, ui gytypes.UserInfo) bool {
		timeDiff := timeNow - ui.LastAccess.Unix()
		if (maxLastTime == 0) || (timeDiff < int64(maxLastTime)) {
			online := gytypes.UserOnline{
				Id:           uid,
				Username:     ui.Username,
				DisplayName:  ui.DisplayName,
				Institution:  ui.Institution,
				Avatar:       ui.Avatar,
				LastTimeDiff: timeDiff,
				TimeStatus:   humanize.Time(ui.LastAccess),
			}
			onlineList = append(onlineList, online)
		}
		return true
	})
	uol := gytypes.UserOnlineList{
		Count: len(onlineList),
		Users: onlineList,
	}
	return uol
}

func (uc *UserController) SetLoggedUserInfo(w http.ResponseWriter, r *http.Request, token string) {
	if sess, err := uc.GetUserSession(r); err == nil {
		sess.Values["UserToken"] = token
		sess.Save(r, w)
	}
}

func (uc *UserController) RefreshUser(userId int) error {
	log := gylib.GetStdLog()
	log.Printf("Refreshing user id no %d", userId)
	db, err := OpenDatabase()
	if err != nil {
		log.Errorf("[uid:%d] Refresh error: %s", userId, err)
		return err
	}
	defer db.Close()
	udm := NewUserDbModel(db)
	ui, err := udm.GetUserById(userId)
	if err != nil {
		log.Errorf("[uid:%d] Refresh error: %s", userId, err)
		return err
	}
	uc.storeUserMap(userId, ui)
	return err
}

func (uc *UserController) UserLogin(username string, password string) (string, error) {
	log := gylib.GetStdLog()
	log.Printf("User %s trying to login...", username)
	db, err := OpenDatabase()
	if err != nil {
		log.Errorf("[%s] UserLogin error: %s", username, err)
		return "", err
	}
	defer db.Close()
	udm := NewUserDbModel(db)
	ui, err := udm.GetUserByLogin(username, password)
	if err != nil {
		log.Errorf("[%s] UserLogin error: %s", username, err)
		return "", err
	}
	// Check token map if this user has logged in before, kick out :p
	uc.rangeTokenMap(func(tk string, tv int) bool {
		if tv == ui.Id {
			uc.UserRemoveFromList(tk)
		}
		return true
	})
	var token string
	// Avoid token collisions
	for {
		token = gylib.GenerateRandomToken()
		if _, tokenExists := uc.loadTokenMap(token); !tokenExists {
			break
		}
	}
	uid := ui.Id
	// Insert into database for persistent token
	if err := udm.DeleteTokenByUserId(uid); err != nil {
		log.Errorf("[%s] UserLogin error: %s", username, err)
		return "", err
	}
	if err := udm.InsertToken(token, uid); err != nil {
		log.Errorf("[%s] UserLogin error: %s", username, err)
		return "", err
	}
	ui.Token = token
	ui.LastAccess = time.Now()
	uc.storeUserMap(uid, ui)
	uc.storeTokenMap(token, uid)
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
	log := gylib.GetStdLog()
	log.Printf("Removing token %s from logged user", token)
	uid, _ := uc.loadTokenMap(token)
	if db, err := OpenDatabase(); err == nil {
		defer db.Close()
		udm := NewUserDbModel(db)
		if err := udm.DeleteTokenByUserId(uid); err != nil {
			log.Warnf("Cannot remove token %s from database: %s", token, err.Error())
		}
	} else {
		log.Warnf("Cannot remove token %s from database: %s", token, err.Error())
	}
	uc.deleteUserMap(uid)
	uc.deleteTokenMap(token)
	appContestAccess.deleteContestMap(uid)
	appNotifications.deleteNotificationMap(uid)
}

func (uc *UserController) UserLogout(token string) {
	uc.UserRemoveFromList(token)
}

func (uc *UserController) UserLogoutFromWebsite(w http.ResponseWriter, r *http.Request) {
	token := uc.GetLoggedUserToken(r)
	uc.UserLogout(token)
	//uc.SetLoggedUserInfo(w, r, "")
	// Manually delete Session Cookies, as broken cookies prevent anyone to login
	uc.CleanCookies(w, r)
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
		case FlashSuccess:
			flashType = "success"
		}
		flashData := fmt.Sprintf("%s;%s", flashType, message)
		sess.AddFlash(flashData)
		sess.Save(r, w)
	}
}
