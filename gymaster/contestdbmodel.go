package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
)

type ContestDbModel struct {
	db DbContext
}

func NewContestDbModel(db DbContext) ContestDbModel {
	cdm := ContestDbModel{
		db: db,
	}
	return cdm
}

func (cdm *ContestDbModel) CreateContest(contest gytypes.ContestData) error {
	/*db := cdm.db
	title := contest.Title
	if title == "" {
		title = "Untitled"
	}
	description := contest.Description
	style := contest.Style
	if style == "" {
		style = "ICPC"
	}
	allowedLang := contest.AllowedLang
	if allowedLang == "" {
		allowedLang = "1,2,4"
	}
	contestGroupId := 0
	enableFreeze := contest.EnableFreeze
	active := true
	allowPublic :=*/
	return nil
}

func (cdm *ContestDbModel) GetContestAccessCount(contestId int, userId int) (int, error) {
	db := cdm.db
	query := `SELECT COUNT(*) FROM {{.TablePrefix}}contest_access
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

func (cdm *ContestDbModel) GetContestAccessOfUserId(contestId int, userId int) (gytypes.ContestAccess, error) {
	db := cdm.db
	ca := gytypes.ContestAccess{}
	query := `SELECT id_user, id_contest, start_time, end_time, allowed FROM {{.TablePrefix}}contest_access
        WHERE id_user = ? AND id_contest = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return ca, err
	}
	defer stmt.Close()
	var utStartTime, utEndTime int64
	err = stmt.QueryRow(userId, contestId).Scan(
		&ca.UserId,
		&ca.ContestId,
		&utStartTime,
		&utEndTime,
		&ca.Allowed,
	)
	if err != nil {
		return ca, err
	}
	ca.StartTime = time.Unix(utStartTime, 0).Local()
	ca.EndTime = time.Unix(utEndTime, 0).Local()
	return ca, nil
}

func (cdm *ContestDbModel) GetContestList() ([]gytypes.ContestData, error) {
	db := cdm.db
	query := `SELECT id, title, description, style, allowed_lang, problem_count, contest_group_id, enable_freeze, active, allow_public, must_stream,
        start_timestamp, end_timestamp, freeze_timestamp, unfreeze_timestamp, max_runtime FROM {{.TablePrefix}}contests ORDER BY id DESC`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	var cl []gytypes.ContestData
	for rows.Next() {
		cd := gytypes.ContestData{}
		var utStartTime, utEndTime, utFreezeTime, utUnfreezeTime int64
		err = rows.Scan(
			&cd.Id,
			&cd.Title,
			&cd.Description,
			&cd.Style,
			&cd.AllowedLang,
			&cd.ProblemCount,
			&cd.GroupId,
			&cd.EnableFreeze,
			&cd.Active,
			&cd.PublicView,
			&cd.MustStream,
			&utStartTime,
			&utEndTime,
			&utFreezeTime,
			&utUnfreezeTime,
			&cd.MaxTime,
		)
		if err != nil {
			return nil, err
		}
		cd.StartTime = time.Unix(utStartTime, 0).Local()
		cd.EndTime = time.Unix(utEndTime, 0).Local()
		cd.FreezeTime = time.Unix(utFreezeTime, 0).Local()
		cd.UnfreezeTime = time.Unix(utUnfreezeTime, 0).Local()
		cd.ContestUrl = "dashboard/problemSet/" + strconv.Itoa(cd.Id)
		if (utStartTime != 0) && (utEndTime != 0) {
			maxTime := cd.MaxTime / 60
			cd.TimeDesc = template.HTML(fmt.Sprintf(`
			<p><table>
			<tr><td>Waktu mulai</td><td>: %s</td></tr>
			<tr><td>Waktu berakhir</td><td>: %s</td></tr>
			<tr><td>Jangka waktu</td><td>: %d menit</td></tr>
			</table>
			</p>
			`, cd.StartTime.Format(time.RFC3339), cd.EndTime.Format(time.RFC3339), maxTime))
		} else {
			cd.TimeDesc = "<p>Waktu: bisa dikerjakan kapanpun</p>"
		}
		cl = append(cl, cd)
	}
	return cl, nil
}

func (cdm *ContestDbModel) GetContestListOfUserId(uid int) ([]gytypes.ContestData, error) {
	ui := appUsers.GetUserById(uid)
	clBase, err := cdm.GetContestList()
	if err != nil {
		return nil, err
	}
	var cl []gytypes.ContestData
	for _, cd := range clBase {
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
				continue
			}
		}
		cl = append(cl, cd)
	}
	return cl, nil
}

