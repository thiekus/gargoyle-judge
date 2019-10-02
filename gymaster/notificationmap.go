package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"github.com/dustin/go-humanize"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"html"
	"sync"
	"time"
)

type NotificationChildDetails struct {
	FromUserDisplayName string `json:"fromUserName"`
	FromUserId          int    `json:"fromUserId"`
	FromUserAvatar      string `json:"fromUserAvatar"`
	HasRead             bool   `json:"hasRead"`
	ReceivedTimestamp   int64  `json:"receivedTime"`
	ReceivedTimeStr     string `json:"receivedTimeStr"`
	Description         string `json:"description"`
	UrlLink             string `json:"urlLink"`
}

type NotificationDetails struct {
	Updated       bool                       `json:"updated"`
	UpdateTime    int64                      `json:"updateTime"`
	Notifications []NotificationChildDetails `json:"notifications"`
}

type NotificationMap struct {
	sync.Map
}

type NotificationController struct {
	nmap NotificationMap
}

func MakeNotificationController() NotificationController {
	nc := NotificationController{
		nmap: NotificationMap{},
	}
	return nc
}

func (nc *NotificationController) makeChildDetails(nt gytypes.NotificationData, udm UserDbModel) (NotificationChildDetails, error) {
	targetId := nt.FromUserId
	if targetId == 0 {
		targetId = nt.UserId
	}
	ui, err := udm.GetUserById(targetId)
	if err != nil {
		return NotificationChildDetails{}, err
	}
	displayName := html.EscapeString(ui.DisplayName)
	avatar := html.EscapeString(ui.Avatar)
	if nt.FromUserId == 0 {
		displayName = "System"
	}
	n := NotificationChildDetails{
		FromUserDisplayName: displayName,
		FromUserId:          nt.FromUserId,
		FromUserAvatar:      "/avatar/" + avatar,
		HasRead:             nt.HasRead,
		ReceivedTimestamp:   nt.ReceivedTime.Unix(),
		ReceivedTimeStr:     humanize.Time(nt.ReceivedTime),
		Description:         nt.Description,
		UrlLink:             nt.Link,
	}
	return n, nil
}

func (nc *NotificationController) loadNotificationMap(userId int) (NotificationDetails, bool) {
	nt, exists := nc.nmap.Load(userId)
	ntResult := NotificationDetails{}
	if nt != nil {
		ntResult = nt.(NotificationDetails)
	}
	return ntResult, exists
}

func (nc *NotificationController) storeNotificationMap(userId int, nd NotificationDetails) {
	// Only currently logged user that have notification caching
	if ui := appUsers.GetUserById(userId); ui != nil {
		nc.nmap.Store(userId, nd)
	}
}

func (nc *NotificationController) deleteNotificationMap(userId int) {
	nc.nmap.Delete(userId)
}

func (nc *NotificationController) AddNotification(targetUserId, fromUserId int, description, link string) error {
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()
	ndm := NewNotificationDbModel(db)
	err = ndm.InsertUserNotification(targetUserId, fromUserId, description, link)
	if err != nil {
		return err
	}
	if ui := appUsers.GetUserById(targetUserId); ui != nil {
		_, err = nc.RefreshAjaxNotificationStatus(targetUserId)
	}
	return err
}

func (nc *NotificationController) GetNotifications(userId int) ([]NotificationChildDetails, error) {
	db, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	ndm := NewNotificationDbModel(db)
	notifications, err := ndm.GetUserNotifications(userId, false)
	if err != nil {
		return nil, err
	}
	var nt []NotificationChildDetails
	udm := NewUserDbModel(db)
	for _, v := range notifications {
		n, err := nc.makeChildDetails(v, udm)
		if err != nil {
			return nil, err
		}
		nt = append(nt, n)
	}
	return nt, err
}

func (nc *NotificationController) GetAjaxNotificationStatus(userId int) (NotificationDetails, error) {
	if nd, exists := nc.loadNotificationMap(userId); exists {
		return nd, nil
	} else {
		return nc.RefreshAjaxNotificationStatus(userId)
	}
}

func (nc *NotificationController) MarkNotificationsAllRead(userId int) error {
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()
	ndm := NewNotificationDbModel(db)
	err = ndm.MarkAsReadAll(userId)
	if err != nil {
		return err
	}
	if nd, exists := nc.loadNotificationMap(userId); exists {
		nd.Updated = false
		nd.UpdateTime = time.Now().Unix()
		nc.storeNotificationMap(userId, nd)
	}
	return nil
}

func (nc *NotificationController) RefreshAjaxNotificationStatus(userId int) (NotificationDetails, error) {
	nd := NotificationDetails{}
	db, err := OpenDatabase()
	if err != nil {
		return nd, err
	}
	defer db.Close()
	ndm := NewNotificationDbModel(db)
	ndata, err := ndm.GetUserNotifications(userId, true)
	if err != nil {
		return nd, err
	}
	var nt []NotificationChildDetails
	udm := NewUserDbModel(db)
	for _, v := range ndata {
		n, err := nc.makeChildDetails(v, udm)
		if err != nil {
			return nd, err
		}
		nt = append(nt, n)
	}
	nd = NotificationDetails{
		Updated:       true,
		UpdateTime:    time.Now().Unix(),
		Notifications: nt,
	}
	nc.storeNotificationMap(userId, nd)
	return nd, nil
}
