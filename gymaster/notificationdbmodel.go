package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"time"

	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
)

type NotificationDbModel struct {
	db DbContext
}

func NewNotificationDbModel(db DbContext) NotificationDbModel {
	ndm := NotificationDbModel{
		db: db,
	}
	return ndm
}

func (ndm *NotificationDbModel) GetUserNotifications(userId int, unreadOnly bool) ([]gytypes.NotificationData, error) {
	db := ndm.db
	query := `SELECT id, id_user, id_user_from, received_time, has_read, description, link FROM {{.TablePrefix}}notifications
        WHERE (id_user = ?) AND ((has_read = ?) OR (1 = ?)) ORDER BY received_time DESC`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(userId, !unreadOnly, gylib.Btoi(!unreadOnly))
	if err != nil {
		return nil, err
	}
	var notifications []gytypes.NotificationData
	for rows.Next() {
		nt := gytypes.NotificationData{}
		var utReceivedTime int64
		err = rows.Scan(
			&nt.Id,
			&nt.UserId,
			&nt.FromUserId,
			&utReceivedTime,
			&nt.HasRead,
			&nt.Description,
			&nt.Link,
		)
		if err != nil {
			return nil, err
		}
		nt.ReceivedTime = time.Unix(utReceivedTime, 0)
		notifications = append(notifications, nt)
	}
	return notifications, nil
}

func (ndm *NotificationDbModel) InsertUserNotification(targetUserId, fromUserId int, description, link string) error {
	db := ndm.db
	query := `INSERT INTO {{.TablePrefix}}notifications (id_user, id_user_from, received_time, has_read, description, link)
        VALUES (?, ?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	sentTime := time.Now().Unix()
	_, err = stmt.Exec(targetUserId, fromUserId, sentTime, false, description, link)
	return err
}

func (ndm *NotificationDbModel) MarkAsReadAll(userId int) error {
	db := ndm.db
	query := `UPDATE {{.TablePrefix}}notifications SET has_read = 1 WHERE id_user = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(userId)
	return err
}
