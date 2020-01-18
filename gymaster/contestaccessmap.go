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
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"sync"
	"time"
)

// map[userId][contestId]
//type ContestAccessMap map[int]map[int]ContestAccess
type ContestInfoMap struct {
	sync.Map
}
type ContestAccessMap struct {
	sync.Map
	cimap ContestInfoMap
}

type ContestAccessController struct {
	cmap ContestAccessMap
}

func MakeContestAccessController() ContestAccessController {
	cac := ContestAccessController{
		cmap: ContestAccessMap{},
	}
	return cac
}

func (cac *ContestAccessController) deleteContestMapFromUser(cim ContestInfoMap, contestId int) {
	cim.Delete(contestId)
}

func (cac *ContestAccessController) deleteContestMap(uid int) {
	cim, exists := cac.loadContestAccessMapOfUser(uid)
	if exists {
		cac.rangeContestMapOfUser(cim, func(contestId int, ca gytypes.ContestAccess) bool {
			cac.deleteContestMapFromUser(cim, contestId)
			return true
		})
	}
	cac.cmap.Delete(uid)
}

func (cac *ContestAccessController) loadContestAccessMapFromUser(cim ContestInfoMap, contestId int) (gytypes.ContestAccess, bool) {
	ca, exists := cim.Load(contestId)
	caResult := gytypes.ContestAccess{}
	if ca != nil {
		caResult = ca.(gytypes.ContestAccess)
	}
	return caResult, exists
}

func (cac *ContestAccessController) loadContestAccessMapOfUser(userId int) (ContestInfoMap, bool) {
	ci, exists := cac.cmap.Load(userId)
	ciResult := ContestInfoMap{}
	if ci != nil {
		ciResult = ci.(ContestInfoMap)
	}
	return ciResult, exists
}

func (cac *ContestAccessController) loadContestAccessMap(userId, contestId int) (gytypes.ContestAccess, bool) {
	ci, exists := cac.loadContestAccessMapOfUser(userId)
	if !exists {
		return gytypes.ContestAccess{}, false
	} else {
		return cac.loadContestAccessMapFromUser(ci, contestId)
	}
}

func (cac *ContestAccessController) rangeContestMapOfUser(cim ContestInfoMap, f func(key int, value gytypes.ContestAccess) bool) {
	cim.Range(func(key, value interface{}) bool {
		return f(key.(int), value.(gytypes.ContestAccess))
	})
}

func (cac *ContestAccessController) storeContestAccessMap(userId int, contestId int, ca gytypes.ContestAccess) {
	ci, exists := cac.loadContestAccessMapOfUser(userId)
	if exists {
		ci.Store(contestId, ca)
	} else {
		ci := ContestInfoMap{}
		ci.Store(contestId, ca)
		cac.cmap.Store(userId, ci)
	}
}

func (cac *ContestAccessController) CalculateRemainingTime(ca *gytypes.ContestAccess) int {
	utEndTime := ca.EndTime.Unix()
	if utEndTime == 0 {
		// Not counted, given negative value
		return -1
	}
	nowTime := time.Now().Unix()
	delta := int(utEndTime - nowTime)
	if delta < 0 {
		delta = 0
	}
	return delta
}

func (cac *ContestAccessController) CheckAccessInfo(ca *gytypes.ContestAccess) error {
	nowTime := time.Now().Unix()
	if !ca.Allowed {
		return errors.New("you not allowed to attend this contest")
	}
	utEndTime := ca.EndTime.Unix()
	if (utEndTime > 0) && (nowTime > utEndTime) {
		return errors.New("your time is over")
	}
	return nil
}

func (cac *ContestAccessController) GetAccessInfoOfUser(userId int, contestId int) (*gytypes.ContestAccess, error) {
	/*if _, exists := cac.loadContestAccessMapOfUser(userId); !exists {
		cac.cmap[userId] = make(map[int]ContestAccess)
	}*/
	// Check if cached in map?
	if ca, exists := cac.loadContestAccessMap(userId, contestId); !exists {
		// Fetch from database
		db, err := OpenDatabase()
		if err != nil {
			return nil, err
		}
		defer db.Close()
		cdm := NewContestDbModel(db)
		// Check if row is available. In this case, err will only return if db-related error occurred
		// Will return row=0 and err=nil if not available but no db error
		rowNum, err := cdm.GetContestAccessCount(contestId, userId)
		if err != nil {
			return nil, err
		}
		if rowNum > 0 {
			// Now get ContestAccess entry
			ca, err = cdm.GetContestAccessOfUserId(contestId, userId)
			if err != nil {
				return nil, err
			}
		} else {
			// Not exists? Insert new
			cd, err := cdm.GetContestDetails(contestId)
			if err != nil {
				return nil, err
			}
			// Check if active
			if !cd.Active {
				return nil, errors.New("contest not active")
			}
			// Check group access
			if cd.GroupId != 0 {
				// Get groups access for selected user
				ui := appUsers.GetUserById(userId)
				if ui == nil {
					return nil, errors.New("unknown user")
				}
				found := false
				for _, v := range ui.Groups {
					if v.GroupId == cd.GroupId {
						found = true
						break
					}
				}
				if !found {
					return nil, errors.New("not allowed to attend because not registered as defined group")
				}
			}
			// Check time
			nowTime := time.Now().Unix()
			endTime := nowTime + int64(cd.MaxTime) // Maximum end time
			utStartTime := cd.StartTime.Unix()
			utEndTime := cd.EndTime.Unix()
			if endTime > utEndTime {
				endTime = utEndTime // Keep not over schedule
			}
			// Time prevention
			if (utStartTime > 0) && (utEndTime > 0) && (cd.MaxTime > 0) {
				if nowTime < utStartTime {
					return nil, errors.New("contest not yet started")
				}
				if nowTime > utEndTime {
					return nil, errors.New("contest is over")
				}
			} else {
				endTime = 0
			}
			ca = gytypes.ContestAccess{
				UserId:    userId,
				ContestId: contestId,
				StartTime: time.Unix(nowTime, 0),
				EndTime:   time.Unix(endTime, 0),
				Allowed:   true,
			}
			err = cdm.InsertContestAccess(ca)
			if err != nil {
				return nil, err
			}
		}
		ca.RemainTime = cac.CalculateRemainingTime(&ca)
		// Save to map cache
		cac.storeContestAccessMap(userId, contestId, ca)
		// Invalidate scoreboard cache after enter
		appScoreboard.InvalidateScoreboardCache(contestId)
		return &ca, nil
	} else {
		ca.RemainTime = cac.CalculateRemainingTime(&ca)
		// Retrieve from map cache as exists
		return &ca, nil
	}
}

func (cac *ContestAccessController) ReleaseMapOfUser(userId int) {
	cac.deleteContestMap(userId)
}
