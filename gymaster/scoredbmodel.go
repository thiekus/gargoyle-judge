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
	"fmt"
	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"log"
	"sort"
	"time"
)

type ScoreDbModel struct {
	db DbContext
}

func NewScoreDbModel(db DbContext) ScoreDbModel {
	sdm := ScoreDbModel{
		db: db,
	}
	return sdm
}

func (sdm *ScoreDbModel) GetContestInfoById(contestId int) (gytypes.ScoreContestInfo, error) {
	scd := gytypes.ScoreContestInfo{}
	db := sdm.db
	// First, query from contest info
	query := `SELECT c.id, c.title, c.style, c.enable_freeze, c.freeze_timestamp, c.unfreeze_timestamp, c.allow_public, 
        c.start_timestamp, c.penalty_time FROM {{.TablePrefix}}contests as c WHERE c.id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return scd, err
	}
	defer stmt.Close()
	var utFreezeTime, utUnfreezeTime int64
	var utStartTime int64
	err = stmt.QueryRow(contestId).Scan(
		&scd.ContestId,
		&scd.Title,
		&scd.Style,
		&scd.EnableFreeze,
		&utFreezeTime,
		&utUnfreezeTime,
		&scd.AllowPublic,
		&utStartTime,
		&scd.PenaltyTime,
	)
	if err != nil {
		return scd, err
	}
	scd.FreezeTime = time.Unix(utFreezeTime, 0)
	scd.UnfreezeTime = time.Unix(utUnfreezeTime, 0)
	scd.StartTimestamp = time.Unix(utStartTime, 0)
	return scd, nil
}

func (sdm *ScoreDbModel) GetContestInfoByProblemId(problemId int) (gytypes.ScoreContestInfo, error) {
	scd := gytypes.ScoreContestInfo{}
	db := sdm.db
	// First, query from contest info
	query := `SELECT c.id, c.title, c.style, c.enable_freeze, c.freeze_timestamp, c.unfreeze_timestamp, c.allow_public, 
        c.start_timestamp, c.penalty_time FROM {{.TablePrefix}}problems as p INNER JOIN 
        {{.TablePrefix}}contests as c ON c.id = p.contest_id WHERE p.id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return scd, err
	}
	defer stmt.Close()
	var utFreezeTime, utUnfreezeTime int64
	var utStartTime int64
	err = stmt.QueryRow(problemId).Scan(
		&scd.ContestId,
		&scd.Title,
		&scd.Style,
		&scd.EnableFreeze,
		&utFreezeTime,
		&utUnfreezeTime,
		&scd.AllowPublic,
		&utStartTime,
		&scd.PenaltyTime,
	)
	if err != nil {
		return scd, err
	}
	scd.FreezeTime = time.Unix(utFreezeTime, 0)
	scd.UnfreezeTime = time.Unix(utUnfreezeTime, 0)
	scd.StartTimestamp = time.Unix(utStartTime, 0)
	return scd, nil
}

