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

type LanguageDbModel struct {
	db DbContext
}

func NewLanguageDbModel(db DbContext) LanguageDbModel {
	ldm := LanguageDbModel{
		db: db,
	}
	return ldm
}

func (ldm *LanguageDbModel) GetLanguageList() (gytypes.LanguageProgramMap, error) {
	db := ldm.db
	query := `SELECT id, ext_name, display_name, enabled, syntax_name, source_name, exe_name, compile_cmd, exec_cmd,
       enable_sandbox, limit_memory, limit_syscall, preg_replace_from, preg_replace_to, forbidden_keys
       FROM {{.TablePrefix}}languages`
	prep, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer prep.Close()
	rows, err := prep.Query()
	if err != nil {
		return nil, err
	}
	langList := make(gytypes.LanguageProgramMap)
	for rows.Next() {
		lp := gytypes.LanguageProgramData{}
		err = rows.Scan(
			&lp.Id,
			&lp.ExtensionName,
			&lp.DisplayName,
			&lp.Enabled,
			&lp.SyntaxName,
			&lp.SourceName,
			&lp.ExecutableName,
			&lp.CompileCommand,
			&lp.ExecuteCommand,
			&lp.EnableSandbox,
			&lp.LimitMemory,
			&lp.LimitSyscall,
			&lp.RegexReplaceFrom,
			&lp.RegexReplaceTo,
			&lp.ForbiddenKeys,
		)
		if err != nil {
			return nil, err
		}
		langList[lp.Id] = lp
	}
	return langList, nil
}
