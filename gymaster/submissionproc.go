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
	"strconv"

	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gyrpc"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
)

type SubmissionProcessor struct {
	slaveMan     *SlaveManager
	problem      gytypes.ProblemData
	submissionId int
	problemId    int
	userId       int
	langId       int
	code         string
}

func NewSubmissionProcessor(slaveMan *SlaveManager, idProblem, idUser, idLang int, code string) (*SubmissionProcessor, error) {
	db, err := OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	cdm := NewContestDbModel(db)
	problem, err := cdm.GetProblemById(idProblem)
	if err != nil {
		return nil, err
	}
	// Check access to avoid access bypass
	ca, err := appContestAccess.GetAccessInfoOfUser(idUser, problem.ContestId)
	if err != nil {
		return nil, err
	}
	err = appContestAccess.CheckAccessInfo(ca)
	if err != nil {
		return nil, err
	}
	sdm := NewSubmissionDbModel(db)
	// Check if problem have maximum attempts
	if problem.MaxAttempts > 0 {
		count, err := sdm.GetSubmissionCount(idUser, idProblem)
		if err != nil {
			return nil, err
		}
		// Disallow if reached or greater
		if count >= problem.MaxAttempts {
			return nil, errors.New("maximum attempt count reached")
		}
	}
	sp := SubmissionProcessor{
		slaveMan:     slaveMan,
		submissionId: 0,
		problem:      problem,
		problemId:    idProblem,
		userId:       idUser,
		langId:       idLang,
		code:         code,
	}
	return &sp, nil
}

func (sp *SubmissionProcessor) processSubmission(client *gyrpc.GargoyleRpcClient, db *DbContext) {
	defer client.Close()
	defer db.Close()
	sdm := NewSubmissionDbModel(*db)
	log := gylib.GetStdLog()
	sub := gytypes.SubmissionData{
		Id:      sp.submissionId,
		Verdict: gytypes.SubmissionError,
		Score:   0,
	}
	// deferred conclusion of this process
	defer func() {
		if sub.Details == "" {
			sub.Details = sub.GetStatusMessage()
		} else {
			sub.Details = sub.GetStatusMessage() + ": " + sub.Details
		}
		log.Printf("Submission [id:%d, verdict:%s]: %s", sub.Id, sub.Verdict, sub.Details)
		// Since db not yet closed, we still can doing db operations
		// Update score
		var err error = nil
		if sub.Verdict == gytypes.SubmissionAccepted {
			err = appScoreboard.SubmitScore(sub.ProblemId, sub.UserId, sub.Score, true)
		} else if (sub.Verdict != gytypes.SubmissionCompilerError) && (sub.Verdict != gytypes.SubmissionOnQueue) &&
			(sub.Verdict != gytypes.SubmissionError) {
			err = appScoreboard.SubmitScore(sub.ProblemId, sub.UserId, sub.Score, false)
		}
		if err != nil {
			log.Errorf("Error while updating score for id %d", sub.Id)
			return
		}
		if err = sdm.UpdateSubmission(sub.Id, sub); err != nil {
			log.Errorf("Error while updating submission for id %d", sub.Id)
			return
		}
		desc := fmt.Sprintf("Your last submission graded as %s (%s)", sub.Verdict, sub.GetStatusMessage())
		link := "/dashboard/userViewSubmission/" + strconv.Itoa(sub.Id)
		if err = appNotifications.AddNotification(sub.UserId, 0, desc, link); err != nil {
			log.Errorf("Error while updating notification for id %d", sub.Id)
			return
		}
	}()
	sbTemp, err := sdm.GetSubmission(sp.submissionId)
	if err != nil {
		sub.Details = err.Error()
		return
	}
	// Replace with gathered sub variable
	sub = sbTemp
	sub.Verdict = gytypes.SubmissionError
	// Get Programming language data
	lang, err := appLangPrograms.GetLanguageFromId(sub.LanguageId)
	if err != nil {
		sub.Details = err.Error()
		return
	}
	// Get Problem details
	cdm := NewContestDbModel(*db)
	prob, err := cdm.GetProblemById(sub.ProblemId)
	if err != nil {
		sub.Details = err.Error()
		return
	}
	// Retrieve test case for current problem for checking
	tests, err := sdm.GetTestCasesOfProblem(sp.problemId)
	if err != nil {
		sub.Details = err.Error()
		return
	}
	// Hit the slave to check it
	resp, err := client.ProcessSubmission(sub, *lang, prob, tests)
	if err != nil {
		sub.Details = err.Error()
		return
	}
	// Insert new rows for test case run result
	for _, testCase := range resp.TestResults {
		if err = sdm.InsertTestResult(testCase); err == nil {
			log.Printf("Inserting %v", testCase)
		} else {
			log.Errorf("TestCase insert error: %s", err.Error())
		}
	}
	sub = resp.Submission
}

func (sp *SubmissionProcessor) DoProcess() error {
	sl, err := sp.slaveMan.GetActiveSlave()
	if err != nil {
		return err
	}
	client, err := gyrpc.NewGargoyleRpcClient(sl.Address)
	if err != nil {
		return err
	}
	// Initialize Database connection and Submission database model
	db, err := OpenDatabase()
	if err != nil {
		client.Close()
		return err
	}
	sdm := NewSubmissionDbModel(db)
	// Insert into submission queue
	subId, err := sdm.InsertSubmissionOnQueue(sp.problemId, sp.userId, sp.langId, sp.code)
	if err != nil {
		db.Close()
		client.Close()
		return err
	}
	sp.submissionId = subId
	go sp.processSubmission(client, &db)
	return nil
}
