// Copyright 2016 ego authors
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

package main

import (
	"log"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

var (
	// searcher is coroutine safe
	searcher = riot.Engine{}

	data = types.DocData{
		Content: "Google Is Experimenting With Virtual Reality Advertising",
	}

	data1 = types.DocData{
		Content: `Google accidentally pushed Bluetooth update for Home
	speaker early`,
	}

	data2 = types.DocData{
		Content: `Google is testing another Search results layout with
	rounded cards, new colors, and the 4 mysterious colored dots again`,
	}

	data3 = types.DocData{
		Content: "Google testing text search",
	}

	rankOpts = types.RankOpts{
		OutputOffset: 0,
		MaxOutputs:   100,
	}
)

func addDocs(search *riot.Engine) {
	// Add the document to the index, docId starts at 1
	search.Index("1", data)
	search.Index("2", data1)
	search.Index("3", data2)
	search.Index("4", data3)

	// Wait for the index to refresh
	search.Flush()
}

func logic1() {
	// Init engine
	searcher.Init(types.EngineOpts{
		// Using:     4,
		IDOnly:    true,
		NotUseGse: true,
	})
	defer searcher.Close()

	addDocs(&searcher)

	// var strArr []string
	strArr := []string{"accidentally"}
	// strArr := []string{"text"}
	query := "google testing"

	// Search "google testing" segmentation `or relation`
	// and not the result of "accidentally"
	logic := types.Logic{
		Should: true,
		Expr: types.Expr{
			NotIn: strArr,
		},
	}

	// The search output format is found in the types.SearchResp structure
	docs := searcher.SearchID(types.SearchReq{
		Text:     query,
		Logic:    logic,
		RankOpts: &rankOpts,
	})

	log.Println("search response: ", len(docs.Docs), docs)
}

func logic2() {
	// Init engine
	var searcher1 = riot.New()

	data4 := types.DocData{Content: "Google testing search"}

	// Add the document to the index, docId starts at 1
	addDocs(searcher1)
	searcher1.Index("5", data4)

	// Wait for the index to refresh
	searcher1.Flush()

	// var strArr []string
	strArr := []string{"accidentally"}
	notArr := []string{"text"}
	query := "google testing"

	// Search "google testing" segmentation `must relation`
	// and the result of "or accidentally"
	logic := types.Logic{
		Should: true,
		Expr: types.Expr{
			// Should: strArr,
			Must:  strArr,
			NotIn: notArr,
		},
	}

	// The search output format is found in the types.SearchResp structure
	docs := searcher1.Search(types.SearchReq{
		Text:     query,
		Logic:    logic,
		RankOpts: &rankOpts,
	})

	log.Println("search response: ", docs.NumDocs, docs)
}

func main() {
	logic1()
	logic2()
}
