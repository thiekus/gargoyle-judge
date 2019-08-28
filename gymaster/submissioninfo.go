package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

type SubmissionInfo struct {
	Id              int
	ProblemId       int
	UserId          int
	CacheId         string
	Language        string
	Code            string
	Status          string
	Details         string
	Score           int
	SubmitTime      int64
	CompileTime     float64
	CompileStdout   string
	CompileStderr   string
	UserDisplayName string
	ProblemName     string
	ContestName     string
}

// Based from ICPC rules guide
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

func (si *SubmissionInfo) GetStatusMessage() string {
	return TranslateSubmissionCode(si.Status)
}

func (si *SubmissionInfo) IsSuccess() bool {
	return si.Status == SubmissionAccepted
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