func (sdm *ScoreDbModel) GetScoreboardForContest(contestId int, publicBoard bool) (*gytypes.ScoreboardData, error) {
	db := sdm.db
	// First, query from contest info
	sci, err := sdm.GetContestInfoById(contestId)
	if err != nil {
		return nil, err
	}
	if (sci.Style != gytypes.ScoreStyleICPC) && (sci.Style != gytypes.ScoreStyleIOI) {
		return nil, errors.New("unsupported contest style")
	}
	if (publicBoard) && (!sci.AllowPublic) {
		return nil, errors.New("public scoreboard not allowed")
	}
	// Second, query all problems information
	queryProblems := `SELECT contest_id, id, problem_name, problem_shortname FROM {{.TablePrefix}}problems 
        WHERE contest_id = ? ORDER BY problem_shortname ASC`
	stmtProblems, err := db.Prepare(queryProblems)
	if err != nil {
		return nil, err
	}
	defer stmtProblems.Close()
	rowsProblems, err := stmtProblems.Query(contestId)
	if err != nil {
		return nil, err
	}
	var contestProblems []gytypes.ScoreContestProblemData
	var contestProblemsName []string
	contestProblemsId := make(map[int]int) // reverse problemId to slice index sorted by problems
	contestProblemsUnsorted := make(map[string]gytypes.ScoreContestProblemData)
	for rowsProblems.Next() {
		var cp gytypes.ScoreContestProblemData
		err = rowsProblems.Scan(
			&cp.ContestId,
			&cp.ProblemId,
			&cp.Name,
			&cp.ShortName,
		)
		if err != nil {
			return nil, err
		}
		contestProblemsUnsorted[cp.ShortName] = cp
		contestProblemsName = append(contestProblemsName, cp.ShortName)
	}
	// Make sure slices sorted by shortname
	sort.Strings(contestProblemsName)
	for i, v := range contestProblemsName {
		problem := contestProblemsUnsorted[v]
		contestProblems = append(contestProblems, problem)
		contestProblemsId[problem.ProblemId] = i
	}
	// Third, query all active contestant of contest
	queryUsers := `SELECT u.id, u.display_name, u.institution, u.country_id, u.avatar FROM {{.TablePrefix}}contest_access as a 
        INNER JOIN {{.TablePrefix}}users as u ON a.id_user = u.id WHERE a.id_contest = ?`
	stmtUsers, err := db.Prepare(queryUsers)
	if err != nil {
		return nil, err
	}
	defer stmtUsers.Close()
	users := make(map[int]gytypes.ScoreContestantData)
	rowsUsers, err := stmtUsers.Query(contestId)
	if err != nil {
		return nil, err
	}
	contestantCount := 0
	for rowsUsers.Next() {
		u := gytypes.ScoreContestantData{
			RankNumber:       0,
			TotalScore:       0,
			TotalPenaltyTime: 0,
			PenaltyTimeStr:   "00:00:00",
		}
		err = rowsUsers.Scan(
			&u.UserId,
			&u.Name,
			&u.Institution,
			&u.CountryCode,
			&u.Avatar,
		)
		if err != nil {
			return nil, err
		}
		// Make blank score, assumed user not yet submit something
		var probs []gytypes.ScoreProblemData
		for _, cp := range contestProblems {
			p := gytypes.ScoreProblemData{
				ContestId:       cp.ContestId,
				ProblemId:       cp.ProblemId,
				UserId:          u.UserId,
				Score:           0,
				AcceptedTime:    0,
				PenaltyTime:     0,
				SubmissionCount: 0,
				OneHit:          false,
				Regraded:        false,
				AcceptedTimeStr: "00:00:00",
			}
			probs = append(probs, p)
		}
		u.Problems = probs
		users[u.UserId] = u
		contestantCount++
	}
	// Finally, query scores using single query
	selTable := "{{.TablePrefix}}scores_private"
	if publicBoard {
		selTable = "{{.TablePrefix}}scores_public"
	}
	queryScores := `SELECT id_contest, id_problem, id_user, score, accepted_time, penalty_time, submission_count,
        one_hit, regraded FROM %s WHERE id_contest = ?`
	queryScores = fmt.Sprintf(queryScores, selTable)
	stmtScore, err := db.Prepare(queryScores)
	if err != nil {
		return nil, err
	}
	defer stmtScore.Close()
	rowsScore, err := stmtScore.Query(contestId)
	if err != nil {
		return nil, err
	}
	for rowsScore.Next() {
		score := gytypes.ScoreProblemData{}
		err = rowsScore.Scan(
			&score.ContestId,
			&score.ProblemId,
			&score.UserId,
			&score.Score,
			&score.AcceptedTime,
			&score.PenaltyTime,
			&score.SubmissionCount,
			&score.OneHit,
			&score.Regraded,
		)
		if err != nil {
			return nil, err
		}
		// Score comparison is unlikely needed as filtered by SQL, but won't we paranoid?
		if score.ContestId == contestId {
			if user, exists := users[score.UserId]; exists {
				if problemIndex, exists := contestProblemsId[score.ProblemId]; exists {
					user.TotalScore += score.Score
					// On ICPC, penalty time considered to
					if sci.Style == gytypes.ScoreStyleICPC {
						user.TotalPenaltyTime += score.PenaltyTime
						if score.AcceptedTime > 0 {
							submitPenaltyTime := int64(0)
							startTime := sci.StartTimestamp.Unix()
							log.Printf("user %s: start from %d", user.Name, startTime)
							if startTime > 0 {
								submitPenaltyTime = score.AcceptedTime - startTime
							} else {
								// TODO: start time for unlimited contest time
								submitPenaltyTime = score.AcceptedTime - 1570348800 // - contest access start time
							}
							if submitPenaltyTime > 0 {
								user.TotalPenaltyTime += submitPenaltyTime
							}
						}
					}
					if score.AcceptedTime > 0 {
						startTime := sci.StartTimestamp.Unix()
						if startTime > 0 {
							score.AcceptedTime = score.AcceptedTime - startTime
						} else {
							// TODO: start time for unlimited contest time
							score.AcceptedTime = score.AcceptedTime - 1570348800 // - contest access start time
						}
						score.AcceptedTimeStr = gylib.TimeToHMS(time.Unix(score.AcceptedTime - 25200, 0))
					}
					user.PenaltyTimeStr = gylib.TimeToHMS(time.Unix(user.TotalPenaltyTime - 25200, 0))
					// Replace again with modified user info
					user.Problems[problemIndex] = score
					users[user.UserId] = user
				}
			}
		}
	}
	// Contestant as to be sorted user by rank
	var contestants gytypes.ScoreContestantDataList
	for _, cs := range users {
		contestants = append(contestants, cs)
	}
	// See scoreinfo.go in gytypes how sort interface method was implemented
	sort.Sort(contestants)
	for i := range contestants {
		contestants[i].RankNumber = i + 1
	}
	// Done!
	sb := gytypes.ScoreboardData{
		ContestId:       contestId,
		ContestName:     sci.Title,
		ContestStyle:    sci.Style,
		ContestantCount: contestantCount,
		Problems:        contestProblems,
		Contestant:      contestants,
		LastUpdate:      time.Now(),
	}
	return &sb, nil
}

