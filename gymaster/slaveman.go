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
	"github.com/thiekus/gargoyle-judge/internal/gyrpc"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"sync"
	"time"
)

type SlaveManager struct {
	slaveCount   int
	activeCount  int
	slaves       []gytypes.SlaveData
	selIndex     int
	refreshMutex sync.Mutex
	fetchMutex   sync.Mutex
}

func MakeSlaveManager() SlaveManager {
	sm := SlaveManager{
		slaveCount:  0,
		activeCount: 0,
		slaves:      []gytypes.SlaveData{},
		selIndex:    0,
	}
	return sm
}

func (sm *SlaveManager) CheckForRefresh() error {
	if sm.activeCount == 0 {
		err := sm.RefreshSlaves()
		return err
	}
	return nil
}

func (sm *SlaveManager) RefreshSlaves() error {
	sm.refreshMutex.Lock()
	defer sm.refreshMutex.Unlock()
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()
	sdm := NewSlaveDbModel(db)
	sl, err := sdm.GetSlaveList()
	if err != nil {
		return err
	}
	activeCount := 0
	for idx, slave := range sl {
		_, err := sm.TestPing(slave.Address)
		sl[idx].Active = err == nil
		if err == nil {
			activeCount++
		}
	}
	sm.activeCount = activeCount
	sm.slaveCount = len(sl)
	sm.slaves = sl
	sm.selIndex = 0
	return nil
}

func (sm *SlaveManager) TestPing(address string) (float64, error) {
	client, err := gyrpc.NewGargoyleRpcClient(address)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	resp, err := client.PingSlave()
	if err != nil {
		return 0, err
	}
	delta := float64(resp.Delta) / float64(time.Millisecond)
	return delta, nil
}

func (sm *SlaveManager) GetActiveSlave() (*gytypes.SlaveData, error) {
	if err := sm.CheckForRefresh(); err != nil {
		return nil, err
	}
	sm.fetchMutex.Lock()
	defer sm.fetchMutex.Unlock()
	idx := sm.selIndex
	originIdx := idx
	var selSlave *gytypes.SlaveData
	for {
		sl := sm.slaves[idx]
		if _, err := sm.TestPing(sl.Address); err == nil {
			// set for next slave role
			sm.selIndex = idx + 1
			if sm.selIndex >= sm.slaveCount {
				sm.selIndex = 0
			}
			// use selected slave
			selSlave = &sl
			break
		}
		idx++
		if idx >= sm.slaveCount {
			idx = 0
		}
		if idx == originIdx {
			return nil, errors.New("no available active slave to process")
		}
	}
	return selSlave, nil
}
