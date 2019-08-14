package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"bytes"
	"github.com/o1egl/govatar"
	"image/jpeg"
	"net/http"
)

func avatarGetEndpoint(w http.ResponseWriter, r *http.Request) {
	img, err := govatar.Generate(govatar.MALE)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var b bytes.Buffer
	err = jpeg.Encode(&b, img, &jpeg.Options{Quality: 70})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Add("Content-Type", "image/jpeg")
	w.Write(b.Bytes())
}
