package gytypes

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import "time"

type NotificationData struct {
	Id           int
	UserId       int
	FromUserId   int
	ReceivedTime time.Time
	HasRead      bool
	Description  string
	Link         string
}
