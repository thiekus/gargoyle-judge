package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
)

type PageAccessPerm struct {
	Prefix     string `json:"prefix"`
	Contestant bool   `json:"contestant"`
	Jury       bool   `json:"jury"`
	Admin      bool   `json:"admin"`
}

type PageAccessPermList []PageAccessPerm

func IsPageAccessHasPermission(pa PageAccessPerm, role gytypes.UserRoleAccess) bool {
	// From highest to lowest privileges
	if pa.Admin && role.SysAdmin {
		return true
	}
	if pa.Jury && role.Jury {
		return true
	}
	if pa.Contestant && role.Contestant {
		return true
	}
	return false
}

func GetPageAccessPermission() (PageAccessPermList, error) {
	listFile := gylib.ConcatByProgramLibDir("./templates/accessperm.json")
	lf, err := os.Open(listFile)
	if err != nil {
		return nil, err
	}
	defer lf.Close()
	accessRawData, err := ioutil.ReadAll(lf)
	if err != nil {
		return nil, err
	}
	var accessData PageAccessPermList
	err = json.Unmarshal(accessRawData, &accessData)
	if err != nil {
		return nil, err
	}
	return accessData, nil
}
