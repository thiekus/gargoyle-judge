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
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"text/template"
)

type DbContext struct {
	db     *sql.DB
	driver string
}

type DbParseVariables struct {
	Driver        string
	DatabaseName  string
	AutoIncrement string
	TablePrefix   string
}

func OpenDatabaseEx(driver string, multiStatements bool) (DbContext, error) {
	ctx := DbContext{}
	switch driver {
	// MySQL or MariaDB
	case "mysql":
		ms := "false"
		if multiStatements {
			ms = "true"
		}
		sqlConnectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?multiStatements=%s",
			appConfig.DbUsername, appConfig.DbPassword, appConfig.DbHost, appConfig.DbName, ms)
		db, err := sql.Open("mysql", sqlConnectionString)
		if err != nil {
			log := gylib.GetStdLog()
			log.Error(err)
			return ctx, err
		}
		ctx.db = db

	// SQLite 3.x
	case "sqlite3":
		db, err := sql.Open("sqlite3", gylib.ConcatByWorkDir(appConfig.DbFile))
		if err != nil {
			log := gylib.GetStdLog()
			log.Error(err)
			return ctx, err
		}
		ctx.db = db

	// Microsoft SQL Server
	case "sqlserver":
		u := &url.URL{
			Scheme:   "sqlserver",
			User:     url.UserPassword(appConfig.DbUsername, appConfig.DbPassword),
			Host:     appConfig.DbHost,
			RawQuery: fmt.Sprintf("database=%s&encrypt=disable", appConfig.DbName),
		}
		db, err := sql.Open("sqlserver", u.String())
		if err != nil {
			log := gylib.GetStdLog()
			log.Error(err)
			return ctx, err
		}
		ctx.db = db

	default:
		return ctx, errors.New("unsupported db driver")

	}
	ctx.driver = driver
	return ctx, nil
}

func OpenDatabase() (DbContext, error) {
	return OpenDatabaseEx(appConfig.DbDriver, false)
}

func (d *DbContext) Close() error {
	return d.db.Close()
}

func (d *DbContext) Db() *sql.DB {
	return d.db
}

func (d *DbContext) DriverName() string {
	return d.driver
}

func (d *DbContext) Exec(query string, args ...interface{}) (sql.Result, error) {
	if q, err := d.ParsePreprocessor(query); err != nil {
		return nil, err
	} else {
		return d.db.Exec(q, args...)
	}
}

func (d *DbContext) ParsePreprocessor(query string) (string, error) {
	vars := DbParseVariables{
		Driver:       d.driver,
		DatabaseName: appConfig.DbName,
		TablePrefix:  appConfig.DbTablePrefix,
	}
	switch d.driver {
	case "mysql":
		vars.AutoIncrement = "AUTO_INCREMENT"
	case "sqlite3":
		vars.AutoIncrement = "" // No Auto Increment keyword in sqlite
	case "sqlserver":
		vars.AutoIncrement = "IDENTITY(1,1)"
	default:
		return "", errors.New("unsupported db driver")
	}
	tpl, err := template.New("").Parse(query)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = tpl.Execute(&b, vars)
	if err != nil {
		return "", err
	}
	queryResult := b.String()
	if d.driver == "sqlserver" {
		if strings.Index(query, "\\{") >= 0 {
			queryResult = strings.ReplaceAll(queryResult, "\\{", "{")
			queryResult = strings.ReplaceAll(queryResult, "\\}", "}")
		}
	}
	return queryResult, nil
}

func (d *DbContext) Ping() error {
	return d.db.Ping()
}

func (d *DbContext) Prepare(query string) (*sql.Stmt, error) {
	if d.driver == "sqlserver" {
		// Dirty job for Ms SQL Server: prepare statement param are differ than MySQL and sqlite
		c := 1
		for strings.Index(query, "?") >= 0 {
			query = strings.Replace(query, "?", "@p"+strconv.Itoa(c), 1)
			c++
		}
	}
	if q, err := d.ParsePreprocessor(query); err != nil {
		return nil, err
	} else {
		return d.db.Prepare(q)
	}
}

func (d *DbContext) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if q, err := d.ParsePreprocessor(query); err != nil {
		return nil, err
	} else {
		return d.db.Query(q, args...)
	}
}

func CreateBlankDatabase() error {
	log := gylib.GetStdLog()
	log.Print("Begin to create new database table")
	db, err := OpenDatabaseEx(appConfig.DbDriver, true)
	if err != nil {
		log.Error(err)
		return err
	}
	defer db.Close()
	createSql, err := ioutil.ReadFile(gylib.ConcatByProgramLibDir("./default.sql"))
	if _, err = db.Exec(fmt.Sprintf("%s", createSql)); err != nil {
		log.Error(err)
		return err
	}
	log.Print("Create new database table done successfully!")
	return nil
}
