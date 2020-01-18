package gyrpc

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"math"
	"regexp"
	"strings"
	"time"
)

type GargoyleRpcTask struct {
	parentServer *GargoyleRpcServer
}

func (grt *GargoyleRpcTask) PingSlave(req RpcPingRequest, resp *RpcPingResponse) error {
	log := gylib.GetStdLog()
	et := time.Now().UnixNano()
	resp.StartTime = req.StartTime
	resp.EndTime = et
	resp.Delta = et - req.StartTime
	delta := float64(resp.Delta) / float64(time.Millisecond)
	log.Printf("Ping from master, delta %fms", delta)
	return nil
}

func (grt *GargoyleRpcTask) ProcessSubmission(req RpcSubmissionRequest, resp *RpcSubmissionResponse) error {
	log := gylib.GetStdLog()
	sub := req.Submission
	lang := req.ProgramLang
	prob := req.ProblemDetails
	// Unless not assert or fatal error, RPC Slave Server should not return any error, but varied by verdict
	// Slave server supposed to have that handlers!
	parent := grt.parentServer
	log.Printf("[subId:%d] Begin processing submission...", sub.Id)
	// If defined (to fix java class name), do regex
	code := sub.Code
	if lang.RegexReplaceFrom != "" {
		log.Printf("[subId:%d] Regex replace from '%s' to '%s'", sub.Id, lang.RegexReplaceFrom, lang.RegexReplaceTo)
		rep := regexp.MustCompile(lang.RegexReplaceFrom)
		code = string(rep.ReplaceAll([]byte(code), []byte(lang.RegexReplaceTo)))
	}
	// Store code as file and get temporary directory
	workDir, err := parent.taskHandler.SlaveSaveCode(code, lang.SourceName)
	if err != nil {
		return err
	}
	defer parent.taskHandler.SlaveFinishProcess(workDir)
	log.Printf("[subId:%d] saved in temporary dir %s", sub.Id, workDir)
	compileArgs := strings.Split(lang.CompileCommand, " ")
	for k, v := range compileArgs {
		parsed := parent.ParseVars(sub, lang, workDir, v)
		compileArgs[k] = parsed
	}
	log.Printf("[subId:%d] Executing compilation with args: %v", sub.Id, compileArgs)
	compileDuration, stdout, stderr, err := parent.taskHandler.SlaveCompileCode(compileArgs, workDir)
	sub.CompileStdout = stdout
	sub.CompileStderr = stderr
	if err != nil {
		log.Errorf("[subId:%d] Error while compiling: %s", sub.Id, stderr)
		sub.Verdict = gytypes.SubmissionCompilerError
		sub.Score = 0
		resp.Submission = sub
		return nil
	}
	sub.CompileTime = compileDuration
	log.Printf("[subId:%d] Compiled after %fms", sub.Id, compileDuration)
	testCases := req.TestCases
	testCount := len(testCases)
	if testCount > 0 {
		isFailed := false
		testPassed := 0
		runArgs := strings.Split(lang.ExecuteCommand, " ")
		for k, v := range runArgs {
			parsed := parent.ParseVars(sub, lang, workDir, v)
			runArgs[k] = parsed
		}
		var testResult []gytypes.TestResultData
		for _, testCase := range testCases {
			log.Printf("[subId:%d] Executing testcase %d with args: %v", sub.Id, testCase.TestNo, runArgs)
			// For obvious reason, timeout are n*2 from defined
			duration, memory, stdout, _, err := parent.taskHandler.SlaveRunCode(runArgs, workDir, testCase.Input, prob.TimeLimit*2)
			result := gytypes.TestResultData{
				ProblemId:    sub.ProblemId,
				SubmissionId: sub.Id,
				TestNo:       testCase.TestNo,
				Verdict:      gytypes.SubmissionError,
				TimeElapsed:  duration,
				MemoryUsed:   memory,
				Score:        0,
			}
			// Seems runtime error occurs, we will investigate for
			if err != nil {
				if duration > 1000 {
					result.Verdict = gytypes.SubmissionTimeLimitExceeded
				} else if lang.LimitMemory && (memory > uint64(prob.MemLimit*1024*1024)) {
					result.Verdict = gytypes.SubmissionMemoryLimitExceeded
				} else {
					result.Verdict = gytypes.SubmissionRuntimeError
				}
			} else {
				// Not returning error, but we will check for time and memory constrains
				if duration > float64(prob.TimeLimit*1000) {
					result.Verdict = gytypes.SubmissionTimeLimitExceeded
				} else if lang.LimitMemory && (memory > uint64(prob.MemLimit*1024*1024)) {
					result.Verdict = gytypes.SubmissionMemoryLimitExceeded
				} else {
					// Now compare for stdout
					testStdoutArr := strings.Split(testCase.Output, "\n")
					resStdoutArr := strings.Split(stdout, "\n")
					// Clean from CR chars
					for k, v := range testStdoutArr {
						testStdoutArr[k] = strings.ReplaceAll(v, "\r", "")
					}
					for k, v := range resStdoutArr {
						resStdoutArr[k] = strings.ReplaceAll(v, "\r", "")
					}
					// Lines count not match, not same
					if len(testStdoutArr) != len(resStdoutArr) {
						result.Verdict = gytypes.SubmissionWrongAnswer
					} else {
						// Further checking...
						match := true
						for k, _ := range testStdoutArr {
							if testStdoutArr[k] != resStdoutArr[k] {
								match = false
								break
							}
						}
						// We got conclusion
						if match {
							result.Verdict = gytypes.SubmissionAccepted
							result.Score = 100 / float64(testCount)
							testPassed++
						} else {
							result.Verdict = gytypes.SubmissionWrongAnswer
						}
					}
				}
			}
			if (!isFailed) && (result.Verdict != gytypes.SubmissionAccepted) {
				sub.Verdict = result.Verdict
				isFailed = true
			}
			log.Printf("[subId:%d] Test case no. %d verdict:%s after %fms score:%f", sub.Id, testCase.TestNo, result.Verdict, result.TimeElapsed, result.Score)
			testResult = append(testResult, result)
		}
		resp.TestResults = testResult
		if testCount == testPassed {
			sub.Verdict = gytypes.SubmissionAccepted
		}
		sub.Score = int(math.Ceil((100 / float64(testCount)) * float64(testPassed)))
	} else {
		// Since no testcase passed, nothing to grade here
		sub.Verdict = gytypes.SubmissionAccepted
		sub.Score = 100
	}
	resp.Submission = sub
	log.Printf("[subId:%d] Graded with verdict:%s and score %d", sub.Id, sub.Verdict, sub.Score)
	return nil
}
