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

	"github.com/go-ego/riot/engine"
	"github.com/go-ego/riot/types"
	// "github.com/go-vgo/gt/zlog"
)

var (
	// Searcher is coroutine safe
	Searcher = engine.Engine{}
	Conf     Config
)

// InitEngine init engine
func InitEngine(conf Config) {
	// os.RemoveAll("./index")
	Conf = conf

	log.Println("conf.config.Etcd", Conf.Etcd)

	var path = "./index"
	if conf.Engine.StorageFolder != "" {
		path = conf.Engine.StorageFolder
	}

	if conf.Engine.StorageFolder != "" {
		path = conf.Engine.StorageFolder
	}

	storageShards := 100
	if conf.Engine.StorageShards != 0 {
		storageShards = conf.Engine.StorageShards
	}

	numShards := 100
	if conf.Engine.NumShards != 0 {
		numShards = conf.Engine.NumShards
	}

	segmenterDict := "../dict/dictionary.txt"
	if conf.Engine.SegmenterDict != "" {
		segmenterDict = conf.Engine.SegmenterDict
	}
	using := conf.Engine.Using

	storageEngine := conf.Engine.StorageEngine
	stopTokenFile := conf.Engine.StopTokenFile

	Searcher.Init(types.EngineInitOptions{
		Using:         using,
		StorageShards: storageShards,
		NumShards:     numShards,
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.DocIdsIndex,
		},
		UseStorage:    true,
		StorageFolder: path,
		StorageEngine: storageEngine,
		SegmenterDict: segmenterDict,
		StopTokenFile: stopTokenFile,
	})

	// defer Searcher.Close()
	os.MkdirAll(path, 0777)

	// 等待索引刷新完毕
	Searcher.FlushIndex()

	log.Println("recover index number:", Searcher.NumDocumentsIndexed())

}

// AddDocInx add index document
func AddDocInx(docId uint64, data types.DocIndexData, forceUpdate bool) {
	Searcher.IndexDocument(docId, data, forceUpdate)

	Searcher.FlushIndex()
}

// SearchArgs search args
type SearchArgs struct {
	Id, Query, Time          string
	OutputOffset, MaxOutputs int
	DocIds                   map[uint64]bool
	Logic                    types.Logic
	// fn                       func(*SearchArgs)
}

// Search search
func Search(sea SearchArgs) types.SearchResponse {

	var docs types.SearchResponse

	docs = Searcher.Search(types.SearchRequest{Text: sea.Query,
		// NotUsingSegmenter: true,
		DocIds: sea.DocIds,
		Logic:  sea.Logic,
		RankOptions: &types.RankOptions{
			OutputOffset: sea.OutputOffset,
			MaxOutputs:   sea.MaxOutputs,
		}})

	return docs
}

// Delete delete document
func Delete(docid uint64, forceUpdate bool) {
	Searcher.RemoveDocument(docid, forceUpdate)
}
