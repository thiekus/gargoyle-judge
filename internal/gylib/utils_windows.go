package gylib

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"fmt"
	"github.com/lxn/win"
	"golang.org/x/sys/windows/registry"
	"syscall"
)

func utilsGetOSName() (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()

	productName, _, err := key.GetStringValue("ProductName")
	if err != nil {
		return "", err
	}
	versionString, _, err := key.GetStringValue("CurrentVersion")
	if err != nil {
		versionMajor, _, err := key.GetIntegerValue("CurrentMajorVersionNumber")
		if err != nil {
			return "", err
		}
		versionMinor, _, err := key.GetIntegerValue("CurrentMinorVersionNumber")
		if err != nil {
			return "", err
		}
		versionString = fmt.Sprintf("%d.%d", versionMajor, versionMinor)
	}
	versionBuild, _, err := key.GetStringValue("CurrentBuild")
	if err != nil {
		return "", err
	}

	osString := fmt.Sprintf("%s (%s build %s)",
		productName, versionString, versionBuild)
	return osString, nil
}

func utilsGetAppDataDirectory() string {
	// TODO: use SHGetKnownFolderPath instead for next time
	lpPathBuf := make([]uint16, win.MAX_PATH)
	if win.SHGetSpecialFolderPath(win.HWND(0), &lpPathBuf[0], win.CSIDL_COMMON_APPDATA, false) {
		retStr := syscall.UTF16ToString(lpPathBuf)
		return retStr
	} else {
		return ""
	}
}
