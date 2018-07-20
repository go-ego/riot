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
)

func logic1() {
	// Init engine
	searcher.Init(types.EngineOpts{
		Using:       4,
		IDOnly:      true,
		NotUseGse: true})
	defer searcher.Close()

	text := "Google Is Experimenting With Virtual Reality Advertising"
	text1 := `Google accidentally pushed Bluetooth update for Home
	speaker early`
	text2 := `Google is testing another Search results layout with 
	rounded cards, new colors, and the 4 mysterious colored dots again`
	text3 := "Google testing text search"

	// Add the document to the index, docId starts at 1
	searcher.Index(1, types.DocData{Content: text})
	searcher.Index(2, types.DocData{Content: text1})
	searcher.Index(3, types.DocData{Content: text2})
	searcher.Index(4, types.DocData{Content: text3})

	// Wait for the index to refresh
	searcher.Flush()

	// var strArr []string
	strArr := []string{"accidentally"}
	// strArr := []string{"text"}
	query := "google testing"

	// The search output format is found in the types.SearchResp structure
	docs := searcher.Search(types.SearchReq{
		Text: query,
		// Search "google testing" segmentation `or relation`
		// and not the result of "accidentally"
		Logic: types.Logic{
			Should: true,
			LogicExpr: types.LogicExpr{
				NotInLabels: strArr,
			},
		},
		RankOpts: &types.RankOpts{
			OutputOffset: 0,
			MaxOutputs:   100,
		}})

	log.Println("search response: ", len(docs.Docs.(types.ScoredIDs)), docs)
}

func logic2() {
	// Init engine
	var searcher1 = riot.New()

	text := "Google Is Experimenting With Virtual Reality Advertising"
	text1 := `Google accidentally pushed Bluetooth update for Home
	speaker early`
	text2 := `Google is testing another Search results layout with 
	rounded cards, new colors, and the 4 mysterious colored dots again`
	text3 := "Google testing text search"
	text4 := "Google testing search"

	// Add the document to the index, docId starts at 1
	searcher1.Index(1, types.DocData{Content: text})
	searcher1.Index(2, types.DocData{Content: text1})
	searcher1.Index(3, types.DocData{Content: text2})
	searcher1.Index(4, types.DocData{Content: text3})
	searcher1.Index(5, types.DocData{Content: text4})

	// Wait for the index to refresh
	searcher1.Flush()

	// var strArr []string
	strArr := []string{"accidentally"}
	notArr := []string{"text"}
	query := "google testing"

	// The search output format is found in the types.SearchResp structure
	docs := searcher1.Search(types.SearchReq{
		Text: query,
		// Search "google testing" segmentation `must relation`
		// and the result of "or accidentally"
		Logic: types.Logic{
			Should: true,
			LogicExpr: types.LogicExpr{
				// ShouldLabels: strArr,
				MustLabels:  strArr,
				NotInLabels: notArr,
			},
		},
		RankOpts: &types.RankOpts{
			OutputOffset: 0,
			MaxOutputs:   100,
		}})

	log.Println("search response: ", docs.NumDocs, docs)
}

func main() {
	logic1()
	logic2()
}
