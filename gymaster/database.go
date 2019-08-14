package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
)

type DbContext struct {
	db *sql.DB
}

func OpenDatabase(multiStatements bool) (DbContext, error) {
	ctx := DbContext{}
	if appConfig.DbDriver == "mysql" {
		ms := "false"
		if multiStatements {
			ms = "true"
		}
		sqlConnectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?multiStatements=%s",
			appConfig.DbUsername, appConfig.DbPassword, appConfig.DbHost, appConfig.DbName, ms)
		db, err := sql.Open("mysql", sqlConnectionString)
		if err != nil {
			log := newLog()
			log.Error(err)
			return ctx, err
		}
		ctx.db = db
	} else if appConfig.DbDriver == "sqlite3" {
		db, err := sql.Open("sqlite3", appConfig.DbFile)
		if err != nil {
			log := newLog()
			log.Error(err)
			return ctx, err
		}
		ctx.db = db
	} else {
		return ctx, errors.New("invalid db driver")
	}
	return ctx, nil
}

func (d *DbContext) Close() error {
	return d.db.Close()
}

func (d *DbContext) Db() *sql.DB {
	return d.db
}

func (d *DbContext) Prepare(query string) (*sql.Stmt, error) {
	return d.db.Prepare(query)
}

func CreateBlankDatabase() error {
	log := newLog()
	log.Print("Begin to create new database table")
	db, err := OpenDatabase(true)
	if err != nil {
		log.Error(err)
		return err
	}
	defer db.Close()
	createSql, err := ioutil.ReadFile("./default.sql")
	if _, err = db.db.Exec(fmt.Sprintf("%s", createSql)); err != nil {
		log.Error(err)
		return err
	}
	log.Print("Create new database table done successfully!")
	return nil
}
