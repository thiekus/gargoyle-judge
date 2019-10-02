package gytypes

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

type LanguageProgramData struct {
	Id               int
	ExtensionName    string
	DisplayName      string
	Enabled          bool
	SyntaxName       string
	SourceName       string
	ExecutableName   string
	CompileCommand   string
	ExecuteCommand   string
	EnableSandbox    bool
	LimitMemory      bool
	LimitSyscall     bool
	RegexReplaceFrom string
	RegexReplaceTo   string
	ForbiddenKeys    string
}

type LanguageProgramMap map[int]LanguageProgramData
