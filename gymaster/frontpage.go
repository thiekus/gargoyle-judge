package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import "net/http"

// Just simple homepage endpoint
func homeGetEndpoint(w http.ResponseWriter, r *http.Request) {
	// To avoid broken Secure Cookies, clean every want to login
	appUsers.CleanCookies(w, r)
	CompileSinglePage(w, r, "index.html", nil)
}
