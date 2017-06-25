package core

import (
	"sync"

	"github.com/go-ego/gwk/types"
)

// 文档信息[shard][id]info
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

// 反向索引表([shard][关键词]反向索引表)
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
