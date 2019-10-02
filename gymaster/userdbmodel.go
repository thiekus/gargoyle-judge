package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"time"
)

type UserDbModel struct {
	db DbContext
}

func NewUserDbModel(db DbContext) UserDbModel {
	udm := UserDbModel{
		db: db,
	}
	return udm
}

func (udm *UserDbModel) CreateUserAccount(username string, password string, roleId int, ui gytypes.UserInfo) error {
	displayName := ui.DisplayName
	if displayName == "" {
		displayName = username
	}
	gender := ui.Gender
	if gender == "" {
		gender = "M"
	}
	address := ui.Address
	if address == "" {
		address = "Somewhere"
	}
	institution := ui.Institution
	if institution == "" {
		institution = "Unknown Organization"
	}
	countryId := ui.CountryId
	if countryId == "" {
		countryId = "id"
	}
	syntaxTheme := ui.SyntaxTheme
	if syntaxTheme == "" {
		syntaxTheme = "eclipse"
	}
	avatar := ui.Avatar
	if avatar == "" {
		avatar = base64.StdEncoding.EncodeToString([]byte(gender + ":" + gylib.GetMD5Hash(username)))
	}
	db := &udm.db
	query := `INSERT INTO {{.TablePrefix}}users 
            (username, password, salt, email, display_name, gender, address, institution, country_id, avatar, syntax_theme, role, verified, banned, create_time, lastaccess_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 0, ?, ?)`
	prep, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer prep.Close()
	passSalt := gylib.GenerateRandomSalt()
	passHash := calculateSaltedHash(password, passSalt)
	createTime := time.Now().Unix()
	_, err = prep.Exec(
		username,
		passHash,
		passSalt,
		ui.Email,
		displayName,
		gender,
		address,
		institution,
		countryId,
		avatar,
		syntaxTheme,
		roleId,
		createTime,
		createTime,
	)
	if err != nil {
		return err
	}
	return nil
}

func (udm *UserDbModel) GetUserAccess(id int) (gytypes.UserRoleAccess, error) {
	ua := gytypes.UserRoleAccess{}
	db := &udm.db
	query := "SELECT rolename, access_contestant, access_jury, access_root FROM {{.TablePrefix}}roles WHERE id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		return ua, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(id).Scan(
		&ua.RoleName,
		&ua.Contestant,
		&ua.Jury,
		&ua.SysAdmin,
	)
	if err != nil {
		return ua, errors.New("invalid role id")
	}
	return ua, nil
}

func (udm *UserDbModel) GetUserById(userId int) (gytypes.UserInfo, error) {
	ui := gytypes.UserInfo{}
	db := &udm.db
	stmt, err := db.Prepare(
		`SELECT id, username, password, salt, email, display_name, gender, address, institution, country_id, avatar, syntax_theme, role
        FROM {{.TablePrefix}}users WHERE id = ?`)
	if err != nil {
		return ui, err
	}
	defer stmt.Close()
	var uid int
	err = stmt.QueryRow(userId).Scan(
		&uid,
		&ui.Username,
		&ui.Password,
		&ui.Salt,
		&ui.Email,
		&ui.DisplayName,
		&ui.Gender,
		&ui.Address,
		&ui.Institution,
		&ui.CountryId,
		&ui.Avatar,
		&ui.SyntaxTheme,
		&ui.RoleId,
	)
	if err != nil {
		return ui, errors.New("username or password either invalid or not exists")
	}
	ui.Id = uid
	ui.Roles, err = udm.GetUserAccess(ui.RoleId)
	if err != nil {
		return ui, err
	}
	ui.Groups, err = udm.GetUserGroupAccess(uid)
	if err != nil {
		return ui, err
	}
	return ui, nil
}

