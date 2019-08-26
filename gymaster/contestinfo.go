package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import "html/template"

type ContestData struct {
	Id           int
	Title        string
	Description  template.HTML
	TimeDesc     template.HTML
	ContestUrl   string
	ProblemCount int
	GroupId      int
	Unlocked     bool
	PublicView   bool
	MustStream   bool
	StartTime    int64
	EndTime      int64
	MaxTime      int
}

type ProblemData struct {
	Id          int
	ContestId   int
	Name        string
	Description template.HTML
	TimeLimit   int
	MemLimit    int
	MaxAttempts int
	ProblemUrl  string
	ContestUrl  string
}

type ContestAccess struct {
	Id         int
	UserId     int
	ContestId  int
	StartTime  int64
	EndTime    int64
	Allowed    bool
	RemainTime int // Used for written in page
}

type ContestList struct {
	Count    int
	Contests []ContestData
}

type ProblemSet struct {
	Count    int
	Contest  ContestData
	Problems []ProblemData
}
