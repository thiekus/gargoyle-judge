package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

func setAssetsWithCaching(r *mux.Router) {
	log := newLog()
	// Initialize assets go-cache
	c := cache.New(30*time.Minute, 1*time.Hour)
	r.PathPrefix("/assets/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		localPath := "." + r.URL.Path
		if fileData, cached := c.Get(localPath); cached {
			//log.Printf("Hit cache for %s", localPath)
			contentType, _ := c.Get(localPath + ":type")
			w.Header().Set("Content-Type", contentType.(string))
			w.Header().Set("Content-Length", strconv.Itoa(len(fileData.([]byte))))
			w.Header().Set("Cache-Control", "public, max-age=3600")
			w.Write(fileData.([]byte))
		} else {
			fileData, err := ioutil.ReadFile(localPath)
			if err == nil {
				var contentType string
				ext := filepath.Ext(localPath)
				switch ext {
				case ".js":
					contentType = "application/javascript"
				case ".css":
					contentType = "text/css"
				case ".svg":
					contentType = "image/svg+xml"
				default:
					contentType = http.DetectContentType(fileData)
				}
				m := minify.New()
				if appConfig.AssetsMinify {
					m.Add("text/html", &html.Minifier{
						KeepDefaultAttrVals: true,
					})
					m.AddFunc("text/css", css.Minify)
					m.AddFunc("image/svg+xml", svg.Minify)
					m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
					m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
					m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
					fileData, err = m.Bytes(contentType, fileData)
				}
				c.Set(localPath, fileData, cache.DefaultExpiration)
				c.Set(localPath+":type", contentType, cache.NoExpiration)
				w.Header().Set("Content-Type", contentType)
				w.Header().Set("Content-Length", strconv.Itoa(len(fileData)))
				w.Header().Set("Cache-Control", "public, max-age=3600")
				w.Write(fileData)
			} else {
				log.Errorf("File %s not found", localPath)
				http.Error(w, "403 Forbidden", 403)
			}
		}
	})
}
