package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"strings"
	"time"

	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
)

type SubmissionDbModel struct {
	db DbContext
}

func NewSubmissionDbModel(db DbContext) SubmissionDbModel {
	sdm := SubmissionDbModel{
		db: db,
	}
	return sdm
}

func (sdm *SubmissionDbModel) GetSubmission(subId int) (gytypes.SubmissionData, error) {
	si := gytypes.SubmissionData{}
	db := sdm.db
	query := `SELECT s.id, s.id_problem, s.id_user, s.id_lang, s.code, s.verdict, s.details, s.score, s.submit_time, s.compile_time,
        s.compile_stdout, s.compile_stderr, u.display_name, p.problem_name, c.title
        FROM ((({{.TablePrefix}}submissions AS s INNER JOIN {{.TablePrefix}}users AS u ON s.id_user = u.id)
        INNER JOIN {{.TablePrefix}}problems AS p ON s.id_problem = p.id)
        INNER JOIN {{.TablePrefix}}contests AS c ON p.contest_id = c.id)
        WHERE s.id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return si, err
	}
	defer stmt.Close()
	var utSubmitTime int64
	err = stmt.QueryRow(subId).Scan(
		&si.Id,
		&si.ProblemId,
		&si.UserId,
		&si.LanguageId,
		&si.Code,
		&si.Verdict,
		&si.Details,
		&si.Score,
		&utSubmitTime,
		&si.CompileTime,
		&si.CompileStdout,
		&si.CompileStderr,
		&si.UserDisplayName,
		&si.ProblemName,
		&si.ContestName,
	)
	if err != nil {
		return si, err
	}
	si.SubmitTime = time.Unix(utSubmitTime, 0)
	lang, err := appLangPrograms.GetLanguageFromId(si.LanguageId)
	if err != nil {
		return si, err
	}
	si.LanguageName = lang.DisplayName
	si.LanguageSyntax = lang.SyntaxName
	return si, nil
}

func (sdm *SubmissionDbModel) GetSubmissionCount(userId int, problemId int) (int, error) {
	db := sdm.db
	query := `SELECT COUNT(*) FROM {{.TablePrefix}}submissions WHERE ((s.id_user = ?) OR (0 = ?)) AND ((s.id_problem = ?) OR (0 = ?))`
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRow(userId, userId, problemId, problemId).Scan(&count)
	return count, err
}

func (sdm *SubmissionDbModel) GetSubmissionList(userId int, problemId int) ([]gytypes.SubmissionData, error) {
	db := sdm.db
	query := `SELECT s.id, s.id_problem, s.id_user, s.id_lang, s.code, s.verdict, s.details, s.score, s.submit_time, s.compile_time,
        s.compile_stdout, s.compile_stderr, u.display_name, p.problem_name, c.title
        FROM ((({{.TablePrefix}}submissions AS s INNER JOIN {{.TablePrefix}}users AS u ON s.id_user = u.id)
        INNER JOIN {{.TablePrefix}}problems AS p ON s.id_problem = p.id)
        INNER JOIN {{.TablePrefix}}contests AS c ON p.contest_id = c.id)
        WHERE ((s.id_user = ?) OR (0 = ?)) AND ((s.id_problem = ?) OR (0 = ?)) ORDER BY s.id DESC`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(userId, userId, problemId, problemId)
	if err != nil {
		return nil, err
	}
	var subs []gytypes.SubmissionData
	for rows.Next() {
		sb := gytypes.SubmissionData{}
		var utSubmitTime int64
		err = rows.Scan(
			&sb.Id,
			&sb.ProblemId,
			&sb.UserId,
			&sb.LanguageId,
			&sb.Code,
			&sb.Verdict,
			&sb.Details,
			&sb.Score,
			&utSubmitTime,
			&sb.CompileTime,
			&sb.CompileStdout,
			&sb.CompileStderr,
			&sb.UserDisplayName,
			&sb.ProblemName,
			&sb.ContestName,
		)
		sb.SubmitTime = time.Unix(utSubmitTime, 0)
		lang, err := appLangPrograms.GetLanguageFromId(sb.LanguageId)
		if err != nil {
			return subs, err
		}
		sb.LanguageName = lang.DisplayName
		sb.LanguageSyntax = lang.SyntaxName
		subs = append(subs, sb)
	}
	return subs, nil
}

func (sdm *SubmissionDbModel) GetTestCasesOfProblem(problemId int) ([]gytypes.TestCaseData, error) {
	db := sdm.db
	query := `SELECT id, id_problem, test_no, input, output FROM {{.TablePrefix}}testcases
       WHERE id_problem = ? ORDER BY test_no ASC`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(problemId)
	if err != nil {
		return nil, err
	}
	var cases []gytypes.TestCaseData
	for rows.Next() {
		testcase := gytypes.TestCaseData{}
		err = rows.Scan(
			&testcase.Id,
			&testcase.ProblemId,
			&testcase.TestNo,
			&testcase.Input,
			&testcase.Output,
		)
		if err != nil {
			return nil, err
		}
		cases = append(cases, testcase)
	}
	return cases, nil
}

func (sdm *SubmissionDbModel) GetTestResultOfSubmission(submissionId int) ([]gytypes.TestResultData, error) {
	db := sdm.db
	query := `SELECT id, id_problem, id_submission, test_no, verdict, time_elapsed, memory_used, score FROM
       {{.TablePrefix}}testresults WHERE id_submission = ? ORDER BY test_no ASC`
	prep, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer prep.Close()
	rows, err := prep.Query(submissionId)
	if err != nil {
		return nil, err
	}
	var trl []gytypes.TestResultData
	for rows.Next() {
		tr := gytypes.TestResultData{}
		err = rows.Scan(
			&tr.Id,
			&tr.ProblemId,
			&tr.SubmissionId,
			&tr.TestNo,
			&tr.Verdict,
			&tr.TimeElapsed,
			&tr.MemoryUsed,
			&tr.Score,
		)
		if err != nil {
			return nil, err
		}
		trl = append(trl, tr)
	}
	return trl, nil
}

func (sdm *SubmissionDbModel) InsertSubmissionOnQueue(idProblem, idUser, idLang int, code string) (int, error) {
	db := sdm.db
	query := `INSERT INTO {{.TablePrefix}}submissions (id_problem, id_user, id_lang, code, verdict, details, submit_time, compile_stdout, compile_stderr)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	prep, err := db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer prep.Close()
	now := time.Now().Unix()
	res, err := prep.Exec(
		idProblem,
		idUser,
		idLang,
		code,
		gytypes.SubmissionOnQueue,
		"",
		now,
		"",
		"",
	)
	if err != nil {
		return 0, err
	}
	idx, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(idx), nil
}

