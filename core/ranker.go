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

package core

import (
	"log"
	"sort"
	"sync"

	"github.com/go-ego/riot/types"
	"github.com/go-ego/riot/utils"
)

// Ranker ranker
type Ranker struct {
	lock struct {
		sync.RWMutex

		fields map[string]interface{}
		docs   map[string]bool
		// new
		content map[string]string
		attri   map[string]interface{}
	}

	idOnly      bool
	initialized bool
}

// Init init ranker
func (ranker *Ranker) Init(onlyID ...bool) {
	if ranker.initialized == true {
		log.Fatal("The Ranker can not be initialized twice.")
	}
	ranker.initialized = true

	if len(onlyID) > 0 {
		ranker.idOnly = onlyID[0]
	}

	ranker.lock.fields = make(map[string]interface{})
	ranker.lock.docs = make(map[string]bool)

	if !ranker.idOnly {
		// new
		ranker.lock.content = make(map[string]string)
		ranker.lock.attri = make(map[string]interface{})
	}
}

// AddDoc add doc
// 给某个文档添加评分字段
func (ranker *Ranker) AddDoc(
	// docId uint64, fields interface{}, content string, attri interface{}) {
	docId string, fields interface{}, content ...interface{}) {
	if ranker.initialized == false {
		log.Fatal("The Ranker has not been initialized.")
	}

	ranker.lock.Lock()
	ranker.lock.fields[docId] = fields
	ranker.lock.docs[docId] = true

	if !ranker.idOnly {
		// new
		if len(content) > 0 {
			ranker.lock.content[docId] = content[0].(string)
		}

		if len(content) > 1 {
			ranker.lock.attri[docId] = content[1]
			// ranker.lock.attri[docId] = attri
		}
	}

	ranker.lock.Unlock()
}

// RemoveDoc 删除某个文档的评分字段
func (ranker *Ranker) RemoveDoc(docId string) {
	if ranker.initialized == false {
		log.Fatal("The Ranker has not been initialized.")
	}

	ranker.lock.Lock()
	delete(ranker.lock.fields, docId)
	delete(ranker.lock.docs, docId)

	if !ranker.idOnly {
		// new
		delete(ranker.lock.content, docId)
		delete(ranker.lock.attri, docId)
	}

	ranker.lock.Unlock()
}

func maxOutput(options types.RankOpts, docsLen int) (int, int) {
	var start, end int
	if options.MaxOutputs != 0 {
		start = utils.MinInt(options.OutputOffset, docsLen)
		end = utils.MinInt(options.OutputOffset+options.MaxOutputs, docsLen)
		return start, end
	}

	start = utils.MinInt(options.OutputOffset, docsLen)
	end = docsLen
	return start, end
}

func (ranker *Ranker) rankOutIDs(docs []types.IndexedDoc, options types.RankOpts,
	countDocsOnly bool) (outputDocs types.ScoredIDs, numDocs int) {
	for _, d := range docs {
		ranker.lock.RLock()
		// 判断 doc 是否存在
		if _, ok := ranker.lock.docs[d.DocId]; ok {

			fs := ranker.lock.fields[d.DocId]
			ranker.lock.RUnlock()

			// 计算评分并剔除没有分值的文档
			scores := options.ScoringCriteria.Score(d, fs)
			if len(scores) > 0 {
				if !countDocsOnly {
					outputDocs = append(outputDocs,
						types.ScoredID{
							DocId:            d.DocId,
							Scores:           scores,
							TokenSnippetLocs: d.TokenSnippetLocs,
							TokenLocs:        d.TokenLocs,
						})
				}
				numDocs++
			}
		} else {
			ranker.lock.RUnlock()
		}
	}

	return
}

// RankDocID rank docs by types.ScoredIDs
func (ranker *Ranker) RankDocID(docs []types.IndexedDoc,
	options types.RankOpts, countDocsOnly bool) (types.ScoredIDs, int) {

	outputDocs, numDocs := ranker.rankOutIDs(docs, options, countDocsOnly)

	// 排序
	if !countDocsOnly {
		if options.ReverseOrder {
			sort.Sort(sort.Reverse(outputDocs))
		} else {
			sort.Sort(outputDocs)
		}
		// 当用户要求只返回部分结果时返回部分结果
		docsLen := len(outputDocs)
		start, end := maxOutput(options, docsLen)

		return outputDocs[start:end], numDocs
	}

	return outputDocs, numDocs
}

func (ranker *Ranker) rankOutDocs(docs []types.IndexedDoc, options types.RankOpts,
	countDocsOnly bool) (outputDocs types.ScoredDocs, numDocs int) {
	for _, d := range docs {
		ranker.lock.RLock()
		// 判断 doc 是否存在
		if _, ok := ranker.lock.docs[d.DocId]; ok {

			fs := ranker.lock.fields[d.DocId]
			content := ranker.lock.content[d.DocId]
			attri := ranker.lock.attri[d.DocId]
			ranker.lock.RUnlock()

			// 计算评分并剔除没有分值的文档
			scores := options.ScoringCriteria.Score(d, fs)
			if len(scores) > 0 {
				if !countDocsOnly {
					scoredID := types.ScoredID{
						DocId:            d.DocId,
						Scores:           scores,
						TokenSnippetLocs: d.TokenSnippetLocs,
						TokenLocs:        d.TokenLocs,
					}

					outputDocs = append(outputDocs,
						types.ScoredDoc{
							ScoredID: scoredID,
							// new
							Fields:  fs,
							Content: content,
							Attri:   attri,
						})
				}
				numDocs++
			}
		} else {
			ranker.lock.RUnlock()
		}
	}

	return
}

// RankDocs rank docs by types.ScoredDocs
func (ranker *Ranker) RankDocs(docs []types.IndexedDoc,
	options types.RankOpts, countDocsOnly bool) (types.ScoredDocs, int) {

	outputDocs, numDocs := ranker.rankOutDocs(docs, options, countDocsOnly)

	// 排序
	if !countDocsOnly {
		if options.ReverseOrder {
			sort.Sort(sort.Reverse(outputDocs))
		} else {
			sort.Sort(outputDocs)
		}
		// 当用户要求只返回部分结果时返回部分结果
		docsLen := len(outputDocs)
		start, end := maxOutput(options, docsLen)

		return outputDocs[start:end], numDocs
	}

	return outputDocs, numDocs
}

// Rank rank docs
// 给文档评分并排序
func (ranker *Ranker) Rank(docs []types.IndexedDoc,
	options types.RankOpts, countDocsOnly bool) (interface{}, int) {

	if ranker.initialized == false {
		log.Fatal("The Ranker has not been initialized.")
	}

	// 对每个文档评分
	if ranker.idOnly {
		outputDocs, numDocs := ranker.RankDocID(docs, options, countDocsOnly)
		return outputDocs, numDocs
	}

	outputDocs, numDocs := ranker.RankDocs(docs, options, countDocsOnly)
	return outputDocs, numDocs
}
