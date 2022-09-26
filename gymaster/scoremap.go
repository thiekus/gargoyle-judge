package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"errors"
	"sync"
	"time"

	"github.com/thiekus/gargoyle-judge/internal/gytypes"
)

type ScoreboardDataContext struct {
	gytypes.ScoreboardData
	NeedRefresh bool
	LastUpdate  time.Time
}

type ScoreboardMap struct {
	sync.Map
}

type ScoreboardController struct {
	smapPublic  ScoreboardMap
	smapPrivate ScoreboardMap
}

func MakeScoreboardController() ScoreboardController {
	sbc := ScoreboardController{
		smapPublic:  ScoreboardMap{},
		smapPrivate: ScoreboardMap{},
	}
	return sbc
}

func (sbc *ScoreboardController) loadPublicScoreboardMap(contestId int) (ScoreboardDataContext, bool) {
	sb, exists := sbc.smapPublic.Load(contestId)
	sbResult := ScoreboardDataContext{}
	if sb != nil {
		sbResult = sb.(ScoreboardDataContext)
	}
	return sbResult, exists
}

func (sbc *ScoreboardController) loadPrivateScoreboardMap(contestId int) (ScoreboardDataContext, bool) {
	sb, exists := sbc.smapPrivate.Load(contestId)
	sbResult := ScoreboardDataContext{}
	if sb != nil {
		sbResult = sb.(ScoreboardDataContext)
	}
	return sbResult, exists
}

func (sbc *ScoreboardController) storePublicScoreboardMap(contestId int, sb ScoreboardDataContext) {
	sbc.smapPublic.Store(contestId, sb)
}

func (sbc *ScoreboardController) storePrivateScoreboardMap(contestId int, sb ScoreboardDataContext) {
	sbc.smapPrivate.Store(contestId, sb)
}

func (sbc *ScoreboardController) deletePublicScoreboardMap(contestId int) {
	sbc.smapPublic.Delete(contestId)
}

func (sbc *ScoreboardController) deletePrivateScoreboardMap(contestId int) {
	sbc.smapPrivate.Delete(contestId)
}

func (sbc *ScoreboardController) GetPublicScoreboard(contestId int) (gytypes.ScoreboardData, error) {
	if sb, exists := sbc.loadPublicScoreboardMap(contestId); exists {
		if sb.NeedRefresh {
			if err := sbc.RefreshPublicScoreboard(contestId); err != nil {
				return gytypes.ScoreboardData{}, err
			}
		} else {
			return sb.ScoreboardData, nil
		}
	} else {
		if err := sbc.RefreshPublicScoreboard(contestId); err != nil {
			return gytypes.ScoreboardData{}, err
		}
	}
	if sb, exists := sbc.loadPublicScoreboardMap(contestId); exists {
		return sb.ScoreboardData, nil
	} else {
		return gytypes.ScoreboardData{}, errors.New("cannot load public scoreboard map")
	}
}

func (sbc *ScoreboardController) GetPrivateScoreboard(contestId int) (gytypes.ScoreboardData, error) {
	if sb, exists := sbc.loadPrivateScoreboardMap(contestId); exists {
		if sb.NeedRefresh {
			if err := sbc.RefreshPrivateScoreboard(contestId); err != nil {
				return gytypes.ScoreboardData{}, err
			}
		} else {
			return sb.ScoreboardData, nil
		}
	} else {
		if err := sbc.RefreshPrivateScoreboard(contestId); err != nil {
			return gytypes.ScoreboardData{}, err
		}
	}
	if sb, exists := sbc.loadPrivateScoreboardMap(contestId); exists {
		return sb.ScoreboardData, nil
	} else {
		return gytypes.ScoreboardData{}, errors.New("cannot load public scoreboard map")
	}
}

func (sbc *ScoreboardController) GetScoreboardByUser(userInfo *gytypes.UserInfo, contestId int) (gytypes.ScoreboardData, error) {
	if userInfo == nil {
		return gytypes.ScoreboardData{}, errors.New("invalid user")
	}
	if (userInfo.IsAdmin()) || (userInfo.IsJury()) {
		return sbc.GetPrivateScoreboard(contestId)
	}
	return sbc.GetPublicScoreboard(contestId)
}