func (sdm *ScoreDbModel) GetProblemScoreByUser(contestId, problemId, userId int, publicScoreboard bool) (*gytypes.ScoreProblemData, error) {
	db := sdm.db
	selTable := "{{.TablePrefix}}scores_private"
	if publicScoreboard {
		selTable = "{{.TablePrefix}}scores_public"
	}
	query := `SELECT id_contest, id_problem, id_user, score, accepted_time, penalty_time, submission_count,
        one_hit, regraded FROM %s WHERE (id_contest = ?) AND (id_problem = ?) AND (id_user = ?)`
	query = fmt.Sprintf(query, selTable)
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var problem gytypes.ScoreProblemData
	err = stmt.QueryRow(contestId, problemId, userId).Scan(
		&problem.ContestId,
		&problem.ProblemId,
		&problem.UserId,
		&problem.Score,
		&problem.AcceptedTime,
		&problem.PenaltyTime,
		&problem.SubmissionCount,
		&problem.OneHit,
		&problem.Regraded,
	)
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

func (sdm *ScoreDbModel) SubmitIntoScoreboard(problemId, userId, score int, accepted bool) error {
	sci, err := sdm.GetContestInfoByProblemId(problemId)
	if err != nil {
		return err
	}
	err = sdm.processSubmitIntoScoreboard(sci, sci.ContestId, problemId, userId, score, accepted, false)
	if err != nil {
		return err
	}
	ft := sci.FreezeTime.Unix()
	// TODO: remove EnableFreeze as doesn't matter
	if ft > 0 {
		if time.Now().Unix() < ft {
			err = sdm.processSubmitIntoScoreboard(sci, sci.ContestId, problemId, userId, score, accepted, true)
		}
	} else {
		err = sdm.processSubmitIntoScoreboard(sci, sci.ContestId, problemId, userId, score, accepted, true)
	}
	return err
}

func (sdm *ScoreDbModel) processSubmitIntoScoreboard(sci gytypes.ScoreContestInfo, contestId, problemId, userId, score int, accepted, publicScoreboard bool) error {
	// Since ICPC doesn't matter what you write out
	if sci.Style == gytypes.ScoreStyleICPC {
		if score >= 100 {
			score = 1
		} else {
			score = 0
		}
	}
	// Fetch score data if exists
	if currentScore, err := sdm.GetProblemScoreByUser(contestId, problemId, userId, publicScoreboard); err == nil {
		// Score entry exists, just update that
		if sci.Style == gytypes.ScoreStyleICPC {
			// Ignore any submission if current problem was solved
			if currentScore.AcceptedTime > 0 {
				return nil
			}
		}
		currentScore.Score = score
		currentScore.SubmissionCount++
		if accepted {
			currentScore.AcceptedTime = time.Now().Unix()
		} else {
			currentScore.PenaltyTime += sci.PenaltyTime
		}
		// Update here
		if err = sdm.UpdateScore(currentScore, publicScoreboard); err != nil {
			return err
		}
	} else {
		// Seems new submission, so we insert as new entry
		spd := gytypes.ScoreProblemData{
			ContestId:       contestId,
			ProblemId:       problemId,
			UserId:          userId,
			Score:           score,
			AcceptedTime:    0,
			PenaltyTime:     0,
			SubmissionCount: 1,
			OneHit:          false,
			Regraded:        false,
		}
		if accepted {
			spd.AcceptedTime = time.Now().Unix()
		} else {
			spd.PenaltyTime = sci.PenaltyTime
		}
		if err = sdm.InsertScore(&spd, publicScoreboard); err != nil {
			return err
		}
	}
	return nil
}

func (sdm *ScoreDbModel) InsertScore(score *gytypes.ScoreProblemData, publicScoreboard bool) error {
	db := sdm.db
	selTable := "{{.TablePrefix}}scores_private"
	if publicScoreboard {
		selTable = "{{.TablePrefix}}scores_public"
	}
	query := `INSERT INTO %s 
        (id_contest, id_problem, id_user, score, accepted_time, penalty_time, submission_count, one_hit, regraded) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	query = fmt.Sprintf(query, selTable)
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		score.ContestId,
		score.ProblemId,
		score.UserId,
		score.Score,
		score.AcceptedTime,
		score.PenaltyTime,
		score.SubmissionCount,
		score.OneHit,
		score.Regraded,
	)
	return err
}

func (sdm *ScoreDbModel) UpdateScore(score *gytypes.ScoreProblemData, publicScoreboard bool) error {
	db := sdm.db
	selTable := "{{.TablePrefix}}scores_private"
	if publicScoreboard {
		selTable = "{{.TablePrefix}}scores_public"
	}
	query := `UPDATE %s SET
        score = ?,
        accepted_time = ?,
        penalty_time = ?,
        submission_count = ?,
        one_hit = ?,
        regraded = ?
        WHERE (id_contest = ?) AND (id_problem = ?) AND (id_user = ?)`
	query = fmt.Sprintf(query, selTable)
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		score.Score,
		score.AcceptedTime,
		score.PenaltyTime,
		score.SubmissionCount,
		score.OneHit,
		score.Regraded,
		score.ContestId,
		score.ProblemId,
		score.UserId,
	)
	return err
}
