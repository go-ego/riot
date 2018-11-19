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

package riot

import (
	"bytes"

	"encoding/gob"
	"sync/atomic"

	"github.com/go-ego/riot/types"
)

type storeIndexDocReq struct {
	docId string
	data  types.DocData
	// data        types.DocumentIndexData
}

func (engine *Engine) storeIndexDoc(shard int) {
	for {
		request := <-engine.storeIndexDocChans[shard]

		// 得到 key
		b := []byte(request.docId)

		// 得到 value
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(request.data)
		if err != nil {
			atomic.AddUint64(&engine.numDocsStored, 1)
			continue
		}

		// has, err := engine.dbs[shard].Has(b[0:length])
		// if err != nil {
		// 	log.Println("engine.dbs[shard].Has(b[0:length]) ", err)
		// }

		// if has {
		// 	engine.dbs[shard].Delete(b[0:length])
		// }

		// 将 key-value 写入数据库
		engine.dbs[shard].Set(b, buf.Bytes())

		atomic.AddUint64(&engine.numDocsStored, 1)
	}
}

func (engine *Engine) storeRemoveDoc(docId string, shard uint32) {
	// 得到 key
	b := []byte(docId)
	// 从数据库删除该key
	engine.dbs[shard].Delete(b)
}

// storeInit persistent storage init worker
func (engine *Engine) storeInit(shard int) {
	engine.dbs[shard].ForEach(func(k, v []byte) error {
		key, value := k, v
		// 得到docID
		docId := string(key)

		// 得到 data
		buf := bytes.NewReader(value)
		dec := gob.NewDecoder(buf)
		var data types.DocData
		err := dec.Decode(&data)
		if err == nil {
			// 添加索引
			engine.internalIndexDoc(docId, data, false)
		}
		return nil
	})
	engine.storeInitChan <- true
}
