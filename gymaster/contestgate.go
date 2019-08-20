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
	"net/http"
	"strconv"
	"time"
)

type ContestData struct {
	Id          int
	ContestType string
	Title       string
	Description template.HTML
	TimeDesc    template.HTML
	ContestUrl  string
	QuestCount  int
	Unlocked    bool
	Private     bool
	Trainer     bool
	StartTime   int
	EndTime     int
	MaxTime     int
}

type QuestionData struct {
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
	Count     int
	Contest   ContestData
	Questions []QuestionData
}

func fetchContestList(r *http.Request, trainer bool) ContestList {
	cl := ContestList{Count: 0}
	db, err := OpenDatabaseEx(false)
	if err != nil {
		return cl
	}
	defer db.Close()
	query := `SELECT id, title, description, quest_count, is_unlocked, is_private, is_trainer,
        start_timestamp, end_timestamp, max_runtime FROM %TABLEPREFIX%contests WHERE is_trainer = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return cl
	}
	defer stmt.Close()
	trainInt := 0
	if trainer {
		trainInt = 1
	}
	rows, err := stmt.Query(trainInt)
	if err != nil {
		return cl
	}
	for rows.Next() {
		cd := ContestData{}
		var unlockedInt, privateInt, trainerInt bool
		err = rows.Scan(
			&cd.Id,
			&cd.Title,
			&cd.Description,
			&cd.QuestCount,
			&unlockedInt,
			&privateInt,
			&trainerInt,
			&cd.StartTime,
			&cd.EndTime,
			&cd.MaxTime,
		)
		if err != nil {
			return cl
		}
		cd.Unlocked = unlockedInt
		cd.Private = privateInt
		cd.Trainer = trainerInt
		if !cd.Trainer {
			cd.ContestType = "kontes"
		} else {
			cd.ContestType = "latihan"
		}
		cd.ContestUrl = getBaseUrl(r) + "dashboard/problem/" + strconv.Itoa(cd.Id)
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
	return cl
}

func fetchProblem(r *http.Request, contestId int) (ContestData, error) {
	cd := ContestData{}
	db, err := OpenDatabaseEx(false)
	if err != nil {
		return cd, err
	}
	defer db.Close()
	query := `SELECT id, title, description, quest_count, is_unlocked, is_private, is_trainer,
        start_timestamp, end_timestamp, max_runtime FROM %TABLEPREFIX%contests WHERE id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return cd, err
	}
	defer stmt.Close()
	var unlockedInt, privateInt, trainerInt bool
	err = stmt.QueryRow(contestId).Scan(
		&cd.Id,
		&cd.Title,
		&cd.Description,
		&cd.QuestCount,
		&unlockedInt,
		&privateInt,
		&trainerInt,
		&cd.StartTime,
		&cd.EndTime,
		&cd.MaxTime,
	)
	if err != nil {
		return cd, err
	}
	cd.Unlocked = unlockedInt
	cd.Private = privateInt
	cd.Trainer = trainerInt
	if !cd.Trainer {
		cd.ContestType = "kontes"
	} else {
		cd.ContestType = "latihan"
	}
	cd.ContestUrl = getBaseUrl(r) + "dashboard/problem/" + strconv.Itoa(cd.Id)
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

func fetchQuestionList(r *http.Request, contestId int) ([]QuestionData, error) {
	var qs []QuestionData
	db, err := OpenDatabaseEx(false)
	if err != nil {
		return qs, err
	}
	defer db.Close()
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
		qd := QuestionData{}
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
		qd.QuestUrl = getBaseUrl(r) + "dashboard/question/" + strconv.Itoa(qd.Id)
		qs = append(qs, qd)
	}
	return qs, nil
}

func fetchQuestion(r *http.Request, id int) (QuestionData, error) {
	qd := QuestionData{}
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
	qd.ContestUrl = getBaseUrl(r) + "dashboard/problem/" + strconv.Itoa(qd.ContestId)
	return qd, nil
}
