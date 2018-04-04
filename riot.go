// Copyright 2018 ego authors
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

package riot

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"encoding/binary"
	"encoding/gob"

	"github.com/go-ego/murmur"
	"github.com/go-ego/riot/core"
	"github.com/go-ego/riot/types"
)

// New create a new engine
func New(dict ...string) *Engine {
	// func (engine *Engine) New(conf com.Config) *Engine{
	var (
		searcher = &Engine{}

		path          = "./riot-index"
		storageShards = 10
		numShards     = 10

		segmenterDict string
	)

	if len(dict) > 0 {
		segmenterDict = dict[0]
	}

	searcher.Init(types.EngineOpts{
		// Using:         using,
		StorageShards: storageShards,
		NumShards:     numShards,
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.DocIdsIndex,
		},
		UseStorage:    true,
		StorageFolder: path,
		// StorageEngine: storageEngine,
		SegmenterDict: segmenterDict,
		// StopTokenFile: stopTokenFile,
	})

	// defer searcher.Close()
	os.MkdirAll(path, 0777)

	// 等待索引刷新完毕
	// searcher.Flush()
	// log.Println("recover index number:", searcher.NumDocsIndexed())

	return searcher
}

// HasDoc doc is exist return true
func (engine *Engine) HasDoc(docId uint64) bool {
	return core.IsDocExist(docId)
}

// DBHasDoc doc is exist return true
func (engine *Engine) DBHasDoc(docId uint64) bool {
	b := make([]byte, 10)
	length := binary.PutUvarint(b, docId)

	shard := murmur.Sum32(fmt.Sprintf("%d", docId)) %
		uint32(engine.initOptions.StorageShards)

	has, err := engine.dbs[shard].Has(b[0:length])
	if err != nil {
		log.Println("engine.dbs[shard].Has(b[0:length]) ", err)
	}

	return has
}

// GetDBAllIds get all the DocId from the storage database and return
// 从数据库遍历所有的 DocId, 并返回
func (engine *Engine) GetDBAllIds() []uint64 {
	docsId := make([]uint64, 0)
	for i := range engine.dbs {
		engine.dbs[i].ForEach(func(k, v []byte) error {
			// fmt.Println(k, v)
			docId, _ := binary.Uvarint(k)
			docsId = append(docsId, docId)
			return nil
		})
	}

	return docsId
}

// GetDBAllDocs get the db all docs
func (engine *Engine) GetDBAllDocs() (
	docsId []uint64, docData []types.DocIndexData) {
	for i := range engine.dbs {
		engine.dbs[i].ForEach(func(key, val []byte) error {
			// fmt.Println(k, v)
			docId, _ := binary.Uvarint(key)
			docsId = append(docsId, docId)

			buf := bytes.NewReader(val)
			dec := gob.NewDecoder(buf)
			var data types.DocIndexData
			err := dec.Decode(&data)
			if err != nil {
				log.Println("dec decode ", err)
			}

			docData = append(docData, data)

			return nil
		})
	}

	return docsId, docData
}

// GetAllDocIds get all the DocId from the storage database and return
// 从数据库遍历所有的 DocId, 并返回
func (engine *Engine) GetAllDocIds() []uint64 {
	return engine.GetDBAllIds()
}

// Try handler(err)
func Try(fun func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			handler(err)
		}
	}()
	fun()
}
