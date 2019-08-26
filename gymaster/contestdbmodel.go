package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"fmt"
	"html/template"
	"strconv"
	"time"
)

type ContestDbModel struct {
	db DbContext
}

func NewContestDbModel() (ContestDbModel, error) {
	cdm := ContestDbModel{}
	db, err := OpenDatabase()
	if err != nil {
		return cdm, err
	}
	cdm.db = db
	return cdm, err
}

func (cdm *ContestDbModel) Close() error {
	return cdm.db.Close()
}

func (cdm *ContestDbModel) GetContestAccessCount(contestId int, userId int) (int, error) {
	db := cdm.db
	query := `SELECT COUNT(*) FROM %TABLEPREFIX%contest_access
        WHERE id_user = ? AND id_contest = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRow(userId, contestId).Scan(&count)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func (cdm *ContestDbModel) GetContestAccessOfUserId(contestId int, userId int) (ContestAccess, error) {
	db := cdm.db
	ca := ContestAccess{}
	query := `SELECT id, id_user, id_contest, start_time, end_time, allowed FROM %TABLEPREFIX%contest_access
        WHERE id_user = ? AND id_contest = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return ca, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(userId, contestId).Scan(
		&ca.Id,
		&ca.UserId,
		&ca.ContestId,
		&ca.StartTime,
		&ca.EndTime,
		&ca.Allowed,
	)
	if err != nil {
		return ca, err
	}
	return ca, nil
}

func (cdm *ContestDbModel) GetContestListOfUserId(uid int, trainer bool) (ContestList, error) {
	ui := appUsers.GetUserById(uid)
	cl := ContestList{Count: 0}
	db := cdm.db
	query := `SELECT id, title, description, problem_count, contest_group_id, is_unlocked, is_public, must_stream, 
        start_timestamp, end_timestamp, max_runtime FROM %TABLEPREFIX%contests WHERE is_trainer = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return cl, err
	}
	defer stmt.Close()
	trainInt := 0
	if trainer {
		trainInt = 1
	}
	rows, err := stmt.Query(trainInt)
	if err != nil {
		return cl, err
	}
	for rows.Next() {
		cd := ContestData{}
		var unlockedInt, publicInt, mustStreamInt bool
		err = rows.Scan(
			&cd.Id,
			&cd.Title,
			&cd.Description,
			&cd.ProblemCount,
			&cd.GroupId,
			&unlockedInt,
			&publicInt,
			&mustStreamInt,
			&cd.StartTime,
			&cd.EndTime,
			&cd.MaxTime,
		)
		if err != nil {
			return cl, err
		}
		// Check group access if contest is restricted to certain group only
		if cd.GroupId != 0 {
			granted := false
			for _, gr := range ui.Groups {
				if gr.GroupId == cd.GroupId {
					granted = true
					break
				}
			}
			if !granted {
				// skip this contest
				break
			}
		}
		cd.Unlocked = unlockedInt
		cd.PublicView = publicInt
		cd.MustStream = mustStreamInt
		cd.ContestUrl = "dashboard/problemSet/" + strconv.Itoa(cd.Id)
		if (cd.StartTime != 0) && (cd.EndTime != 0) {
			utStart := time.Unix(int64(cd.StartTime), 0)
			utEnd := time.Unix(int64(cd.EndTime), 0)
			maxTime := cd.MaxTime / 60
			cd.TimeDesc = template.HTML(fmt.Sprintf(`
			<p><table>
			<tr><td>Waktu mulai</td><td>: %s</td></tr>
			<tr><td>Waktu berakhir</td><td>: %s</td></tr>
			<tr><td>Jangka waktu</td><td>: %d menit</td></tr>
			</table>
			</p>
			`, utStart.Format(time.RFC3339), utEnd.Format(time.RFC3339), maxTime))
		} else {
			cd.TimeDesc = "<p>Waktu: bisa dikerjakan kapanpun</p>"
		}
		cl.Contests = append(cl.Contests, cd)
		cl.Count++
	}
	return cl, nil
}

func (cdm *ContestDbModel) GetContestDetails(contestId int) (ContestData, error) {
	cd := ContestData{}
	db := cdm.db
	query := `SELECT id, title, description, problem_count, contest_group_id, is_unlocked, is_public, must_stream, 
        start_timestamp, end_timestamp, max_runtime FROM %TABLEPREFIX%contests WHERE id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return cd, err
	}
	defer stmt.Close()
	var unlockedInt, publicInt, mustStreamInt bool
	err = stmt.QueryRow(contestId).Scan(
		&cd.Id,
		&cd.Title,
		&cd.Description,
		&cd.ProblemCount,
		&cd.GroupId,
		&unlockedInt,
		&publicInt,
		&mustStreamInt,
		&cd.StartTime,
		&cd.EndTime,
		&cd.MaxTime,
	)
	if err != nil {
		return cd, err
	}
	cd.Unlocked = unlockedInt
	cd.PublicView = publicInt
	cd.ContestUrl = "dashboard/problemSet/" + strconv.Itoa(cd.Id)
	if (cd.StartTime != 0) && (cd.EndTime != 0) {
		utStart := time.Unix(int64(cd.StartTime), 0)
		utEnd := time.Unix(int64(cd.EndTime), 0)
		maxTime := cd.MaxTime / 60
		cd.TimeDesc = template.HTML(fmt.Sprintf(`
			<p><table>
			<tr><td>Waktu mulai</td><td>: %s</td></tr>
			<tr><td>Waktu berakhir</td><td>: %s</td></tr>
			<tr><td>Jangka waktu</td><td>: %d menit</td></tr>
			</table>
			</p>
			`, utStart.Format(time.RFC3339), utEnd.Format(time.RFC3339), maxTime))
	} else {
		cd.TimeDesc = "<p>Waktu: bisa dikerjakan kapanpun</p>"
	}
	return cd, nil
}

