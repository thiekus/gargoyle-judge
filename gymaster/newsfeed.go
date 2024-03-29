package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"html/template"
	"time"
)

type NewsData struct {
	Id       int
	Title    string
	Author   string
	PostDate string
	Contents template.HTML
}

type NewsFeed struct {
	Count int
	News  []NewsData
}

func fetchNewsFeed() NewsFeed {
	nf := NewsFeed{
		Count: 0,
		News:  nil,
	}
	db, err := OpenDatabase()
	if err != nil {
		return nf
	}
	defer db.Close()
	newsQuery := `SELECT n.id, u.display_name, n.post_time, n.title, n.body
       FROM {{.TablePrefix}}news AS n INNER JOIN {{.TablePrefix}}users AS u
       ON u.id = n.author_id`
	stmt, err := db.Prepare(newsQuery)
	if err != nil {
		return nf
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return nf
	}
	for rows.Next() {
		nd := NewsData{}
		newsTime := 0
		err = rows.Scan(
			&nd.Id,
			&nd.Author,
			&newsTime,
			&nd.Title,
			&nd.Contents,
		)
		if err != nil {
			return nf
		}
		ut := time.Unix(int64(newsTime), 0).Local()
		nd.PostDate = ut.Format(time.RFC3339)
		nf.News = append(nf.News, nd)
		nf.Count++
	}
	return nf
}
