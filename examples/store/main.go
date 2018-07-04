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
	"fmt"
	"log"
	"os"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

var (
	// searcher is coroutine safe
	searcher = riot.Engine{}
)

func initEngine() {
	var path = "./riot-index"

	searcher.Init(types.EngineOpts{
		Using: 1,
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.DocIdsIndex,
		},
		UseStorage:    true,
		StorageFolder: path,
		StorageEngine: "bg", // bg: badger, lbd: leveldb, bolt: bolt
		// GseDict: "../../data/dict/dictionary.txt",
		GseDict:       "../../testdata/test_dict.txt",
		StopTokenFile: "../../data/dict/stop_tokens.txt",
	})
	defer searcher.Close()
	os.MkdirAll(path, 0777)

	text := "Google Is Experimenting With Virtual Reality Advertising"
	text1 := `Google accidentally pushed Bluetooth update for Home
	speaker early`
	text2 := `Google is testing another Search results layout with 
	rounded cards, new colors, and the 4 mysterious colored dots again`

	// Add the document to the index, docId starts at 1
	searcher.Index(1, types.DocData{Content: text})
	searcher.Index(2, types.DocData{Content: text1})
	searcher.Index(3, types.DocData{Content: text2})
	searcher.Index(5, types.DocData{Content: text2})

	// Wait for the index to refresh
	// searcher.Flush()

	searcher.RemoveDoc(5)
	searcher.Flush()

	log.Println("recover index number: ", searcher.NumDocsIndexed())
}

func restoreIndex() {
	var path = "./riot-index"

	searcher.Init(types.EngineOpts{
		Using: 1,
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.DocIdsIndex,
		},
		UseStorage:    true,
		StorageFolder: path,
		StorageEngine: "bg", // bg: badger, lbd: leveldb, bolt: bolt
		GseDict:       "../../data/dict/dictionary.txt",
		StopTokenFile: "../../data/dict/stop_tokens.txt",
	})
	defer searcher.Close()
	os.MkdirAll(path, 0777)

	// Wait for the index to refresh
	searcher.Flush()

	log.Println("recover index number: ", searcher.NumDocsIndexed())
}

func main() {
	initEngine()

	sea := searcher.Search(types.SearchReq{
		Text: "google testing",
		RankOpts: &types.RankOpts{
			OutputOffset: 0,
			MaxOutputs:   100,
		}})

	fmt.Println("search response: ", sea, "; docs = ", sea.Docs)

	// os.RemoveAll("riot-index")
}
