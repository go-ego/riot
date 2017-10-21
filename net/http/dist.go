// Copyright 2017 ego authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/go-ego/riot/net/com"
)

var wgdata [][]byte
var wg sync.WaitGroup

// WgPost post
func WgPost(url string, param url.Values) {

	data, err := Post(url, param)

	if string(data) == "" || err != nil {
		log.Println("data is null ...")
		defer wg.Done()
		return
	}

	wgdata = append(wgdata, data)

	defer wg.Done()
}

// WgDist dist
func WgDist(w http.ResponseWriter, req *http.Request) {
	wgdata = nil
	var distData []byte

	userid := req.URL.Query().Get("userid")
	query := req.URL.Query().Get("query")
	outputOffset := req.URL.Query().Get("outputOffset")
	maxOutputs := req.URL.Query().Get("maxOutputs")
	atime := req.URL.Query().Get("time")

	maxOuts, _ := strconv.Atoi(maxOutputs)
	if maxOuts == 0 {
		maxOuts = config.Engine.MaxOutputs
		outputOffset = strconv.Itoa(config.Engine.OutputOffset)
		maxOutputs = strconv.Itoa(config.Engine.MaxOutputs)
	}

	param := url.Values{}
	param.Set("userid", userid)
	param.Set("query", query)
	param.Set("outputOffset", outputOffset)
	param.Set("maxOutputs", maxOutputs)
	param.Set("time", atime)

	config = com.Conf
	for i := 0; i < len(config.Url); i++ {
		wg.Add(1)
		url := config.Url[i] + "/search"
		go WgPost(url, param)
	}

	wg.Wait()

	if len(wgdata) == 1 {
		distData = wgdata[0]
	} else {
		var docs []Text
		for i := 0; i < len(wgdata); i++ {
			var jsonRes JsonResponse
			json.Unmarshal(wgdata[i], &jsonRes)
			for d := 0; d < len(jsonRes.Docs); d++ {
				docs = append(docs, jsonRes.Docs[d])
			}
		}
		sort.Sort(docsSlice(docs))

		if len(docs) > maxOuts {
			end := maxOuts - 1
			docs = docs[0:end]
		}

		timestamp := time.Now().Unix()
		response, _ := json.Marshal(&JsonResponse{
			Len:       len(docs),
			Timestamp: timestamp,
			Docs:      docs})

		distData = response
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	io.WriteString(w, string(distData))
}
