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
	"os"

	"github.com/go-ego/riot/engine"
	"github.com/go-ego/riot/types"
)

var (
	// searcher是协程安全的
	searcher = engine.Engine{}
)

func initEngine() {
	var path = "./index"

	searcher.Init(types.EngineInitOptions{
		// Using: 1,
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.DocIdsIndex,
		},
		UseStorage:    true,
		StorageFolder: path,
		SegmenterDict: "../../data/dict/dictionary.txt",
		// StopTokenFile:           "../../riot/data/dict/stop_tokens.txt",
	})
	defer searcher.Close()
	os.MkdirAll(path, 0777)

	tokens := searcher.PinYin("在路上, in the way")

	fmt.Println("tokens...", tokens)
	var tokenDatas []types.TokenData
	// tokens := []string{"z", "zl"}
	for i := 0; i < len(tokens); i++ {
		tokenData := types.TokenData{Text: tokens[i]}
		tokenDatas = append(tokenDatas, tokenData)
	}

	searcher.IndexDocument(10, types.DocIndexData{Tokens: tokenDatas, Fields: "在路上"}, false)
	searcher.IndexDocument(11, types.DocIndexData{Content: "在路上, in the way", Tokens: tokenDatas}, false)

	// 等待索引刷新完毕
	searcher.FlushIndex()

}

func main() {
	initEngine()

	sea := searcher.Search(types.SearchRequest{
		Text: "zl",
		RankOptions: &types.RankOptions{
			OutputOffset: 0,
			MaxOutputs:   100,
		}})

	fmt.Println("search---------", sea, "; docs=", sea.Docs)
}
