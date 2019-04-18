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

package com

import (
	"log"
	"os"

	"github.com/oGre222/tea"
	"github.com/oGre222/tea/types"
	// "github.com/go-vgo/gt/zlog"
)

var (
	// Searcher is coroutine safe
	Searcher = riot.Engine{}
	// Conf is config
	Conf Config
)

// InitEngine init engine
func InitEngine(conf Config) {
	// os.RemoveAll("./riot-index")
	Conf = conf

	log.Println("conf.config.Etcd: ", Conf.Etcd)

	var path = "./riot-index"
	if conf.Engine.StoreFolder != "" {
		path = conf.Engine.StoreFolder
	}

	if conf.Engine.StoreFolder != "" {
		path = conf.Engine.StoreFolder
	}

	storageShards := 10
	if conf.Engine.StoreShards != 0 {
		storageShards = conf.Engine.StoreShards
	}

	numShards := 10
	if conf.Engine.NumShards != 0 {
		numShards = conf.Engine.NumShards
	}

	// var segDict string
	segmentDict := "../dict/dictionary.txt"
	if conf.Engine.GseDict != "" {
		segmentDict = conf.Engine.GseDict
	}
	using := conf.Engine.Using

	storageEngine := conf.Engine.StoreEngine
	stopTokenFile := conf.Engine.StopTokenFile

	Searcher.Init(types.EngineOpts{
		Using:       using,
		StoreShards: storageShards,
		NumShards:   numShards,
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.DocIdsIndex,
		},
		UseStore:      true,
		StoreFolder:   path,
		StoreEngine:   storageEngine,
		GseDict:       segmentDict,
		StopTokenFile: stopTokenFile,
	})

	// defer Searcher.Close()
	os.MkdirAll(path, 0777)

	// 等待索引刷新完毕
	Searcher.Flush()

	log.Println("recover index number: ", Searcher.NumDocsIndexed())

}

// AddDocInx add index document
func AddDocInx(docId string, data types.DocData, forceUpdate bool) {
	Searcher.Index(docId, data, forceUpdate)
	// Searcher.Flush()
}

// Flush flsuh the Searcher engine
func Flush() {
	Searcher.Flush()
}

// SearchArgs search args
type SearchArgs struct {
	Id, Query, Time          string
	OutputOffset, MaxOutputs int
	DocIds                   map[string]bool
	Logic                    types.Logic
	// fn                       func(*SearchArgs)
}

// Search search
func Search(sea SearchArgs) types.SearchResp {

	var docs types.SearchResp

	docs = Searcher.Search(types.SearchReq{
		Text: sea.Query,
		// NotUseGse: true,
		DocIds: sea.DocIds,
		Logic:  sea.Logic,
		RankOpts: &types.RankOpts{
			OutputOffset: sea.OutputOffset,
			MaxOutputs:   sea.MaxOutputs,
		}})

	return docs
}

// Delete delete document
func Delete(docid string, forceUpdate bool) {
	Searcher.RemoveDoc(docid, forceUpdate)
}
