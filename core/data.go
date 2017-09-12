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

package core

import (
	"sync"

	"github.com/go-ego/riot/types"
)

// DocInfoGroup 文档信息[shard][id]info
var DocInfoGroup = make(map[int]*types.DocInfosShard)
var docInfosGroupRWMutex sync.RWMutex

func AddDocInfosShard(shard int) {
	docInfosGroupRWMutex.Lock()
	defer docInfosGroupRWMutex.Unlock()
	if _, found := DocInfoGroup[shard]; !found {
		DocInfoGroup[shard] = &types.DocInfosShard{
			DocInfos: make(map[uint64]*types.DocInfo),
		}
	}
}

func AddDocInfo(shard int, docId uint64, docinfo *types.DocInfo) {
	docInfosGroupRWMutex.Lock()
	defer docInfosGroupRWMutex.Unlock()
	if _, ok := DocInfoGroup[shard]; !ok {
		DocInfoGroup[shard] = &types.DocInfosShard{
			DocInfos: make(map[uint64]*types.DocInfo),
		}
	}
	DocInfoGroup[shard].DocInfos[docId] = docinfo
	DocInfoGroup[shard].NumDocuments++
}

// func IsDocExist(docId uint64) bool {
// 	docInfosGroupRWMutex.RLock()
// 	defer docInfosGroupRWMutex.RUnlock()
// 	for _, docInfosShard := range DocInfoGroup {
// 		_, found := docInfosShard.DocInfos[docId]
// 		if found {
// 			return true
// 		}
// 	}
// 	return false
// }

// InvertedIndexGroup 反向索引表([shard][关键词]反向索引表)
var InvertedIndexGroup = make(map[int]*types.InvertedIndexShard)
var invertedIndexGroupRWMutex sync.RWMutex

func AddInvertedIndexShard(shard int) {
	invertedIndexGroupRWMutex.Lock()
	defer invertedIndexGroupRWMutex.Unlock()
	if _, found := InvertedIndexGroup[shard]; !found {
		InvertedIndexGroup[shard] = &types.InvertedIndexShard{
			InvertedIndex: make(map[string]*types.KeywordIndices),
		}
	}
}

func AddKeywordIndices(shard int, keyword string, keywordIndices *types.KeywordIndices) {
	invertedIndexGroupRWMutex.Lock()
	defer invertedIndexGroupRWMutex.Unlock()
	if _, ok := InvertedIndexGroup[shard]; !ok {
		InvertedIndexGroup[shard] = &types.InvertedIndexShard{
			InvertedIndex: make(map[string]*types.KeywordIndices),
		}
	}
	InvertedIndexGroup[shard].InvertedIndex[keyword] = keywordIndices
	InvertedIndexGroup[shard].TotalTokenLength++
}