func (cdm *ContestDbModel) GetContestListForScoreboard(publicScoreboard bool) ([]gytypes.ScoreboardListData, error) {
	var contestList []gytypes.ScoreboardListData
	db := cdm.db
	query := `SELECT c.id, c.title, (SELECT COUNT(*) FROM {{.TablePrefix}}contest_access as a WHERE a.id_contest = c.id), c.allow_public
        FROM {{.TablePrefix}}contests as c WHERE (c.allow_public = ?) OR (0 = ?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(publicScoreboard, gylib.Btoi(publicScoreboard))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		contest := gytypes.ScoreboardListData{}
		err = rows.Scan(
			&contest.ContestId,
			&contest.ContestName,
			&contest.ContestantCount,
			&contest.AllowPublic,
		)
		if err != nil {
			return nil, err
		}
		contest.Updated = false
		contestList = append(contestList, contest)
	}
	return contestList, nil
}

func (cdm *ContestDbModel) GetContestDetails(contestId int) (gytypes.ContestData, error) {
	cd := gytypes.ContestData{}
	db := cdm.db
	query := `SELECT id, title, description, style, allowed_lang, problem_count, contest_group_id, enable_freeze, active, allow_public, must_stream,
        start_timestamp, end_timestamp, freeze_timestamp, unfreeze_timestamp, max_runtime FROM {{.TablePrefix}}contests WHERE id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return cd, err
	}
	defer stmt.Close()
	var utStartTime, utEndTime, utFreezeTime, utUnfreezeTime int64
	err = stmt.QueryRow(contestId).Scan(
		&cd.Id,
		&cd.Title,
		&cd.Description,
		&cd.Style,
		&cd.AllowedLang,
		&cd.ProblemCount,
		&cd.GroupId,
		&cd.EnableFreeze,
		&cd.Active,
		&cd.PublicView,
		&cd.MustStream,
		&utStartTime,
		&utEndTime,
		&utFreezeTime,
		&utUnfreezeTime,
		&cd.MaxTime,
	)
	if err != nil {
		return cd, err
	}
	cd.StartTime = time.Unix(utStartTime, 0).Local()
	cd.EndTime = time.Unix(utEndTime, 0).Local()
	cd.FreezeTime = time.Unix(utFreezeTime, 0).Local()
	cd.UnfreezeTime = time.Unix(utUnfreezeTime, 0).Local()
	cd.ContestUrl = "dashboard/problemSet/" + strconv.Itoa(cd.Id)
	if (utStartTime != 0) && (utEndTime != 0) {
		maxTime := cd.MaxTime / 60
		cd.TimeDesc = template.HTML(fmt.Sprintf(`
			<p><table>
			<tr><td>Waktu mulai</td><td>: %s</td></tr>
			<tr><td>Waktu berakhir</td><td>: %s</td></tr>
			<tr><td>Jangka waktu</td><td>: %d menit</td></tr>
			</table>
			</p>
			`, cd.StartTime.Format(time.RFC3339), cd.EndTime.Format(time.RFC3339), maxTime))
	} else {
		cd.TimeDesc = "<p>Waktu: bisa dikerjakan kapanpun</p>"
	}
	return cd, nil
}

func (cdm *ContestDbModel) GetProblemSet(contestId int) ([]gytypes.ProblemData, error) {
	var qs []gytypes.ProblemData
	db := cdm.db
	query := `SELECT id, contest_id, problem_name, problem_shortname, description, time_limit, mem_limit, max_attempts
		FROM {{.TablePrefix}}problems WHERE contest_id = ? ORDER BY problem_name ASC`
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
		qd := gytypes.ProblemData{}
		err = rows.Scan(
			&qd.Id,
			&qd.ContestId,
			&qd.Name,
			&qd.ShortName,
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

func (cdm *ContestDbModel) GetProblemById(problemId int) (gytypes.ProblemData, error) {
	qd := gytypes.ProblemData{}
	db := cdm.db
	query := `SELECT p.id, p.contest_id, p.problem_name, p.problem_shortname, p.description, p.time_limit, p.mem_limit, p.max_attempts, c.allowed_lang
        FROM {{.TablePrefix}}problems AS p INNER JOIN {{.TablePrefix}}contests AS c ON p.contest_id = c.id WHERE p.id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return qd, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(problemId).Scan(
		&qd.Id,
		&qd.ContestId,
		&qd.Name,
		&qd.ShortName,
		&qd.Description,
		&qd.TimeLimit,
		&qd.MemLimit,
		&qd.MaxAttempts,
		&qd.AllowedLang,
	)
	if err != nil {
		return qd, err
	}
	qd.ContestUrl = "dashboard/problemSet/" + strconv.Itoa(qd.ContestId)
	return qd, nil
}

func (cdm *ContestDbModel) InsertContestAccess(access gytypes.ContestAccess) error {
	db := cdm.db
	query := `INSERT INTO {{.TablePrefix}}contest_access (id_user, id_contest, start_time, end_time, allowed)
        VALUES (?, ?, ?, ?, ?)`
	prep, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer prep.Close()
	utStartTime := access.StartTime.Unix()
	utEndTime := access.EndTime.Unix()
	_, err = prep.Exec(
		access.UserId,
		access.ContestId,
		utStartTime,
		utEndTime,
		access.Allowed,
	)
	if err != nil {
		return err
	}
	return nil
}
