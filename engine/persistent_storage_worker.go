// Copyright 2013 Hui Chen
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

package engine

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"sync/atomic"

	"github.com/go-ego/riot/types"
)

type storageIndexDocRequest struct {
	docId uint64
	data  types.DocIndexData
	// data        types.DocumentIndexData
}

func (engine *Engine) storageIndexDocWorker(shard int) {
	for {
		request := <-engine.storageIndexDocChannels[shard]

		// 得到key
		b := make([]byte, 10)
		length := binary.PutUvarint(b, request.docId)

		// 得到value
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(request.data)
		if err != nil {
			atomic.AddUint64(&engine.numDocumentsStored, 1)
			continue
		}

		// 将key-value写入数据库
		engine.dbs[shard].Set(b[0:length], buf.Bytes())
		atomic.AddUint64(&engine.numDocumentsStored, 1)
	}
}

func (engine *Engine) storageRemoveDocWorker(docId uint64, shard uint32) {
	// 得到key
	b := make([]byte, 10)
	length := binary.PutUvarint(b, docId)

	// 从数据库删除该key
	engine.dbs[shard].Delete(b[0:length])
}

// storageInitWorker persistent Storage init worker
func (engine *Engine) storageInitWorker(shard int) {
	engine.dbs[shard].ForEach(func(k, v []byte) error {
		key, value := k, v
		// 得到docID
		docId, _ := binary.Uvarint(key)

		// 得到data
		buf := bytes.NewReader(value)
		dec := gob.NewDecoder(buf)
		var data types.DocIndexData
		err := dec.Decode(&data)
		if err == nil {
			// 添加索引
			engine.internalIndexDocument(docId, data, false)
		}
		return nil
	})
	engine.persistentStorageInitChannel <- true
}
