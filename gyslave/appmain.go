package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"
	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gyrpc"
)

const appVersion = "0.7r69"

var appOSName string

type SlaveTaskHandler struct{}

func (sth SlaveTaskHandler) SlaveSaveCode(code string, sourceName string) (string, error) {
	var tempDir string
	for {
		tempDir = fmt.Sprintf("%s/%s", gylib.GetCacheDir(), gylib.GenerateRandomSalt())
		if !gylib.IsDirectoryExists(tempDir) {
			break
		}
	}
	err := os.Mkdir(tempDir, os.ModePerm)
	if err != nil {
		return "", err
	}
	codePath := tempDir + "/" + sourceName
	err = ioutil.WriteFile(codePath, []byte(code), os.ModePerm)
	if err != nil {
		return "", err
	}
	return tempDir, nil
}

func (sth SlaveTaskHandler) SlaveCompileCode(args []string, dir string) (float64, string, string, error) {
	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0])
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = dir
	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)
	strStdout := stdout.String()
	strStderr := stderr.String()
	durationMs := float64(duration) / float64(time.Millisecond)
	return durationMs, strStdout, strStderr, err
}

func (sth SlaveTaskHandler) SlaveRunCode(args []string, dir string, stdin string, timeout int) (float64, uint64, string, string, error) {
	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0])
	}
	//
	stdinBuf := bytes.Buffer{}
	stdinArr := strings.Split(stdin, "\n")
	for k, v := range stdinArr {
		stdinArr[k] = strings.ReplaceAll(v, "\r", "")
		stdinBuf.Write([]byte(stdinArr[k] + "\n"))
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdin = &stdinBuf
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = dir
	memoryPeakUsage := uint64(0)
	done := make(chan error)
	if err := cmd.Start(); err != nil {
		return 0, 0, "", "", err
	}
	startTime := time.Now()
	running := true
	// Goroutine for measuring memory peak usage
	pid := cmd.Process.Pid
	go func(pid int) {
		for running {
			if proc, err := process.NewProcess(int32(pid)); err == nil {
				if mem, err := proc.MemoryInfo(); err == nil {
					memUsage := mem.RSS
					if memUsage > memoryPeakUsage {
						memoryPeakUsage = memUsage
					}
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}(pid)
	// Wait until finishes
	go func() {
		done <- cmd.Wait()
	}()
	var duration time.Duration
	var err error
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		duration = time.Since(startTime)
		err = cmd.Process.Kill()

	case err = <-done:
		duration = time.Since(startTime)
	}
	running = false
	durationMs := float64(duration) / float64(time.Millisecond)
	return durationMs, memoryPeakUsage, stdout.String(), stderr.String(), err
}

func (sth SlaveTaskHandler) SlaveFinishProcess(dir string) error {
	return os.RemoveAll(dir)
}

func main() {
	fmt.Printf("Gargoyle Judgement System v%s (Slave Server)\n", appVersion)
	fmt.Println("Copyright (C) Thiekus 2019")
	fmt.Printf("Built using %s\n", runtime.Version())
	if osName, err := gylib.GetOSName(); err != nil {
		panic(err)
	} else {
		appOSName = osName
	}
	fmt.Printf("Running on %s\n\n", appOSName)

	log := gylib.GetStdLog()
	log.Print("Initializing slave server...")
	server, err := gyrpc.NewGargoyleRpcServer(":28499", SlaveTaskHandler{})
	if err != nil {
		panic(err)
	}
	log.Print("Listening and serve RPC Server...")
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