func (sbc *ScoreboardController) GetPublicScoreboardList() ([]gytypes.ScoreboardListData, error) {
	db, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	cdm := NewContestDbModel(db)
	list, err := cdm.GetContestListForScoreboard(true)
	if err != nil {
		return nil, err
	}
	for i, v := range list {
		if _, exists := sbc.loadPublicScoreboardMap(v.ContestId); exists {
			list[i].LastUpdate = v.LastUpdate
			list[i].Updated = true
		}
	}
	return list, nil
}

func (sbc *ScoreboardController) GetPrivateScoreboardList() ([]gytypes.ScoreboardListData, error) {
	db, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	cdm := NewContestDbModel(db)
	list, err := cdm.GetContestListForScoreboard(false)
	if err != nil {
		return nil, err
	}
	for i, v := range list {
		if _, exists := sbc.loadPrivateScoreboardMap(v.ContestId); exists {
			list[i].LastUpdate = v.LastUpdate
			list[i].Updated = true
		}
	}
	return list, nil
}

func (sbc *ScoreboardController) GetScoreboardListByUser(userInfo *gytypes.UserInfo) ([]gytypes.ScoreboardListData, error) {
	if userInfo == nil {
		return nil, errors.New("invalid user")
	}
	if (userInfo.IsAdmin()) || (userInfo.IsJury()) {
		return sbc.GetPrivateScoreboardList()
	}
	return sbc.GetPublicScoreboardList()
}

func (sbc *ScoreboardController) InvalidateScoreboardCache(contestId int) {
	// Invalidate caches
	if sdc, exists := sbc.loadPrivateScoreboardMap(contestId); exists {
		sdc.NeedRefresh = true
		sbc.storePrivateScoreboardMap(contestId, sdc)
	}
	if sdc, exists := sbc.loadPublicScoreboardMap(contestId); exists {
		sdc.NeedRefresh = true
		sbc.storePublicScoreboardMap(contestId, sdc)
	}
}

func (sbc *ScoreboardController) RefreshPublicScoreboard(contestId int) error {
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()
	sdm := NewScoreDbModel(db)
	sb, err := sdm.GetScoreboardForContest(contestId, true)
	if err != nil {
		return err
	}
	cc := sbc.GetColorWheel(len(sb.Problems))
	for i := range sb.Problems {
		sb.Problems[i].CircleColor = cc[i]
	}
	sbm := ScoreboardDataContext{
		ScoreboardData: *sb,
		NeedRefresh:    false,
		LastUpdate:     sb.LastUpdate,
	}
	sbc.storePublicScoreboardMap(contestId, sbm)
	return nil
}

func (sbc *ScoreboardController) RefreshPrivateScoreboard(contestId int) error {
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()
	sdm := NewScoreDbModel(db)
	sb, err := sdm.GetScoreboardForContest(contestId, false)
	if err != nil {
		return err
	}
	cc := sbc.GetColorWheel(len(sb.Problems))
	for i := range sb.Problems {
		sb.Problems[i].CircleColor = cc[i]
	}
	sbm := ScoreboardDataContext{
		ScoreboardData: *sb,
		NeedRefresh:    false,
		LastUpdate:     sb.LastUpdate,
	}
	sbc.storePrivateScoreboardMap(contestId, sbm)
	return nil
}

func (sbc *ScoreboardController) GetColorWheel(arrayLen int) []string {
	colors := []string{"#FF4136", "#85144B", "#F012BE", "#B10DC9", "#0074D9", "#7FDBFF",
		"#39CCCC", "#3D9970", "#2ECC40", "#01FF70", "#FFDC00", "#FF851B"}
	cl := len(colors)
	var c []string
	for i := 0; i < arrayLen; i++ {
		c = append(c, colors[i%cl])
	}
	return c
}

func (sbc *ScoreboardController) SubmitScore(problemId, userId, score int, accepted bool) error {
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()
	sdm := NewScoreDbModel(db)
	ci, err := sdm.GetContestInfoByProblemId(problemId)
	if err != nil {
		return err
	}
	err = sdm.SubmitIntoScoreboard(problemId, userId, score, accepted)
	if err != nil {
		return err
	}
	// Now invalidate caches
	sbc.InvalidateScoreboardCache(ci.ContestId)
	return nil
}
