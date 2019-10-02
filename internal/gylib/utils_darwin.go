package gylib

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"os/exec"
	"strings"
)

func utilsGetOSName() (string, error) {
	productName, err := exec.Command("sw_vers", "-productName").Output()
	if err != nil {
		return "", err
	}
	productNameStr := strings.TrimSuffix(string(productName), "\n")
	productVer, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "", err
	}
	productVerStr := strings.TrimSuffix(string(productVer), "\n")
	kernelVer, err := exec.Command("uname", "-srm").Output()
	if err != nil {
		return "", err
	}
	kernelVerStr := strings.TrimSuffix(string(kernelVer), "\n")
	osStr := fmt.Sprintf("%s %s (%s)", productNameStr, productVerStr, kernelVerStr)
	return osStr, nil
}

func utilsGetAppDataDirectory() string {
	homeDir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return homeDir
}
