package gylib

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

const dateTimePickerFormatTime = "2006/01/02 15:04"

// Based from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func GetBaseUrl(r *http.Request) string {
	if r.TLS != nil {
		return fmt.Sprintf("https://%s", r.Host)
	} else {
		return fmt.Sprintf("http://%s", r.Host)
	}
}

func GetBaseUrlWithSlash(r *http.Request) string {
	if r.TLS != nil {
		return fmt.Sprintf("https://%s/", r.Host)
	} else {
		return fmt.Sprintf("http://%s/", r.Host)
	}
}

func IsDirectoryExists(dir string) bool {
	return IsFileExists(dir)
}

func IsFileExists(file string) bool {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		return true
	}
	return false
}

func IsHexValue(s string) bool {
	for _, c := range s {
		if (c < 'a' || c > 'f') && (c < 'A' || c > 'F') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

var appLog *logrus.Logger

func GetStdLog() *logrus.Logger {
	if appLog == nil {
		appLog = logrus.New()
		// Much better logging
		appLog.SetFormatter(&nested.Formatter{
			HideKeys: true,
		})
		logDir := GetWorkDir() + "/log"
		if !IsDirectoryExists(logDir) {
			if err := os.Mkdir(logDir, os.ModePerm); err != nil {
				panic(err)
			}
		}
		t := time.Now()
		logPath := fmt.Sprintf("%s/%s_%x_%02d-%02d-%d_%02d-%02d-%02d.log", logDir, filepath.Base(os.Args[0]), os.Getpid(), t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second())
		logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		logMulti := io.MultiWriter(colorable.NewColorableStdout(), logFile)
		appLog.SetOutput(logMulti)
		appLog.Printf("Log file will be printed on %s", logPath)
	}
	return appLog
}

func GetProgramLibDir() string {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir := filepath.ToSlash(filepath.Dir(filepath.Dir(exePath)) + "/lib")
	return dir
}

func ConcatByProgramLibDir(path string) string {
	if (path[0] == '.') && (path[1] == '/') {
		dir := GetProgramLibDir()
		resultDir := filepath.ToSlash(strings.Replace(path, ".", dir, 1))
		return resultDir
	}
	return path
}

func GetOSName() (string, error) {
	return utilsGetOSName()
}

func GetAppDataDirectory() string {
	var appData string
	if runtime.GOOS == "windows" {
		appData = utilsGetAppDataDirectory() + "/Gargoyle Judge"
	} else {
		appData = utilsGetAppDataDirectory() + "/.gargoyle_judge"
	}
	appData = filepath.ToSlash(appData)
	if !IsDirectoryExists(appData) {
		if err := os.Mkdir(appData, os.ModePerm); err != nil {
			panic(err)
		}
	}
	return appData
}

func GetWorkDir() string {
	programDir := GetProgramLibDir()
	if IsFileExists(programDir + "/gy_not_portable") {
		appDir := GetAppDataDirectory()
		return appDir
	} else {
		return programDir
	}
}

func ConcatByWorkDir(path string) string {
	if (path[0] == '.') && (path[1] == '/') {
		dir := GetWorkDir()
		resultDir := filepath.ToSlash(strings.Replace(path, ".", dir, 1))
		return resultDir
	}
	return path
}

func GetCacheDir() string {
	cacheDir := GetWorkDir() + "/caches"
	if !IsDirectoryExists(cacheDir) {
		if err := os.Mkdir(cacheDir, os.ModePerm); err != nil {
			panic(err)
		}
	}
	return cacheDir
}

func GetSHA256Hash(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}

func GetMD5Hash(password string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(password)))
}

func GenerateRandomSalt() string {
	return RandString(64)
}

func GenerateRandomToken() string {
	return RandString(64)
}

func StringToTime(dtsStr string) time.Time {
	t, _ := time.Parse(dateTimePickerFormatTime, dtsStr)
	return t
}

func TimeToString(t time.Time) string {
	return t.Format(dateTimePickerFormatTime)
}

func TimeToHMS(t time.Time) string {
	timeInt := t.Unix()
	if timeInt > int64(24*3600) {
		timeDays := timeInt / int64(24*3600)
		return fmt.Sprintf("%dd %.2d:%.2d:%.2d", timeDays, t.Hour(), t.Minute(), t.Second())
	}
	return fmt.Sprintf("%.2d:%.2d:%.2d", t.Hour(), t.Minute(), t.Second())
}

func Btoi(val bool) int {
	if val {
		return 1
	}
	return 0
}