func (cdm *ContestDbModel) GetProblemSet(contestId int) ([]ProblemData, error) {
	var qs []ProblemData
	db := cdm.db
	query := `SELECT id, contest_id, problem_name, description, time_limit, mem_limit, max_attempts
		FROM %TABLEPREFIX%problems WHERE contest_id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return qs, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(contestId)
	if err != nil {
		return qs, err
	}
	for rows.Next() {
		qd := ProblemData{}
		err = rows.Scan(
			&qd.Id,
			&qd.ContestId,
			&qd.Name,
			&qd.Description,
			&qd.TimeLimit,
			&qd.MemLimit,
			&qd.MaxAttempts,
		)
		if err != nil {
			return qs, err
		}
		qd.ProblemUrl = "dashboard/problem/" + strconv.Itoa(qd.Id)
		qs = append(qs, qd)
	}
	return qs, nil
}

func (cdm *ContestDbModel) GetProblemById(problemId int) (ProblemData, error) {
	qd := ProblemData{}
	db := cdm.db
	query := `SELECT id, contest_id, problem_name, description, time_limit, mem_limit, max_attempts FROM %TABLEPREFIX%problems
    	WHERE id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return qd, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(problemId).Scan(
		&qd.Id,
		&qd.ContestId,
		&qd.Name,
		&qd.Description,
		&qd.TimeLimit,
		&qd.MemLimit,
		&qd.MaxAttempts,
	)
	if err != nil {
		return qd, err
	}
	qd.ContestUrl = "dashboard/problemSet/" + strconv.Itoa(qd.ContestId)
	return qd, nil
}

func (cdm *ContestDbModel) InsertContestAccess(access ContestAccess) error {
	db := cdm.db
	query := `INSERT INTO %TABLEPREFIX%contest_access (id_user, id_contest, start_time, end_time, allowed) 
        VALUES (?, ?, ?, ?, ?)`
	prep, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer prep.Close()
	_, err = prep.Exec(
		access.UserId,
		access.ContestId,
		access.StartTime,
		access.EndTime,
		access.Allowed,
	)
	if err != nil {
		return err
	}
	return nil
}
