package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gorilla/securecookie"
	"io/ioutil"
	"os"
)

type ConfigData struct {
	HasFirstSetup bool   `json:"hasFirstSetup"`
	SessionKey    string `json:"sessionKey"`
	Hostname      string `json:"hostname"`
	ListeningPort int    `json:"listeningPort"`
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
}

const ConfigDefaultHasFirstSetup = false
const ConfigDefaultHostname = "localhost"
const ConfigDefaultListeningPort = 28498
const ConfigDefaultUseTLS = false
const ConfigDefaultForceTLS = false
const ConfigDefaultCrtFile = "./cert/server.crt"
const ConfigDefaultKeyFile = "./cert/server.key"
const ConfigDefaultCompressOnFly = true
const ConfigDefaultPageMinify = true
const ConfigDefaultAssetsCaching = true
const ConfigDefaultAssetsMinify = true
const ConfigDefaultDbDriver = "mysql"
const ConfigDefaultDbHost = "localhost:3306"
const ConfigDefaultDbUsername = "root"
const ConfigDefaultDbPassword = ""
const ConfigDefaultDbName = "gargoyle"
const ConfigDefaultDbFile = "./database.db"

const ConfigFilename = "master_config.json"

// Get configuration from config file
func getConfigData() ConfigData {
	log:= newLog()
	configPath:= "./" + ConfigFilename
	cfg:= ConfigData{}
	if !isFileExists(configPath) {
		log.Warn("Config file doesn't exists yet, recreating new configuration...")
		cfg.HasFirstSetup = ConfigDefaultHasFirstSetup
		cfg.SessionKey = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%x", securecookie.GenerateRandomKey(32)))))
		cfg.Hostname = ConfigDefaultHostname
		cfg.ListeningPort = ConfigDefaultListeningPort
		cfg.UseTLS = ConfigDefaultUseTLS
		cfg.ForceTLS = ConfigDefaultForceTLS
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
		saveConfigData(cfg)
	}
	if jsonData, err:= ioutil.ReadFile(configPath); err == nil {
		if err = json.Unmarshal(jsonData, &cfg); err == nil {
			log.Print("Config data loaded...")
		} else {
			log.Fatalf("Cannot parse configuration: %s", err.Error())
		}
	} else {
		log.Fatal("Cannot read configuration: %s", err.Error())
	}
	return cfg
}

// Save current configuration to config file
func saveConfigData(config ConfigData) {
	configPath:= "./" + ConfigFilename
	if jsonData, err:= json.Marshal(config); err == nil {
		if err = ioutil.WriteFile(configPath, jsonData, os.ModePerm); err == nil {
			log:= newLog()
			log.Print("Config data have been saved!")
		}
	}
}
