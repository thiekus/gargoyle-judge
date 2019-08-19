package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"errors"
	"time"
)

type ImageStreamData struct {
	LastUpdate int64
	Data       []byte
	DataType   string
}

type ImageStreamMap map[int]ImageStreamData

type ImageStreamList struct {
	streams ImageStreamMap
}

func MakeImageStream(data []byte) ImageStreamData {
	isd := ImageStreamData{
		LastUpdate: time.Now().Unix(),
		Data:       data,
		DataType:   "image/jpeg",
	}
	return isd
}

func MakeImageStreamList() ImageStreamList {
	isl := ImageStreamList{
		streams: make(ImageStreamMap),
	}
	return isl
}

func (isl *ImageStreamList) AddImageStream(id int, data []byte) error {
	if _, ok := isl.streams[id]; ok {
		return errors.New("stream already exists")
	}
	isd := MakeImageStream(data)
	isl.streams[id] = isd
	return nil
}

func (isl *ImageStreamList) GetImageStream(id int) (*ImageStreamData, error) {
	if isd, ok := isl.streams[id]; ok {
		return &isd, nil
	} else {
		return nil, errors.New("image stream not found")
	}
}

func (isl *ImageStreamList) GetImageStreamAge(id int) int64 {
	if isd, err := isl.GetImageStream(id); err == nil {
		age := time.Now().Unix() - isd.LastUpdate
		return age
	} else {
		return -1
	}
}

func (isl *ImageStreamList) UpdateImageStream(id int, data []byte) error {
	if isd, ok := isl.streams[id]; ok {
		isd.LastUpdate = time.Now().Unix()
		isd.Data = data
		isl.streams[id] = isd
		return nil
	} else {
		return isl.AddImageStream(id, data)
	}
}
