package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import "html/template"

type ContestData struct {
	Id          int
	ContestType string
	Title       string
	Description template.HTML
	TimeDesc    template.HTML
	ContestUrl  string
	QuestCount  int
	GroupId     int
	Unlocked    bool
	PublicView  bool
	Trainer     bool
	MustStream  bool
	StartTime   int
	EndTime     int
	MaxTime     int
}

type ProblemData struct {
	Id          int
	ContestId   int
	Name        string
	Description template.HTML
	TimeLimit   int
	MemLimit    int
	MaxAttempts int
	QuestUrl    string
	ContestUrl  string
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
