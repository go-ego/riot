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
	"sync/atomic"

	"github.com/oGre222/tea/types"
)

type indexerAddDocReq struct {
	doc         *types.DocIndex
	forceUpdate bool
}

type indexerLookupReq struct {
	countDocsOnly bool
	tokens        []string
	labels        []string

	docIds           map[string]bool
	options          types.RankOpts
	rankerReturnChan chan rankerReturnReq
	orderless        bool
	logic            types.Logic
}

type indexerRemoveDocReq struct {
	docId       string
	forceUpdate bool
}

func (engine *Engine) indexerAddDoc(shard int) {
	for {
		request := <-engine.indexerAddDocChans[shard]
		engine.indexers[shard].AddDocToCache(request.doc, request.forceUpdate)
		if request.doc != nil {
			atomic.AddUint64(&engine.numTokenIndexAdded,
				uint64(len(request.doc.Keywords)))

			atomic.AddUint64(&engine.numDocsIndexed, 1)
		}
		if request.forceUpdate {
			atomic.AddUint64(&engine.numDocsForceUpdated, 1)
		}
	}
}

func (engine *Engine) indexerRemoveDoc(shard int) {
	for {
		request := <-engine.indexerRemoveDocChans[shard]
		engine.indexers[shard].RemoveDocToCache(request.docId, request.forceUpdate)
		if request.docId != "0" {
			atomic.AddUint64(&engine.numDocsRemoved, 1)
		}
		if request.forceUpdate {
			atomic.AddUint64(&engine.numDocsForceUpdated, 1)
		}
	}
}

func (engine *Engine) orderLess(
	request indexerLookupReq, docs []types.IndexedDoc) {

	if engine.initOptions.IDOnly {
		var outputDocs types.ScoredIDs
		for _, d := range docs {
			outputDocs = append(outputDocs, types.ScoredID{
				DocId:            d.DocId,
				TokenSnippetLocs: d.TokenSnippetLocs,
				TokenLocs:        d.TokenLocs,
			})
		}

		request.rankerReturnChan <- rankerReturnReq{
			docs:    outputDocs,
			numDocs: len(outputDocs),
		}

		return
	}

	var outputDocs types.ScoredDocs
	for _, d := range docs {
		ids := types.ScoredID{
			DocId:            d.DocId,
			TokenSnippetLocs: d.TokenSnippetLocs,
			TokenLocs:        d.TokenLocs,
		}

		outputDocs = append(outputDocs, types.ScoredDoc{
			ScoredID: ids,
		})
	}

	request.rankerReturnChan <- rankerReturnReq{
		docs:    outputDocs,
		numDocs: len(outputDocs),
	}
}

func (engine *Engine) indexerLookup(shard int) {
	for {
		request := <-engine.indexerLookupChans[shard]

		docs, numDocs := engine.indexers[shard].Lookup(
			request.tokens, request.labels,
			request.docIds, request.countDocsOnly, request.logic)

		if request.countDocsOnly {
			request.rankerReturnChan <- rankerReturnReq{numDocs: numDocs}
			continue
		}

		if len(docs) == 0 {
			request.rankerReturnChan <- rankerReturnReq{}
			continue
		}

		if request.orderless {
			// var outputDocs interface{}
			engine.orderLess(request, docs)

			continue
		}

		rankerRequest := rankerRankReq{
			countDocsOnly:    request.countDocsOnly,
			docs:             docs,
			options:          request.options,
			rankerReturnChan: request.rankerReturnChan,
		}
		engine.rankerRankChans[shard] <- rankerRequest
	}
}
