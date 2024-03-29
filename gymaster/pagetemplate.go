package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"runtime"
	"time"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	jsonmin "github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
	"github.com/thiekus/gargoyle-judge/internal/gylib"
	"github.com/thiekus/gargoyle-judge/internal/gytypes"
)

type PageDataInfo struct {
	BaseUrl     string
	BaseUrlNS   string
	AppVersion  string
	GoVersion   string
	GoPlatform  string
	OSName      string
	UserData    gytypes.UserInfo
	PageData    interface{}
	MessageStr  string
	MessageType string
	UAString    string
}

type DashboardPageDataInfo struct {
	PageName  string
	MainTitle string
	Title     string
	Content   template.HTML
	Menu      MenuTOC
}

type MenuNode struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	IconClass string `json:"iconClass"`
	Location  string `json:"location"`
	Selected  bool
}

type MenuTOC struct {
	GeneralMenu   []MenuNode `json:"generalMenu"`
	ContestMenu   []MenuNode `json:"contestMenu"`
	JuryMenu      []MenuNode `json:"juryMenu"`
	AdminMenu     []MenuNode `json:"adminMenu"`
	SelectedTitle string
}

func ParsePageMessage(w http.ResponseWriter, r *http.Request, retrieveMsg bool) (string, string) {
	if retrieveMsg {
		return appUsers.GetFlashMessage(w, r)
	}
	return "", ""
}

func GenerateMenuTOC(selectedName string) (MenuTOC, error) {
	jsonData, err := ioutil.ReadFile(gylib.ConcatByProgramLibDir("./templates/menutoc.json"))
	if err != nil {
		return MenuTOC{}, err
	}
	menu := MenuTOC{}
	err = json.Unmarshal(jsonData, &menu)
	if err != nil {
		return MenuTOC{}, err
	}
	menu.SelectedTitle = ""
	// General menu
	for i, n := range menu.GeneralMenu {
		menu.GeneralMenu[i].Selected = false
		if n.Name == selectedName {
			menu.GeneralMenu[i].Selected = true
			menu.SelectedTitle = n.Title
		}
	}
	// Contestant menu
	for i, n := range menu.ContestMenu {
		menu.ContestMenu[i].Selected = false
		if n.Name == selectedName {
			menu.ContestMenu[i].Selected = true
			menu.SelectedTitle = n.Title
		}
	}
	// Jury menu
	for i, n := range menu.JuryMenu {
		menu.JuryMenu[i].Selected = false
		if n.Name == selectedName {
			menu.JuryMenu[i].Selected = true
			menu.SelectedTitle = n.Title
		}
	}
	// Admin menu
	for i, n := range menu.AdminMenu {
		menu.AdminMenu[i].Selected = false
		if n.Name == selectedName {
			menu.AdminMenu[i].Selected = true
			menu.SelectedTitle = n.Title
		}
	}

	return menu, nil
}

func NewPageInfoData(w http.ResponseWriter, r *http.Request, pageData interface{}, retrieveMsg bool) PageDataInfo {
	msg, msgType := ParsePageMessage(w, r, retrieveMsg)
	userInfo := gytypes.UserInfo{}
	if userInfoPtr := appUsers.GetLoggedUserInfo(r); userInfoPtr != nil {
		userInfo = *userInfoPtr
	}
	uas := r.UserAgent()
	data := PageDataInfo{
		BaseUrl:     GetAppUrlWithSlash(r),
		BaseUrlNS:   GetAppUrl(r),
		AppVersion:  appVersion,
		GoVersion:   runtime.Version(),
		GoPlatform:  fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
		OSName:      appOSName,
		UserData:    userInfo,
		PageData:    pageData,
		MessageStr:  msg,
		MessageType: msgType,
		UAString:    uas,
	}
	return data
}

func CompileSinglePage(w http.ResponseWriter, r *http.Request, templatePath string, pageData interface{}) {
	execStart := time.Now()
	log := gylib.GetStdLog()
	tpl := template.Must(template.ParseFiles(fmt.Sprintf(gylib.ConcatByProgramLibDir("./templates/%s"), templatePath)))
	data := NewPageInfoData(w, r, pageData, true)
	var byteData bytes.Buffer
	if err := tpl.Execute(&byteData, data); err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error: template compile failed", 500)
		return
	}
	var b []byte
	if appConfig.PageMinify {
		m := minify.New()
		m.Add("text/html", &html.Minifier{
			KeepDefaultAttrVals: true,
		})
		m.AddFunc("text/css", css.Minify)
		m.AddFunc("image/svg+xml", svg.Minify)
		m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
		m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), jsonmin.Minify)
		m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
		bm, err := m.Bytes("text/html", byteData.Bytes())
		if err != nil {
			log.Error(err)
			http.Error(w, "Internal Server Error: minify error", 500)
			return
		}
		b = bm
	} else {
		b = byteData.Bytes()
	}
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write(b)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error: write error", 500)
		return
	}
	execEnd := time.Since(execStart)
	fmt.Fprintf(w, "\n<!-- Generated by Gargoyle Judgement System %s, took %s -->", appVersion, execEnd)
}

func CompileDashboardPage(w http.ResponseWriter, r *http.Request, templateDash string, templateContent string,
	pageName string, data interface{}, customTitle string) {
	execStart := time.Now()
	log := gylib.GetStdLog()
	tplContent := template.Must(template.ParseFiles(fmt.Sprintf(gylib.ConcatByProgramLibDir("./templates/subviews/%s"), templateContent)))
	dataContent := NewPageInfoData(w, r, data, false)
	var byteContent bytes.Buffer
	if err := tplContent.Execute(&byteContent, dataContent); err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error: template compile for content failed", 500)
		return
	}
	tplDash := template.Must(template.ParseFiles(fmt.Sprintf(gylib.ConcatByProgramLibDir("./templates/%s"), templateDash)))
	menuToc, _ := GenerateMenuTOC(pageName)
	mainTitle := menuToc.SelectedTitle
	if customTitle != "" {
		mainTitle = customTitle
	}
	dashPage := DashboardPageDataInfo{
		PageName:  pageName,
		MainTitle: mainTitle,
		Title:     menuToc.SelectedTitle,
		Content:   template.HTML(byteContent.Bytes()),
		Menu:      menuToc,
	}
	dataDash := NewPageInfoData(w, r, dashPage, true)
	var byteDash bytes.Buffer
	if err := tplDash.Execute(&byteDash, dataDash); err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error: template compile for content failed", 500)
		return
	}
	var b []byte
	if appConfig.PageMinify {
		m := minify.New()
		m.Add("text/html", &html.Minifier{
			KeepDefaultAttrVals: true,
		})
		m.AddFunc("text/css", css.Minify)
		m.AddFunc("image/svg+xml", svg.Minify)
		m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
		m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), jsonmin.Minify)
		m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
		bm, err := m.Bytes("text/html", byteDash.Bytes())
		if err != nil {
			log.Error(err)
			http.Error(w, "Internal Server Error: minify error", 500)
			return
		}
		b = bm
	} else {
		b = byteDash.Bytes()
	}
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write(b)
	if err != nil {
		log.Error(err)
		http.Error(w, "Internal Server Error: write error", 500)
		return
	}
	execEnd := time.Since(execStart)
	fmt.Fprintf(w, "\n<!-- Generated by Gargoyle Judgement System %s, took %s -->", appVersion, execEnd)
}
