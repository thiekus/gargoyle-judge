package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"github.com/o1egl/govatar"
	"github.com/patrickmn/go-cache"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var avatarCache *cache.Cache = nil

func initializeAvatarCache() {
	// Lazy initialization of avatar cache
	if avatarCache == nil {
		avatarCache = cache.New(30*time.Minute, 1*time.Hour)
	}
}

func getPersonalizedUserAvatar(uid int, avatarType string) string {
	switch avatarType {
	case "genFaceUsername":
		ui := appUsers.GetUserById(uid)
		avatarStr := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", ui.Gender, getMD5Hash(ui.Username))))
		return avatarStr
	case "gravatar":
		ui := appUsers.GetUserById(uid)
		avatarStr := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("G:%s", getMD5Hash(ui.Email))))
		return avatarStr
	default:
		// genFaceRandom
		ui := appUsers.GetUserById(uid)
		avatarStr := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", ui.Gender, generateRandomSalt())))
		return avatarStr
	}
}

// Force clean cached avatar if user desired
func dropUserAvatarCache(avatarStr string) {
	initializeAvatarCache()
	avatarHash := getMD5Hash(avatarStr)
	// Must be checked if exists before deleting
	if _, cached := avatarCache.Get(avatarHash); cached {
		avatarCache.Delete(avatarHash)
		avatarCache.Delete(avatarHash + ":type")
	}
}

func generateAvatar(cacheHash string, w http.ResponseWriter, r *http.Request, gender string, seed string) {
	genderSel := govatar.MALE
	if gender == "F" {
		genderSel = govatar.FEMALE
	}
	img, err := govatar.GenerateFromUsername(genderSel, seed)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Resize generated avatar, max to 200px
	imgResized := resize.Resize(200, 0, img, resize.Lanczos3)
	var b bytes.Buffer
	err = jpeg.Encode(&b, imgResized, &jpeg.Options{Quality: 75})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Set for caching
	avatarCache.Set(cacheHash, b.Bytes(), cache.NoExpiration)
	avatarCache.Set(cacheHash+":type", "image/jpeg", cache.NoExpiration)
	// Transmit actual avatar
	w.Header().Add("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(b.Bytes())
}

func getGravatarAvatar(cacheHash string, w http.ResponseWriter, r *http.Request, emailHash string) {
	// Prevent someone to abuse gravatar request point
	if !isHexValue(emailHash) {
		http.Error(w, "500 Internal Server Error: invalid gravatar hash", http.StatusInternalServerError)
		return
	}
	gravatarUrl := fmt.Sprintf("https://www.gravatar.com/avatar/%s?s=200&d=robohash", emailHash)
	resp, err := http.Get(gravatarUrl)
	if err != nil {
		http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	contentType := resp.Header.Get("Content-Type")
	avatarCache.Set(cacheHash, body, cache.DefaultExpiration)
	avatarCache.Set(cacheHash+":type", contentType, cache.DefaultExpiration)
	w.Header().Add("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(body)
}

func avatarGetEndpoint(w http.ResponseWriter, r *http.Request) {
	initializeAvatarCache()
	vars := mux.Vars(r)
	avatarInfo := vars["avatarInfo"]
	cacheHash := getMD5Hash(avatarInfo)
	if avatarData, cached := avatarCache.Get(cacheHash); cached {
		contentType, _ := avatarCache.Get(cacheHash + ":type")
		w.Header().Set("Content-Type", contentType.(string))
		w.Header().Set("Content-Length", strconv.Itoa(len(avatarData.([]byte))))
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write(avatarData.([]byte))
	} else {
		// Not cached, go ahead
		decodedData, err := base64.StdEncoding.DecodeString(avatarInfo)
		if err != nil {
			http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		decoded := strings.SplitN(string(decodedData), ":", 2)
		switch decoded[0] {
		case "M":
			generateAvatar(cacheHash, w, r, "M", decoded[1])
		case "F":
			generateAvatar(cacheHash, w, r, "F", decoded[1])
		case "G":
			getGravatarAvatar(cacheHash, w, r, decoded[1])
		default:
			http.Error(w, "404 Not Found", http.StatusNotFound)
		}
	}
}
