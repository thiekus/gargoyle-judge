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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/thiekus/gargoyle-judge/internal/gylib"
)

type ConfigData struct {
	HasFirstSetup bool   `json:"hasFirstSetup"`
	SessionKey    string `json:"sessionKey"`
	Hostname      string `json:"hostname"`
	ListeningPort int    `json:"listeningPort"`
	RootSubPath   string `json:"rootSubPath"`
	UseTLS        bool   `json:"useTLS"`
	ForceTLS      bool   `json:"forceTLS"`
	CrtFile       string `json:"crtFile"`
	KeyFile       string `json:"keyFile"`
	CompressOnFly bool   `json:"compressOnFly"`
	PageMinify    bool   `json:"pageMinify"`
	AssetsCaching bool   `json:"assetsCaching"`
	AssetsMinify  bool   `json:"assetsMinify"`
	DbDriver      string `json:"dbDriver"`
	DbHost        string `json:"dbHost"`
	DbUsername    string `json:"dbUsername"`
	DbPassword    string `json:"dbPassword"`
	DbFile        string `json:"dbFile"`
	DbName        string `json:"dbName"`
	DbTablePrefix string `json:"dbTablePrefix"`
}

const (
	ConfigDefaultHasFirstSetup = false
	ConfigDefaultListeningPort = 28498
	ConfigDefaultRootSubPath   = "/"
	ConfigDefaultUseTLS        = false
	ConfigDefaultCrtFile       = "./cert/server.crt"
	ConfigDefaultKeyFile       = "./cert/server.key"
	ConfigDefaultCompressOnFly = true
	ConfigDefaultPageMinify    = true
	ConfigDefaultAssetsCaching = true
	ConfigDefaultAssetsMinify  = true
	ConfigDefaultDbDriver      = "mysql"
	ConfigDefaultDbHost        = "localhost"
	ConfigDefaultDbUsername    = "root"
	ConfigDefaultDbPassword    = ""
	ConfigDefaultDbName        = "gargoyle"
	ConfigDefaultDbFile        = "./database.db"
	ConfigDefaultDbTablePrefix = "gy_"
)

const ConfigFilename = "master_config.json"

// Get configuration from config file
func getConfigData() ConfigData {
	log := gylib.GetStdLog()
	configPath := gylib.ConcatByWorkDir("./" + ConfigFilename)
	cfg := ConfigData{}
	if !gylib.IsFileExists(configPath) {
		log.Warn("Config file doesn't exists yet, recreating new configuration...")
		cfg.HasFirstSetup = ConfigDefaultHasFirstSetup
		cfg.SessionKey = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%x", securecookie.GenerateRandomKey(32)))))
		cfg.ListeningPort = ConfigDefaultListeningPort
		cfg.RootSubPath = ConfigDefaultRootSubPath
		cfg.UseTLS = ConfigDefaultUseTLS
		cfg.CrtFile = ConfigDefaultCrtFile
		cfg.KeyFile = ConfigDefaultKeyFile
		cfg.CompressOnFly = ConfigDefaultCompressOnFly
		cfg.PageMinify = ConfigDefaultPageMinify
		cfg.AssetsCaching = ConfigDefaultAssetsCaching
		cfg.AssetsMinify = ConfigDefaultAssetsMinify
		cfg.DbDriver = ConfigDefaultDbDriver
		cfg.DbHost = ConfigDefaultDbHost
		cfg.DbUsername = ConfigDefaultDbUsername
		cfg.DbPassword = ConfigDefaultDbPassword
		cfg.DbName = ConfigDefaultDbName
		cfg.DbFile = ConfigDefaultDbFile
		cfg.DbTablePrefix = ConfigDefaultDbTablePrefix
		saveConfigData(cfg)
	}
	if jsonData, err := ioutil.ReadFile(configPath); err == nil {
		if err = json.Unmarshal(jsonData, &cfg); err == nil {
			log.Print("Config data loaded...")
		} else {
			log.Fatalf("Cannot parse configuration: %s", err.Error())
		}
	} else {
		log.Fatalf("Cannot read configuration: %s", err.Error())
	}
	return cfg
}

// Save current configuration to config file
func saveConfigData(config ConfigData) {
	configPath := gylib.ConcatByWorkDir("./" + ConfigFilename)
	if jsonData, err := json.Marshal(config); err == nil {
		var jsonBuf bytes.Buffer
		if err = json.Indent(&jsonBuf, jsonData, "", "\t"); err == nil {
			if err = ioutil.WriteFile(configPath, jsonBuf.Bytes(), os.ModePerm); err == nil {
				log := gylib.GetStdLog()
				log.Print("Config data have been saved!")
			}
		}
	}
}
