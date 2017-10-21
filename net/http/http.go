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
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/go-ego/riot/net/com"
	"github.com/go-ego/riot/types"
)

var (
	// config   = com.Conf
	config   com.Config
	searcher = com.Searcher
)

// Post http, params is url.Values type
func Post(apiUrl string, params url.Values) (rs []byte, err error) {
	c := &http.Client{
		Timeout: 1000 * time.Millisecond,
	}

	resp, err := c.PostForm(apiUrl, params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// Search search for documents
func Search(w http.ResponseWriter, req *http.Request) {
	var (
		userid       string
		query        string
		atime        string
		outputOffset int
		maxOutputs   int
	)

	req.ParseForm()
	if len(req.Form["userid"]) > 0 {
		userid = req.Form["userid"][0]
	}

	if len(req.Form["query"]) > 0 {
		query = req.Form["query"][0]
	}

	if len(req.Form["outputOffset"]) > 0 {
		outputOffset, _ = strconv.Atoi(req.Form["outputOffset"][0])
	}

	if len(req.Form["maxOutputs"]) > 0 {
		maxOutputs, _ = strconv.Atoi(req.Form["maxOutputs"][0])
	}

	if len(req.Form["time"]) > 0 {
		atime = req.Form["time"][0]
	}

	config = com.Conf
	log.Println("config", config, "; com.Conf", com.Conf)
	if maxOutputs == 0 {
		outputOffset = config.Engine.OutputOffset
		maxOutputs = config.Engine.MaxOutputs
	}

	// searcher.FlushIndex() /// todo

	sea := com.SearchArgs{
		Id:           userid,
		Query:        query,
		Time:         atime,
		OutputOffset: outputOffset,
		MaxOutputs:   maxOutputs,
	}
	docs := com.Search(sea)

	var textArr []Text
	for i := 0; i < len(docs.Docs); i++ {
		text := Text{
			Id:      docs.Docs[i].DocId,
			Content: docs.Docs[i].Content,
			Score:   docs.Docs[i].Scores,
			Attri:   docs.Docs[i].Attri.(types.Attri),
		}
		textArr = append(textArr, text)
	}

	sort.Sort(docsSlice(textArr))

	if len(textArr) > maxOutputs {
		textArr = textArr[0:maxOutputs]
	}

	log.Println("len...", len(textArr))

	timestamp := time.Now().Unix()
	response, _ := json.Marshal(&JsonResponse{
		Len:       len(textArr),
		Timestamp: timestamp,
		Docs:      textArr})

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	io.WriteString(w, string(response))
}

// AddIndex add search engine index
func AddIndex(w http.ResponseWriter, req *http.Request) {
	var (
		docid string
		query string
	)

	req.ParseForm()
	if len(req.Form["docid"]) > 0 {
		docid = req.Form["docid"][0]
	}

	if len(req.Form["query"]) > 0 {
		query = req.Form["query"][0]
	}

	timeFormat := "2006-01-02 15:04:05"

	attri := types.Attri{
		Time: time.Now().Format(timeFormat),
		Ts:   time.Now().UnixNano(),
	}

	inxid, _ := strconv.ParseUint(docid, 10, 64)
	com.AddDocInx(inxid, types.DocIndexData{Content: query, Attri: attri}, false)

	timestamp := time.Now().Unix()
	response, _ := json.Marshal(&JsonResponse{
		Timestamp: timestamp,
		Docs:      nil})

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	io.WriteString(w, string(response))
}

// DelIndex remove search engine index
func DelIndex(w http.ResponseWriter, req *http.Request) {
	docid := req.URL.Query().Get("docid")

	// docid := string(indexid)
	inxId, _ := strconv.ParseUint(docid, 10, 64)
	com.Delete(inxId, false)
}
