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
	"github.com/dustin/go-humanize"
	"net/http"
	"time"
)

type AjaxCommonStatus struct {
	Succeeded bool `json:"succeeded"`
}

func ajaxGetNotifications(w http.ResponseWriter, r *http.Request) {
	uid := appUsers.GetLoggedUserId(r)
	if uid == 0 {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	nd, err := appNotifications.GetAjaxNotificationStatus(uid)
	if err != nil {
		http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	for i := range nd.Notifications {
		nd.Notifications[i].ReceivedTimeStr = humanize.Time(time.Unix(nd.Notifications[i].ReceivedTimestamp, 0))
	}
	if data, err := json.Marshal(nd); err == nil {
		w.Write(data)
	} else {
		http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func ajaxReadAllNotifications(w http.ResponseWriter, r *http.Request) {
	uid := appUsers.GetLoggedUserId(r)
	if uid == 0 {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	err := appNotifications.MarkNotificationsAllRead(uid)
	if err != nil {
		http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if data, err := json.Marshal(AjaxCommonStatus{Succeeded: true}); err == nil {
		w.Write(data)
	} else {
		http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
