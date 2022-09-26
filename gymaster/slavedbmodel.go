package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
)

type SlaveDbModel struct {
	db DbContext
}

func NewSlaveDbModel(db DbContext) SlaveDbModel {
	sdm := SlaveDbModel{
		db: db,
	}
	return sdm
}

func (sdm *SlaveDbModel) GetSlaveList() ([]gytypes.SlaveData, error) {
	db := sdm.db
	query := `SELECT id, name, address, enable FROM {{.TablePrefix}}slaves`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	var slist []gytypes.SlaveData
	for rows.Next() {
		slave := gytypes.SlaveData{}
		err = rows.Scan(
			&slave.Id,
			&slave.Name,
			&slave.Address,
			&slave.Enable,
		)
		if err != nil {
			return nil, err
		}
		slist = append(slist, slave)
	}
	return slist, nil
}