func (udm *UserDbModel) GetUserByLogin(username string, password string) (gytypes.UserInfo, error) {
	ui := gytypes.UserInfo{}
	db := &udm.db
	stmt, err := db.Prepare(
		`SELECT id, username, password, salt, email, display_name, gender, address, institution, country_id, avatar, syntax_theme, role
        FROM {{.TablePrefix}}users WHERE username = ?`)
	if err != nil {
		return ui, err
	}
	defer stmt.Close()
	var uid int
	var roleId int
	err = stmt.QueryRow(username).Scan(
		&uid,
		&ui.Username,
		&ui.Password,
		&ui.Salt,
		&ui.Email,
		&ui.DisplayName,
		&ui.Gender,
		&ui.Address,
		&ui.Institution,
		&ui.CountryId,
		&ui.Avatar,
		&ui.SyntaxTheme,
		&roleId,
	)
	if err != nil {
		log := gylib.GetStdLog()
		log.Errorf("[%s] Select error: %s", username, err.Error())
		return ui, errors.New("username invalid or not exists")
	}
	passHash := calculateSaltedHash(password, ui.Salt)
	if passHash != ui.Password {
		err = errors.New("password authentication for this username is invalid")
		log := gylib.GetStdLog()
		log.Errorf("[%s] Password error: %s", username, err.Error())
		return ui, err
	}
	ui.Id = uid
	ui.Roles, err = udm.GetUserAccess(roleId)
	if err != nil {
		return ui, err
	}
	ui.Groups, err = udm.GetUserGroupAccess(uid)
	if err != nil {
		return ui, err
	}
	// For security reason, unsalted password will be re-encrypted again by same login passphrase
	if ui.Salt == "" {
		ui.Salt = gylib.GenerateRandomSalt()
		ui.Password = calculateSaltedHash(password, ui.Salt)
		ui.RoleId = roleId // catch this bug since updating password would null its role id
		err = udm.ModifyUserAccount(ui.Id, ui)
		if err != nil {
			return ui, err
		}
	}
	return ui, nil
}

func (udm *UserDbModel) ModifyUserAccount(userId int, ui gytypes.UserInfo) error {
	db := &udm.db
	query := `UPDATE {{.TablePrefix}}users SET 
        password = ?,
        salt = ?,
        email = ?,
        display_name = ?,
        gender = ?,
        address = ?,
        institution = ?,
        country_id = ?,
        avatar = ?,
        syntax_theme = ?,
        role = ?
        WHERE id = ?`
	prep, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer prep.Close()
	_, err = prep.Exec(
		ui.Password,
		ui.Salt,
		ui.Email,
		ui.DisplayName,
		ui.Gender,
		ui.Address,
		ui.Institution,
		ui.CountryId,
		ui.Avatar,
		ui.SyntaxTheme,
		ui.RoleId,
		userId,
	)
	return nil
}

func (udm *UserDbModel) GetUserGroupAccess(userId int) (gytypes.UserGroupAccess, error) {
	db := udm.db
	query := `SELECT gm.group_id, (SELECT gr.name FROM {{.TablePrefix}}groups as gr WHERE id=gm.group_id) 
        FROM {{.TablePrefix}}group_members as gm WHERE gm.user_id = ?`
	prep, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer prep.Close()
	rows, err := prep.Query(userId)
	if err != nil {
		return nil, err
	}
	var groups gytypes.UserGroupAccess
	for rows.Next() {
		ug := gytypes.UserGroup{}
		err = rows.Scan(
			&ug.GroupId,
			&ug.GroupName,
		)
		groups = append(groups, ug)
	}
	return groups, nil
}

func calculateSaltedHash(password string, salt string) string {
	// For ease for compatibility with Infest registration info, allow unsalted password
	// GetUserByLogin method will be responsible to reset into salted form to avoid password leak while first login
	if salt != "" {
		full := len(salt)
		half := full / 2
		left := salt[:half]
		right := salt[half:full]
		return gylib.GetSHA256Hash(fmt.Sprintf("$%s$%s$%s$", left, password, right))
	} else {
		return gylib.GetSHA256Hash(password)
	}
}
