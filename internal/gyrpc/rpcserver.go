package gyrpc

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"bytes"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"html/template"
	"net"
	"net/rpc"
	"path/filepath"
	"runtime"
)

type GargoyleRpcServerVars struct {
	ExeName    string
	SourceName string
	WorkPath   string
	MemLimit   int
}

type GargoyleRpcServerTaskHandler interface {
	SlaveSaveCode(code string, sourceName string) (string, error)
	SlaveCompileCode(args []string, dir string) (float64, string, string, error)
	SlaveRunCode(args []string, dir string, stdin string, timeout int) (float64, uint64, string, string, error)
	SlaveFinishProcess(dir string) error
}

type GargoyleRpcServer struct {
	address     string
	server      rpc.Server
	listener    net.Listener
	task        GargoyleRpcTask
	taskHandler GargoyleRpcServerTaskHandler
}

func (grs *GargoyleRpcServer) SetTaskHandler(taskHandler GargoyleRpcServerTaskHandler) {
	grs.taskHandler = taskHandler
}

func NewGargoyleRpcServer(address string, taskHandler GargoyleRpcServerTaskHandler) (*GargoyleRpcServer, error) {
	grs := GargoyleRpcServer{
		address:     address,
		server:      rpc.Server{},
		task:        GargoyleRpcTask{},
		taskHandler: taskHandler,
	}
	grs.task.parentServer = &grs
	err := grs.server.Register(&grs.task)
	if err != nil {
		return nil, err
	}
	return &grs, nil
}

func (grs *GargoyleRpcServer) Address() string {
	return grs.address
}

func (grs *GargoyleRpcServer) ListenAndServe() error {
	listener, err := net.Listen("tcp", grs.address)
	if err != nil {
		return err
	}
	grs.listener = listener
	defer grs.listener.Close()
	grs.server.Accept(grs.listener)
	return nil
}

func (grs *GargoyleRpcServer) ParseVars(sub gytypes.SubmissionData, lang gytypes.LanguageProgramData, workDir string, input string) string {
	vars := GargoyleRpcServerVars{
		ExeName:    lang.ExecutableName,
		SourceName: lang.SourceName,
		WorkPath:   workDir,
		MemLimit:   32,
	}
	// For windows, it's better if PE executable is exe file
	if runtime.GOOS == "windows" {
		if filepath.Ext(vars.ExeName) == "" {
			vars.ExeName += ".exe"
		}
	}
	tpl, err := template.New("").Parse(input)
	if err != nil {
		return input
	}
	var b bytes.Buffer
	err = tpl.Execute(&b, vars)
	if err != nil {
		return input
	}
	return b.String()
}