func (sdm *SubmissionDbModel) InsertTestResult(testResult gytypes.TestResultData) error {
	db := sdm.db
	query := `INSERT INTO {{.TablePrefix}}testresults (id_problem, id_submission, test_no, verdict, time_elapsed, memory_used, score)
        VALUES (?, ?, ?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		testResult.ProblemId,
		testResult.SubmissionId,
		testResult.TestNo,
		testResult.Verdict,
		testResult.TimeElapsed,
		testResult.MemoryUsed,
		testResult.Score,
	)
	return err
}

func sanitizePath(log string) string {
	progDir := gylib.GetProgramBaseDir()
	cachesDir := progDir + "/lib/caches"
	log = strings.ReplaceAll(log, cachesDir, "/tmp")
	log = strings.ReplaceAll(log, progDir, "/opt/gargoyle")
	return log
}

func (sdm *SubmissionDbModel) UpdateSubmission(id int, submission gytypes.SubmissionData) error {
	db := sdm.db
	query := `UPDATE {{.TablePrefix}}submissions SET
        verdict = ?,
        details = ?,
        score = ?,
        compile_time = ?,
        compile_stdout = ?,
        compile_stderr = ?
        WHERE id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		submission.Verdict,
		submission.Details,
		submission.Score,
		submission.CompileTime,
		sanitizePath(submission.CompileStdout),
		sanitizePath(submission.CompileStderr),
		id,
	)
	return err
}
