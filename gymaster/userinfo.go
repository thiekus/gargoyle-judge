package main

import "time"

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

type UserRoleAccess struct {
	RoleName   string
	Contestant bool
	SysAdmin   bool
	Operator   bool
}

type UserGroup struct {
	GroupName string
	GroupId   int
}

type UserGroupAccess []UserGroup

type UserInfo struct {
	Id          int
	Token       string
	Username    string
	Password    string
	Salt        string
	Email       string
	DisplayName string
	Gender      string
	Address     string
	Institution string
	CountryId   string
	Avatar      string
	RoleId      int
	Roles       UserRoleAccess
	Groups      UserGroupAccess
	LastAccess  int64
}

type UserOnline struct {
	Id           int
	Username     string
	DisplayName  string
	Institution  string
	Avatar       string
	LastAccess   int64
	LastTimeDiff int64
	TimeStatus   string
}

type UserOnlineList struct {
	Count int
	Users []UserOnline
}

func (ui *UserInfo) IsAdmin() bool {
	return ui.Roles.SysAdmin
}

func (ui *UserInfo) IsJury() bool {
	return ui.Roles.Operator
}

func (ui *UserInfo) IsContestant() bool {
	return ui.Roles.Contestant
}

func (ui *UserInfo) RefreshLastAccess() {
	ui.LastAccess = time.Now().Unix()
}
