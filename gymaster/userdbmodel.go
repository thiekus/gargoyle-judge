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
	"time"

	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
	"golang.org/x/crypto/bcrypt"
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
            (username, password, email, display_name, gender, address, institution, country_id, avatar, syntax_theme, role, active, banned, create_time, lastaccess_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 0, ?, ?)`
	prep, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer prep.Close()
	passHash := generatePasswordHash(password)
	createTime := time.Now().Unix()
	_, err = prep.Exec(
		username,
		passHash,
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

func (udm *UserDbModel) DeleteUserById(uid int) error {
	db := udm.db
	query := "DELETE FROM {{.TablePrefix}}users WHERE id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(uid)
	return err
}

func (udm *UserDbModel) GetAccessRoleList() ([]gytypes.UserRoleAccess, error) {
	db := udm.db
	query := "SELECT id, rolename, access_contestant, access_jury, access_root FROM {{.TablePrefix}}roles"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	var al []gytypes.UserRoleAccess
	for rows.Next() {
		ua := gytypes.UserRoleAccess{}
		err = rows.Scan(
			&ua.Id,
			&ua.RoleName,
			&ua.Contestant,
			&ua.Jury,
			&ua.SysAdmin,
		)
		if err != nil {
			return nil, err
		}
		al = append(al, ua)
	}
	return al, nil
}

func (udm *UserDbModel) GetUserAccess(id int) (gytypes.UserRoleAccess, error) {
	ua := gytypes.UserRoleAccess{}
	db := udm.db
	query := "SELECT id, rolename, access_contestant, access_jury, access_root FROM {{.TablePrefix}}roles WHERE id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		return ua, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(id).Scan(
		&ua.Id,
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

func (udm *UserDbModel) GetUserList() ([]gytypes.UserInfo, error) {
	db := udm.db
	query := `SELECT id, username, password, email, display_name, gender, address, institution, country_id, avatar,
        syntax_theme, role, active, banned
        FROM {{.TablePrefix}}users`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	var ul []gytypes.UserInfo
	for rows.Next() {
		ui := gytypes.UserInfo{}
		err = rows.Scan(
			&ui.Id,
			&ui.Username,
			&ui.Password,
			&ui.Email,
			&ui.DisplayName,
			&ui.Gender,
			&ui.Address,
			&ui.Institution,
			&ui.CountryId,
			&ui.Avatar,
			&ui.SyntaxTheme,
			&ui.RoleId,
			&ui.Active,
			&ui.Banned,
		)
		if err != nil {
			return nil, err
		}
		roles, err := udm.GetUserAccess(ui.RoleId)
		if err != nil {
			return nil, err
		}
		ui.Roles = roles
		ul = append(ul, ui)
	}
	return ul, nil
}

func (udm *UserDbModel) GetUserById(userId int) (gytypes.UserInfo, error) {
	ui := gytypes.UserInfo{}
	db := udm.db
	stmt, err := db.Prepare(
		`SELECT id, username, password, email, display_name, gender, address, institution, country_id, avatar,
        syntax_theme, role, active, banned
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
		&ui.Email,
		&ui.DisplayName,
		&ui.Gender,
		&ui.Address,
		&ui.Institution,
		&ui.CountryId,
		&ui.Avatar,
		&ui.SyntaxTheme,
		&ui.RoleId,
		&ui.Active,
		&ui.Banned,
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
	db := udm.db
	stmt, err := db.Prepare(
		`SELECT id, username, password, email, display_name, gender, address, institution, country_id, avatar,
        syntax_theme, role, active, banned
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
		&ui.Email,
		&ui.DisplayName,
		&ui.Gender,
		&ui.Address,
		&ui.Institution,
		&ui.CountryId,
		&ui.Avatar,
		&ui.SyntaxTheme,
		&roleId,
		&ui.Active,
		&ui.Banned,
	)
	if err != nil {
		log := gylib.GetStdLog()
		log.Errorf("[%s] Select error: %s", username, err.Error())
		return ui, errors.New("username invalid or not exists")
	}
	if comparePasswordHash(ui.Password, password) {
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
	return ui, nil
}

func (udm *UserDbModel) ModifyUserAccount(userId int, ui gytypes.UserInfo) error {
	db := udm.db
	query := `UPDATE {{.TablePrefix}}users SET
        password = ?,
        email = ?,
        display_name = ?,
        gender = ?,
        address = ?,
        institution = ?,
        country_id = ?,
        avatar = ?,
        syntax_theme = ?,
        role = ?,
        active = ?,
        banned = ?
        WHERE id = ?`
	prep, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer prep.Close()
	_, err = prep.Exec(
		ui.Password,
		ui.Email,
		ui.DisplayName,
		ui.Gender,
		ui.Address,
		ui.Institution,
		ui.CountryId,
		ui.Avatar,
		ui.SyntaxTheme,
		ui.RoleId,
		ui.Active,
		ui.Banned,
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

func (udm *UserDbModel) DeleteTokenByUserId(uid int) error {
	db := udm.db
	query := "DELETE FROM {{.TablePrefix}}tokens WHERE id_user = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(uid)
	return err
}

func (udm *UserDbModel) GetTokenList() ([]gytypes.UserTokenInfo, error) {
	db := udm.db
	query := `SELECT token, id_user, login_time FROM {{.TablePrefix}}tokens`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	var list []gytypes.UserTokenInfo
	for rows.Next() {
		tok := gytypes.UserTokenInfo{}
		err = rows.Scan(
			&tok.Token,
			&tok.UserId,
			&tok.LoginTime,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, tok)
	}
	return list, nil
}

func (udm *UserDbModel) InsertToken(token string, uid int) error {
	db := udm.db
	query := `INSERT INTO {{.TablePrefix}}tokens (token, id_user, login_time) VALUES (?, ?, ?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(token, uid, time.Now().Unix())
	return err
}

func (udm *UserDbModel) CleanTokenOfUser(uid int) error {
	db := udm.db
	query := `DELETE FROM {{.TablePrefix}}tokens WHERE id_user = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(uid)
	return err
}

func generatePasswordHash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func comparePasswordHash(passHash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passHash), []byte(password))
	return err != nil
}
