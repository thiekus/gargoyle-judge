package gyrpc

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
)

// Default Request Header
type RpcDefaultRequest struct {
	//AuthKey string
}

// Default Response Header
type RpcDefaultResponse struct {
	//AuthKey string
}

type RpcPingRequest struct {
	RpcDefaultRequest
	StartTime int64
}

type RpcPingResponse struct {
	RpcDefaultResponse
	StartTime int64
	EndTime   int64
	Delta     int64
}

type RpcSubmissionRequest struct {
	RpcDefaultRequest
	Submission     gytypes.SubmissionData
	ProgramLang    gytypes.LanguageProgramData
	ProblemDetails gytypes.ProblemData
	TestCases      []gytypes.TestCaseData
}

type RpcSubmissionResponse struct {
	RpcDefaultResponse
	Submission  gytypes.SubmissionData
	TestResults []gytypes.TestResultData
}
