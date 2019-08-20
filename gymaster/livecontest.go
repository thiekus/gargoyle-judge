package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/mux"
	mjpeg "github.com/mattn/go-mjpeg"
	"net/http"
	"strconv"
	"time"
)

type CaptureResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func liveHomeGetEndpoint(w http.ResponseWriter, r *http.Request) {
	CompileDashboardPage(w, r, "live_base.html", "live_home.html",
		"", nil, "")
}

func liveCaptureGetEndpoint(w http.ResponseWriter, r *http.Request) {
	user := appUsers.GetLoggedUserInfo(r)
	if user == nil {
		appUsers.AddFlashMessage(w, r, "Please login first!", FlashError)
		urlBase64 := base64.StdEncoding.EncodeToString([]byte(r.URL.Path))
		http.Redirect(w, r, getBaseUrl(r)+"login?target="+urlBase64, 302)
		return
	}
	CompileDashboardPage(w, r, "live_base.html", "live_capture.html",
		"", nil, "")
}

func liveCapturePostEndpoint(w http.ResponseWriter, r *http.Request) {
	cr := CaptureResult{
		Success: false,
		Message: "unknown error",
	}
	if user := appUsers.GetLoggedUserInfo(r); user == nil {
		cr.Message = "User not exists!"
	} else {
		if err := r.ParseMultipartForm(500 * 1024); err != nil {
			cr.Message = err.Error()
		} else {
			//log.Printf("Form: %v", r.Form)
			dataBase64 := r.PostFormValue("data")
			data, err := base64.StdEncoding.DecodeString(dataBase64)
			if err != nil {
				cr.Message = err.Error()
			} else {
				//log.Printf("Image file stream is: %v", data)
				err = appImageStreams.UpdateImageStream(user.Id, data)
				if err != nil {
					cr.Message = err.Error()
				} else {
					cr.Success = true
					cr.Message = "Success"
				}
			}
		}
	}
	//log.Printf("Image live stream result: %s", cr.Message)
	if jsonData, err := json.Marshal(cr); err == nil {
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.Write(jsonData)
	}
}

func liveImageStreamGetEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	if img, err := appImageStreams.GetImageStream(id); err == nil {
		stream := mjpeg.NewStreamWithInterval(500 * time.Millisecond)
		defer stream.Close()
		// Do first frame stream update
		firstUpdate := img.LastUpdate
		err = stream.Update(img.Data)
		if err != nil {
			http.Error(w, "500 Internal Server Error: cannot update mjpeg frame", http.StatusInternalServerError)
			return
		}
		// Update stream on another goroutine
		go func(stream *mjpeg.Stream, id int, firstUpdate int64) {
			lastUpdate := firstUpdate
			for {
				if img, err := appImageStreams.GetImageStream(id); err == nil {
					//log.Printf("Image update on %d, size %d", img.LastUpdate, len(img.Data))
					if img.LastUpdate > lastUpdate {
						lastUpdate = img.LastUpdate
						err := stream.Update(img.Data)
						if err != nil {
							break
						}
					}
				} else {
					break
				}
				time.Sleep(1 * time.Second)
			}
		}(stream, id, firstUpdate)
		stream.ServeHTTP(w, r)
	} else {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}
