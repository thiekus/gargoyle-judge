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

func (cdm *ContestDbModel) GetContestListByUserId(uid int, trainer bool) (ContestList, error) {
	ui := appUsers.GetUserById(uid)
	cl := ContestList{Count: 0}
	db := cdm.db
	query := `SELECT id, title, description, quest_count, contest_group_id, is_unlocked, is_public, is_trainer,
        must_stream, start_timestamp, end_timestamp, max_runtime FROM %TABLEPREFIX%contests WHERE is_trainer = ?`
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
		var unlockedInt, publicInt, trainerInt, mustStreamInt bool
		err = rows.Scan(
			&cd.Id,
			&cd.Title,
			&cd.Description,
			&cd.QuestCount,
			&cd.GroupId,
			&unlockedInt,
			&publicInt,
			&trainerInt,
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
		cd.Trainer = trainerInt
		cd.MustStream = mustStreamInt
		if !cd.Trainer {
			cd.ContestType = "kontes"
		} else {
			cd.ContestType = "latihan"
		}
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
	query := `SELECT id, title, description, quest_count, contest_group_id, is_unlocked, is_public, is_trainer,
        must_stream, start_timestamp, end_timestamp, max_runtime FROM %TABLEPREFIX%contests WHERE id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return cd, err
	}
	defer stmt.Close()
	var unlockedInt, publicInt, trainerInt, mustStreamInt bool
	err = stmt.QueryRow(contestId).Scan(
		&cd.Id,
		&cd.Title,
		&cd.Description,
		&cd.QuestCount,
		&cd.GroupId,
		&unlockedInt,
		&publicInt,
		&trainerInt,
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
	cd.Trainer = trainerInt
	cd.MustStream = mustStreamInt
	if !cd.Trainer {
		cd.ContestType = "kontes"
	} else {
		cd.ContestType = "latihan"
	}
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

func (cdm *ContestDbModel) GetQuestionList(contestId int) ([]ProblemData, error) {
	var qs []ProblemData
	db := cdm.db
	queryQuest := `SELECT id, contest_id, quest_name, description, time_limit, mem_limit, max_attempts
		FROM %TABLEPREFIX%quests WHERE contest_id = ?`
	stmt, err := db.Prepare(queryQuest)
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
		qd.QuestUrl = "dashboard/problem/" + strconv.Itoa(qd.Id)
		qs = append(qs, qd)
	}
	return qs, nil
}

func (cdm *ContestDbModel) GetQuestionById(id int) (ProblemData, error) {
	qd := ProblemData{}
	db, err := OpenDatabaseEx(false)
	if err != nil {
		return qd, err
	}
	defer db.Close()
	query := `SELECT id, contest_id, quest_name, description, time_limit, mem_limit, max_attempts FROM %TABLEPREFIX%quests
    	WHERE id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return qd, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(id).Scan(
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
