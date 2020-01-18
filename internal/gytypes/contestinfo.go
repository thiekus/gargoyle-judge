package gytypes

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"html/template"
	"time"
)

type ContestData struct {
	Id           int
	Title        string
	Description  template.HTML
	Style        string
	AllowedLang  string
	TimeDesc     template.HTML
	ContestUrl   string
	ProblemCount int
	GroupId      int
	EnableFreeze bool
	Active       bool
	PublicView   bool
	MustStream   bool
	StartTime    time.Time
	EndTime      time.Time
	FreezeTime   time.Time
	UnfreezeTime time.Time
	MaxTime      int
}

type ProblemData struct {
	Id          int
	ContestId   int
	Name        string
	ShortName   string
	Description template.HTML
	TimeLimit   int
	MemLimit    int
	MaxAttempts int
	ProblemUrl  string
	ContestUrl  string
	AllowedLang string
}

type ContestAccess struct {
	UserId     int
	ContestId  int
	StartTime  time.Time
	EndTime    time.Time
	Allowed    bool
	RemainTime int // Used for written in page
}

type ProblemSet struct {
	Count    int
	Contest  ContestData
	Problems []ProblemData
}
