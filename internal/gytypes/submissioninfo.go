package gytypes

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"time"
)

type SubmissionData struct {
	Id            int
	ProblemId     int
	UserId        int
	LanguageId    int
	Code          string
	Verdict       string
	Details       string
	Score         int
	SubmitTime    time.Time
	CompileTime   float64
	CompileStdout string
	CompileStderr string
	// retrieved from another tables
	UserDisplayName string
	ProblemName     string
	ContestName     string
	LanguageName    string
	LanguageSyntax  string
}

type TestCaseData struct {
	Id        int
	ProblemId int
	TestNo    int
	Input     string
	Output    string
}

type TestResultData struct {
	Id           int
	ProblemId    int
	SubmissionId int
	TestNo       int
	Verdict      string
	TimeElapsed  float64
	MemoryUsed   uint64
	Score        float64
}

// Based from ACM-ICPC rules guide
// https://icpcarchive.ecs.baylor.edu/index.php?option=com_content&task=view&id=14&Itemid=30
const (
	SubmissionOnQueue             = "QU"
	SubmissionAccepted            = "AC"
	SubmissionPresentationError   = "PE"
	SubmissionWrongAnswer         = "WA"
	SubmissionCompilerError       = "CE"
	SubmissionRuntimeError        = "RE"
	SubmissionTimeLimitExceeded   = "TL"
	SubmissionMemoryLimitExceeded = "ML"
	SubmissionOutputLimitExceeded = "OL"
	SubmissionError               = "SE"
	SubmissionRestrictedFunction  = "RF"
	SubmissionCantJudged          = "CJ"
)

func (si *SubmissionData) GetStatusMessage() string {
	return TranslateSubmissionCode(si.Verdict)
}

func (si *SubmissionData) IsSuccess() bool {
	return si.Verdict == SubmissionAccepted
}

func TranslateSubmissionCode(code string) string {
	switch code {
	case SubmissionOnQueue:
		return "On Queue"
	case SubmissionAccepted:
		return "Accepted"
	case SubmissionPresentationError:
		return "Presentation Error"
	case SubmissionWrongAnswer:
		return "Wrong Answer"
	case SubmissionCompilerError:
		return "Compiler Error"
	case SubmissionRuntimeError:
		return "Runtime Error"
	case SubmissionTimeLimitExceeded:
		return "Time Limit Exceeded"
	case SubmissionMemoryLimitExceeded:
		return "Memory Limit Exceeded"
	case SubmissionOutputLimitExceeded:
		return "Output Limit Exceeded"
	case SubmissionRestrictedFunction:
		return "Restricted Function"
	case SubmissionCantJudged:
		return "Can't be Judged"
	}
	return SubmissionError
}
