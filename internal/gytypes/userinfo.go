package gytypes

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import "time"

type UserRoleAccess struct {
	RoleName   string
	Contestant bool
	SysAdmin   bool
	Jury       bool
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
	SyntaxTheme string
	RoleId      int
	Roles       UserRoleAccess
	Groups      UserGroupAccess
	// Unused at this moment, but exists in database schema
	CreateTime time.Time
	LastAccess time.Time
}

type UserOnline struct {
	Id           int
	Username     string
	DisplayName  string
	Institution  string
	Avatar       string
	LastAccess   time.Time
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
	return ui.Roles.Jury
}

func (ui *UserInfo) IsContestant() bool {
	return ui.Roles.Contestant
}

func (ui *UserInfo) RefreshLastAccess() {
	ui.LastAccess = time.Now()
}
