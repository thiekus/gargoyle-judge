package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

type UserAccess struct {
	RoleName   string
	Contestant bool
	SysAdmin   bool
	Operator   bool
}

type UserGroups struct {
	GroupName string
	GroupId   int
}

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
	Roles       UserAccess
	Groups      []UserGroups
	LastAccess  int
}
