package gyrpc

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"net"
	"net/rpc"
	"time"

	"github.com/thiekus/gargoyle-judge/internal/gytypes"
)

type GargoyleRpcClient struct {
	address string
	client  *rpc.Client
}

func (grc *GargoyleRpcClient) Address() string {
	return grc.address
}

func NewGargoyleRpcClient(address string) (*GargoyleRpcClient, error) {
	// Allow set timeout to 15 seconds instead default 3 minutes
	conn, err := net.DialTimeout("tcp", address, 15*time.Second)
	if err != nil {
		return nil, err
	}
	client := rpc.NewClient(conn)
	grc := GargoyleRpcClient{
		address: address,
		client:  client,
	}
	return &grc, nil
}

func (grc *GargoyleRpcClient) Close() error {
	return grc.client.Close()
}

func (grc *GargoyleRpcClient) PingSlave() (RpcPingResponse, error) {
	req := RpcPingRequest{
		StartTime: time.Now().UnixNano(),
	}
	var resp RpcPingResponse
	err := grc.client.Call("GargoyleRpcTask.PingSlave", req, &resp)
	return resp, err
}

func (grc *GargoyleRpcClient) ProcessSubmission(submission gytypes.SubmissionData, lang gytypes.LanguageProgramData, problem gytypes.ProblemData, tests []gytypes.TestCaseData) (RpcSubmissionResponse, error) {
	req := RpcSubmissionRequest{
		Submission:     submission,
		ProgramLang:    lang,
		ProblemDetails: problem,
		TestCases:      tests,
	}
	var resp RpcSubmissionResponse
	err := grc.client.Call("GargoyleRpcTask.ProcessSubmission", req, &resp)
	return resp, err
}
