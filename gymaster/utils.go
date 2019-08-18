package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func getBaseUrl(r *http.Request) string {
	if r.TLS != nil {
		return fmt.Sprintf("https://%s/", r.Host)
	} else {
		return fmt.Sprintf("http://%s/", r.Host)
	}
}

func isFileExists(file string) bool {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		return true
	}
	return false
}

func newLog() *logrus.Logger {
	log := logrus.New()
	// Much better logging
	log.SetFormatter(&nested.Formatter{
		HideKeys: true,
	})
	log.SetOutput(colorable.NewColorableStdout())
	return log
}
