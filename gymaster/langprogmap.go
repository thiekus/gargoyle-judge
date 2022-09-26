package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"errors"
	"strconv"
	"strings"

	"github.com/thiekus/gargoyle-judge/internal/gytypes"
)

type LanguageProgramController struct {
	lmap    gytypes.LanguageProgramMap
	fetched bool
}

func MakeLanguageProgramController() LanguageProgramController {
	lpc := LanguageProgramController{
		lmap:    nil,
		fetched: false,
	}
	return lpc
}

func (lpc *LanguageProgramController) RefreshLanguageList() error {
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()
	ldm := NewLanguageDbModel(db)
	lm, err := ldm.GetLanguageList()
	if err != nil {
		return err
	}
	lpc.lmap = lm
	lpc.fetched = true
	return nil
}

func (lpc *LanguageProgramController) GetLanguageMap() (gytypes.LanguageProgramMap, error) {
	if !lpc.fetched {
		err := lpc.RefreshLanguageList()
		if err != nil {
			return nil, err
		}
	}
	return lpc.lmap, nil
}

func (lpc *LanguageProgramController) GetLanguageFromId(id int) (*gytypes.LanguageProgramData, error) {
	// If not fetched, must fetched first!
	if !lpc.fetched {
		err := lpc.RefreshLanguageList()
		if err != nil {
			return nil, err
		}
	}
	if lp, exists := lpc.lmap[id]; exists {
		return &lp, nil
	}
	return nil, errors.New("language not found")
}

func (lpc *LanguageProgramController) GetLanguageOfContest(langList string) ([]gytypes.LanguageProgramData, error) {
	// If not fetched, must fetched first!
	if !lpc.fetched {
		err := lpc.RefreshLanguageList()
		if err != nil {
			return nil, err
		}
	}
	var langResult []gytypes.LanguageProgramData
	langStrArr := strings.Split(langList, ",")
	for _, langStrId := range langStrArr {
		langId, err := strconv.Atoi(langStrId)
		if err != nil {
			return nil, err
		}
		if lp, exists := lpc.lmap[langId]; exists {
			if lp.Enabled {
				langResult = append(langResult, lp)
			}
		}
	}
	return langResult, nil
}
