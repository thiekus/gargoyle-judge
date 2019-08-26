package main

import (
	"errors"
	"time"
)

// map[userId][contestId]
type ContestAccessMap map[int]map[int]ContestAccess

type ContestAccessController struct {
	cmap ContestAccessMap
}

func MakeContestAccessController() ContestAccessController {
	cac := ContestAccessController{
		cmap: make(ContestAccessMap),
	}
	return cac
}

func (cac *ContestAccessController) CalculateRemainingTime(ca *ContestAccess) int {
	if ca.EndTime == 0 {
		// Not counted, given negative value
		return -1
	}
	nowTime := time.Now().Unix()
	delta := int(ca.EndTime - nowTime)
	if delta < 0 {
		delta = 0
	}
	return delta
}

func (cac *ContestAccessController) CheckAccessInfo(ca *ContestAccess) error {
	nowTime := time.Now().Unix()
	if !ca.Allowed {
		return errors.New("you not allowed to attend this contest")
	}
	if (ca.EndTime > 0) && (nowTime > ca.EndTime) {
		return errors.New("your time is over")
	}
	return nil
}

func (cac *ContestAccessController) GetAccessInfoOfUser(userId int, contestId int) (*ContestAccess, error) {
	if _, exists := cac.cmap[userId]; !exists {
		cac.cmap[userId] = make(map[int]ContestAccess)
	}
	// Check if cached in map?
	if ca, exists := cac.cmap[userId][contestId]; !exists {
		// Fetch from database
		cdm, err := NewContestDbModel()
		if err != nil {
			return nil, err
		}
		defer cdm.Close()
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
			if endTime > cd.EndTime {
				endTime = cd.EndTime // Keep not over schedule
			}
			// Time prevention
			if (cd.StartTime > 0) && (cd.EndTime > 0) {
				if nowTime < cd.StartTime {
					return nil, errors.New("contest not yet started")
				}
				if nowTime > cd.EndTime {
					return nil, errors.New("contest is over")
				}
			} else {
				endTime = 0
			}
			ca = ContestAccess{
				UserId:    userId,
				ContestId: contestId,
				StartTime: nowTime,
				EndTime:   endTime,
				Allowed:   true,
			}
			err = cdm.InsertContestAccess(ca)
			if err != nil {
				return nil, err
			}
		}
		ca.RemainTime = cac.CalculateRemainingTime(&ca)
		// Save to map cache
		cac.cmap[userId][contestId] = ca
		return &ca, nil
	} else {
		ca.RemainTime = cac.CalculateRemainingTime(&ca)
		// Retrieve from map cache as exists
		return &ca, nil
	}
}

func (cac *ContestAccessController) ReleaseMapOfUser(userId int) {
	delete(cac.cmap, userId)
}
