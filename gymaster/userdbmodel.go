package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gorilla/securecookie"
	"time"
)

type UserDbModel struct {
	db DbContext
}

func NewUserDbModel() (UserDbModel, error) {
	udm := UserDbModel{}
	db, err := OpenDatabase()
	if err != nil {
		return udm, err
	}
	udm.db = db
	return udm, err
}

func (udm *UserDbModel) Close() error {
	return udm.db.Close()
}

func (udm *UserDbModel) CreateUserAccount(username string, password string, roleId int, ui UserInfo) error {
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
	avatar := ui.Avatar
	if avatar == "" {
		avatar = base64.StdEncoding.EncodeToString([]byte(gender + ":" + getMD5Hash(username)))
	}
	db := &udm.db
	query := `INSERT INTO {{.TablePrefix}}users 
            (username, password, salt, email, display_name, gender, address, institution, country_id, avatar, role, verified, banned, create_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 0, ?)`
	prep, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer prep.Close()
	passSalt := generateRandomSalt()
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
		roleId,
		createTime,
	)
	if err != nil {
		return err
	}
	return nil
}

func (udm *UserDbModel) GetUserAccess(id int) (UserRoleAccess, error) {
	ua := UserRoleAccess{}
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

func (udm *UserDbModel) GetUserById(userId int) (UserInfo, error) {
	ui := UserInfo{}
	db := &udm.db
	stmt, err := db.Prepare(
		`SELECT id, username, password, salt, email, display_name, gender, address, institution, country_id, avatar, role
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

func (udm *UserDbModel) GetUserByLogin(username string, password string) (UserInfo, error) {
	ui := UserInfo{}
	db := &udm.db
	stmt, err := db.Prepare(
		`SELECT id, username, password, salt, email, display_name, gender, address, institution, country_id, avatar, role
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
		&roleId,
	)
	if err != nil {
		log := newLog()
		log.Errorf("[%s] Select error: %s", username, err.Error())
		return ui, errors.New("username invalid or not exists")
	}
	passHash := calculateSaltedHash(password, ui.Salt)
	if passHash != ui.Password {
		log := newLog()
		log.Errorf("[%s] Password error: %s", username, err.Error())
		return ui, errors.New("password for this username is invalid")
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

func (udm *UserDbModel) ModifyUserAccount(userId int, ui UserInfo) error {
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
		ui.RoleId,
		userId,
	)
	return nil
}

func (udm *UserDbModel) GetUserGroupAccess(userId int) (UserGroupAccess, error) {
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
	var groups UserGroupAccess
	for rows.Next() {
		ug := UserGroup{}
		err = rows.Scan(
			&ug.GroupId,
			&ug.GroupName,
		)
		groups = append(groups, ug)
	}
	return groups, nil
}

func calculateSaltedHash(password string, salt string) string {
	full := len(salt)
	half := full / 2
	left := salt[:half]
	right := salt[half:full]
	return getSHA256Hash(fmt.Sprintf("$%s$%s$%s$", left, password, right))
}

func getSHA256Hash(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}

func getMD5Hash(password string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(password)))
}

func generateRandomSalt() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%x", securecookie.GenerateRandomKey(32)))))
}

func generateRandomToken() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%x", securecookie.GenerateRandomKey(32)))))
}
