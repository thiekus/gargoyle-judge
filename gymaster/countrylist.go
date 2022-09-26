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
	"strings"

	"github.com/thiekus/gargoyle-judge/internal/gylib"
)

type CountryListName map[string]string

func GetCountryListName() (CountryListName, error) {
	listFile := gylib.ConcatByProgramLibDir("./templates/country.json")
	lf, err := os.Open(listFile)
	if err != nil {
		return nil, err
	}
	defer lf.Close()
	countryData, err := ioutil.ReadAll(lf)
	if err != nil {
		return nil, err
	}
	var countryListRaw CountryListName
	err = json.Unmarshal(countryData, &countryListRaw)
	if err != nil {
		return nil, err
	}
	var countryList CountryListName = make(CountryListName)
	for k, v := range countryListRaw {
		countryList[strings.ToLower(k)] = v
	}
	return countryList, nil
}
