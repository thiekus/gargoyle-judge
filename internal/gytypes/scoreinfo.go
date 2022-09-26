package gytypes

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"time"
)

type ScoreboardData struct {
	ContestId       int
	ContestName     string
	ContestStyle    string
	ContestantCount int
	Contestant      ScoreContestantDataList
	Problems        []ScoreContestProblemData
	LastUpdate      time.Time
}

type ScoreContestInfo struct {
	ContestId      int
	Title          string
	Style          string
	EnableFreeze   bool
	FreezeTime     time.Time
	UnfreezeTime   time.Time
	AllowPublic    bool
	StartTimestamp time.Time
	PenaltyTime    int64
}

type ScoreContestProblemData struct {
	ContestId   int
	ProblemId   int
	Name        string
	ShortName   string
	CircleColor string
}

type ScoreContestantData struct {
	UserId           int
	Name             string
	Institution      string
	Avatar           string
	CountryCode      string
	RankNumber       int
	TotalScore       int
	TotalPenaltyTime int64
	Problems         []ScoreProblemData
	PenaltyTimeStr   string
	StartTime        int64
	EndTime          int64
}

type ScoreContestantDataList []ScoreContestantData

type ScoreProblemData struct {
	ContestId       int
	ProblemId       int
	UserId          int
	Score           int
	AcceptedTime    int64
	PenaltyTime     int64
	SubmissionCount int
	OneHit          bool
	Regraded        bool
	AcceptedTimeStr string
}

type ScoreboardListData struct {
	ContestId       int
	ContestName     string
	ContestantCount int
	AllowPublic     bool
	Updated         bool
	LastUpdate      time.Time
}

const (
	ScoreStyleICPC = "ICPC"
	ScoreStyleIOI  = "IOI"
)

func (cd ScoreContestantDataList) Len() int {
	return len(cd)
}

func (cd ScoreContestantDataList) Swap(i, j int) {
	cd[i], cd[j] = cd[j], cd[i]
}

func (cd ScoreContestantDataList) Less(i, j int) bool {
	// Sort from biggest point to lesser point
	// TODO: implements scoring sort other than ICPC
	if cd[i].TotalScore == cd[j].TotalScore {
		// Bigger penalty, then lower rank for same score
		return cd[i].TotalPenaltyTime < cd[j].TotalPenaltyTime
	}
	return cd[i].TotalScore > cd[j].TotalScore
}
