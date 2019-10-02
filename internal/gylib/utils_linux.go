package gylib

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"github.com/mitchellh/go-homedir"
	"os/exec"
	"strings"
)

func utilsGetOSName() (string, error) {
	out, err := exec.Command("/bin/uname", "-srm").Output()
	if err != nil {
		return "", err
	}
	osStr := strings.TrimSuffix(string(out), "\n")
	return osStr, nil
}

func utilsGetAppDataDirectory() string {
	homeDir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return homeDir
}
