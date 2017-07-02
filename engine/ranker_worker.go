package engine

import (
	"github.com/go-ego/gwk/types"
)

type rankerAddDocRequest struct {
	docId  uint64
	fields interface{}
	// new
	content string
	// new 属性
	attri interface{}
}

type rankerRankRequest struct {
	docs                []types.IndexedDocument
	options             types.RankOptions
	rankerReturnChannel chan rankerReturnRequest
	countDocsOnly       bool
}

type rankerReturnRequest struct {
	docs    types.ScoredDocuments
	numDocs int
}

type rankerRemoveDocRequest struct {
	docId uint64
}

func (engine *Engine) rankerAddDocWorker(shard int) {
	for {
		request := <-engine.rankerAddDocChannels[shard]
		// engine.rankers[shard].AddDoc(request.docId, request.fields)
		engine.rankers[shard].AddDoc(request.docId, request.fields, request.content, request.attri)
	}
}

func (engine *Engine) rankerRankWorker(shard int) {
	for {
		request := <-engine.rankerRankChannels[shard]
		if request.options.MaxOutputs != 0 {
			request.options.MaxOutputs += request.options.OutputOffset
		}
		request.options.OutputOffset = 0
		outputDocs, numDocs := engine.rankers[shard].Rank(request.docs, request.options, request.countDocsOnly)
		request.rankerReturnChannel <- rankerReturnRequest{docs: outputDocs, numDocs: numDocs}
	}
}

func (engine *Engine) rankerRemoveDocWorker(shard int) {
	for {
		request := <-engine.rankerRemoveDocChannels[shard]
		engine.rankers[shard].RemoveDoc(request.docId)
	}
}
