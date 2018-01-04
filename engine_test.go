package riot

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/go-ego/riot/types"
	"github.com/go-ego/riot/utils"
)

type ScoringFields struct {
	A, B, C float32
}

func AddDocs(engine *Engine) {
	docId := uint64(1)
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "中国有十三亿人口人口",
		Fields:  ScoringFields{1, 2, 3},
	})
	docId++
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "中国人口",
		Fields:  nil,
	})
	docId++
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "有人口",
		Fields:  ScoringFields{2, 3, 1},
	})
	docId++
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "有十三亿人口",
		Fields:  ScoringFields{2, 3, 3},
	})
	docId++
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "中国十三亿人口",
		Fields:  ScoringFields{0, 9, 1},
	})
	engine.FlushIndex()
}

func addDocsWithLabels(engine *Engine) {
	docId := uint64(1)
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "此次百度收购将成中国互联网最大并购",
		Labels:  []string{"百度", "中国"},
	})
	docId++
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "百度宣布拟全资收购91无线业务",
		Labels:  []string{"百度"},
	})
	docId++
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "百度是中国最大的搜索引擎",
		Labels:  []string{"百度"},
	})
	docId++
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "百度在研制无人汽车",
		Labels:  []string{"百度"},
	})
	docId++
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "BAT是中国互联网三巨头",
		Labels:  []string{"百度"},
	})
	engine.FlushIndex()
}

type RankByTokenProximity struct {
}

func (rule RankByTokenProximity) Score(
	doc types.IndexedDoc, fields interface{}) []float32 {
	if doc.TokenProximity < 0 {
		return []float32{}
	}
	return []float32{1.0 / (float32(doc.TokenProximity) + 1)}
}

func TestEngineIndexDoc(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchReq{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "中国", outputs.Tokens[0])
	utils.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	utils.Expect(t, "3", len(outDocs))

	log.Println("TestEngineIndexDoc:", outDocs)
	utils.Expect(t, "2", outDocs[0].DocId)
	utils.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	utils.Expect(t, "[0 6]", outDocs[0].TokenSnippetLocs)

	utils.Expect(t, "5", outDocs[1].DocId)
	utils.Expect(t, "100", int(outDocs[1].Scores[0]*1000))
	utils.Expect(t, "[0 15]", outDocs[1].TokenSnippetLocs)

	utils.Expect(t, "1", outDocs[2].DocId)
	utils.Expect(t, "76", int(outDocs[2].Scores[0]*1000))
	utils.Expect(t, "[0 18]", outDocs[2].TokenSnippetLocs)
}

func TestReverseOrder(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchReq{Text: "中国人口"})

	outDocs := outputs.Docs.(types.ScoredDocs)
	utils.Expect(t, "3", len(outDocs))

	utils.Expect(t, "1", outDocs[0].DocId)
	utils.Expect(t, "5", outDocs[1].DocId)
	utils.Expect(t, "2", outDocs[2].DocId)
}

func TestOffsetAndMaxOutputs(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ReverseOrder:    true,
			OutputOffset:    1,
			MaxOutputs:      3,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchReq{Text: "中国人口"})

	outDocs := outputs.Docs.(types.ScoredDocs)
	utils.Expect(t, "2", len(outDocs))

	utils.Expect(t, "5", outDocs[0].DocId)
	utils.Expect(t, "2", outDocs[1].DocId)
}

type TestScoringCriteria struct {
}

func (criteria TestScoringCriteria) Score(
	doc types.IndexedDoc, fields interface{}) []float32 {
	if reflect.TypeOf(fields) != reflect.TypeOf(ScoringFields{}) {
		return []float32{}
	}
	fs := fields.(ScoringFields)
	return []float32{float32(doc.TokenProximity)*fs.A + fs.B*fs.C}
}

func TestSearchWithCriteria(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ScoringCriteria: TestScoringCriteria{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchReq{Text: "中国人口"})

	outDocs := outputs.Docs.(types.ScoredDocs)
	utils.Expect(t, "2", len(outDocs))

	log.Println(outDocs)
	utils.Expect(t, "1", outDocs[0].DocId)
	utils.Expect(t, "18000", int(outDocs[0].Scores[0]*1000))

	utils.Expect(t, "5", outDocs[1].DocId)
	utils.Expect(t, "9000", int(outDocs[1].Scores[0]*1000))
}

func TestCompactIndex(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ScoringCriteria: TestScoringCriteria{},
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchReq{Text: "中国人口"})

	outDocs := outputs.Docs.(types.ScoredDocs)
	utils.Expect(t, "2", len(outDocs))

	utils.Expect(t, "5", outDocs[0].DocId)
	utils.Expect(t, "9000", int(outDocs[0].Scores[0]*1000))

	utils.Expect(t, "1", outDocs[1].DocId)
	utils.Expect(t, "6000", int(outDocs[1].Scores[0]*1000))
}

type BM25ScoringCriteria struct {
}

func (criteria BM25ScoringCriteria) Score(
	doc types.IndexedDoc, fields interface{}) []float32 {
	if reflect.TypeOf(fields) != reflect.TypeOf(ScoringFields{}) {
		return []float32{}
	}
	return []float32{doc.BM25}
}

func TestFrequenciesIndex(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ScoringCriteria: BM25ScoringCriteria{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.FrequenciesIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchReq{Text: "中国人口"})

	outDocs := outputs.Docs.(types.ScoredDocs)
	utils.Expect(t, "2", len(outDocs))

	utils.Expect(t, "1", outDocs[0].DocId)
	utils.Expect(t, "2500", int(outDocs[0].Scores[0]*1000))

	utils.Expect(t, "5", outDocs[1].DocId)
	utils.Expect(t, "1818", int(outDocs[1].Scores[0]*1000))
}

func TestRemoveDoc(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ScoringCriteria: TestScoringCriteria{},
		},
	})

	AddDocs(&engine)
	engine.RemoveDoc(5)
	engine.RemoveDoc(6)
	engine.FlushIndex()
	engine.IndexDoc(6, types.DocIndexData{
		Content: "中国人口有十三亿",
		Fields:  ScoringFields{0, 9, 1},
	})
	engine.FlushIndex()

	outputs := engine.Search(types.SearchReq{Text: "中国人口"})

	outDocs := outputs.Docs.(types.ScoredDocs)
	utils.Expect(t, "2", len(outDocs))

	utils.Expect(t, "6", outDocs[0].DocId)
	utils.Expect(t, "9000", int(outDocs[0].Scores[0]*1000))
	utils.Expect(t, "1", outDocs[1].DocId)
	utils.Expect(t, "6000", int(outDocs[1].Scores[0]*1000))
}

func TestEngineIndexDocWithTokens(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	docId := uint64(1)
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "",
		Tokens: []types.TokenData{
			{"中国", []int{0}},
			{"人口", []int{18, 24}},
		},
		Fields: ScoringFields{1, 2, 3},
	})
	docId++
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "",
		Tokens: []types.TokenData{
			{"中国", []int{0}},
			{"人口", []int{6}},
		},
		Fields: ScoringFields{1, 2, 3},
	})
	docId++
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "中国十三亿人口",
		Fields:  ScoringFields{0, 9, 1},
	})
	engine.FlushIndex()

	outputs := engine.Search(types.SearchReq{Text: "中国人口"})
	log.Println("TestEngineIndexDocWithTokens", outputs)
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "中国", outputs.Tokens[0])
	utils.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	utils.Expect(t, "3", len(outDocs))

	utils.Expect(t, "2", outDocs[0].DocId)
	utils.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	utils.Expect(t, "[0 6]", outDocs[0].TokenSnippetLocs)

	utils.Expect(t, "3", outDocs[1].DocId)
	utils.Expect(t, "100", int(outDocs[1].Scores[0]*1000))
	utils.Expect(t, "[0 15]", outDocs[1].TokenSnippetLocs)

	utils.Expect(t, "1", outDocs[2].DocId)
	utils.Expect(t, "76", int(outDocs[2].Scores[0]*1000))
	utils.Expect(t, "[0 18]", outDocs[2].TokenSnippetLocs)
}

func TestEngineIndexDocWithContentAndLabels(t *testing.T) {
	var engine1, engine2 Engine
	engine1.Init(types.EngineOpts{
		SegmenterDict: "./data/dict/dictionary.txt",
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})
	engine2.Init(types.EngineOpts{
		SegmenterDict: "./data/dict/dictionary.txt",
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.DocIdsIndex,
		},
	})

	addDocsWithLabels(&engine1)
	addDocsWithLabels(&engine2)

	outputs1 := engine1.Search(types.SearchReq{Text: "百度"})
	outputs2 := engine2.Search(types.SearchReq{Text: "百度"})
	utils.Expect(t, "1", len(outputs1.Tokens))
	utils.Expect(t, "1", len(outputs2.Tokens))
	utils.Expect(t, "百度", outputs1.Tokens[0])
	utils.Expect(t, "百度", outputs2.Tokens[0])

	outDocs := outputs1.Docs.(types.ScoredDocs)
	utils.Expect(t, "5", len(outDocs))
	utils.Expect(t, "5", len(outputs2.Docs.(types.ScoredDocs)))
}

func TestEngineIndexDocWithPersistentStorage(t *testing.T) {
	gob.Register(ScoringFields{})
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
		UseStorage:    true,
		StorageFolder: "riot.persistent",
		StorageShards: 2,
	})
	AddDocs(&engine)
	engine.RemoveDoc(5, true)
	engine.Close()

	var engine1 Engine
	engine1.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
		UseStorage:    true,
		StorageFolder: "riot.persistent",
		StorageShards: 2,
	})
	engine1.FlushIndex()

	outputs := engine1.Search(types.SearchReq{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "中国", outputs.Tokens[0])
	utils.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	utils.Expect(t, "2", len(outDocs))

	utils.Expect(t, "2", outDocs[0].DocId)
	utils.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	utils.Expect(t, "[0 6]", outDocs[0].TokenSnippetLocs)

	utils.Expect(t, "1", outDocs[1].DocId)
	utils.Expect(t, "76", int(outDocs[1].Scores[0]*1000))
	utils.Expect(t, "[0 18]", outDocs[1].TokenSnippetLocs)

	engine1.Close()
	os.RemoveAll("riot.persistent")
}

func TestEngineIndexDocWithNewStorage(t *testing.T) {
	gob.Register(ScoringFields{})
	var engine = New("./testdata/test_dict.txt")
	log.Println("engine.............")
	// engine = engine.New()
	AddDocs(engine)
	engine.RemoveDoc(5, true)
	engine.Close()

	var engine1 = New("./testdata/test_dict.txt")
	// engine1 = engine1.New()
	log.Println("test")
	engine1.FlushIndex()
	log.Println("engine1.............")

	outputs := engine1.Search(types.SearchReq{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "中国", outputs.Tokens[0])
	utils.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	utils.Expect(t, "2", len(outDocs))

	// utils.Expect(t, "2", outDocs[0].DocId)
	utils.Expect(t, "0", int(outDocs[0].Scores[0]*1000))
	utils.Expect(t, "[]", outDocs[0].TokenSnippetLocs)

	// utils.Expect(t, "1", outDocs[1].DocId)
	utils.Expect(t, "0", int(outDocs[1].Scores[0]*1000))
	utils.Expect(t, "[]", outDocs[1].TokenSnippetLocs)

	engine1.Close()
	os.RemoveAll("riot-index")
}

func TestCountDocsOnly(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      1,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)
	engine.RemoveDoc(5)
	engine.FlushIndex()

	outputs := engine.Search(types.SearchReq{Text: "中国人口", CountDocsOnly: true})
	// utils.Expect(t, "0", len(outputs.Docs))
	if outputs.Docs == nil {
		utils.Expect(t, "0", 0)
	}
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "2", outputs.NumDocs)
}

func TestDocOnlyID(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		IDOnly:        true,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      1,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)
	engine.RemoveDoc(5)
	engine.FlushIndex()

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	outputs := engine.Search(types.SearchReq{
		Text:   "中国人口",
		DocIds: docIds})

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		utils.Expect(t, "1", len(outDocs))
	}
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "1", outputs.NumDocs)
}

func TestDocRank(t *testing.T) {
	var engine Engine

	rankOpts := types.RankOpts{
		ReverseOrder:    true,
		OutputOffset:    0,
		MaxOutputs:      1,
		ScoringCriteria: &RankByTokenProximity{},
	}
	engine.Init(types.EngineOpts{
		Using:           1,
		IDOnly:          true,
		SegmenterDict:   "./testdata/test_dict.txt",
		DefaultRankOpts: &rankOpts,
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)
	engine.RemoveDoc(5)
	engine.FlushIndex()

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true

	request := types.SearchReq{
		Text:   "中国人口",
		DocIds: docIds}

	tokens := engine.Tokens(request)
	// 建立排序器返回的通信通道
	rankerReturnChan := make(
		chan rankerReturnReq, engine.initOptions.NumShards)

	// 生成查找请求
	lookupRequest := indexerLookupReq{
		countDocsOnly:    request.CountDocsOnly,
		tokens:           tokens,
		labels:           request.Labels,
		docIds:           request.DocIds,
		options:          rankOpts,
		rankerReturnChan: rankerReturnChan,
		orderless:        request.Orderless,
		logic:            request.Logic,
	}

	// 向索引器发送查找请求
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerLookupChans[shard] <- lookupRequest
	}

	outputs := engine.Rank(request, rankOpts, tokens, rankerReturnChan)

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		utils.Expect(t, "1", len(outDocs))
	}
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "1", outputs.NumDocs)
}

func TestDocRanks(t *testing.T) {
	var engine Engine

	rankOpts := types.RankOpts{
		ReverseOrder:    true,
		OutputOffset:    0,
		MaxOutputs:      1,
		ScoringCriteria: &RankByTokenProximity{},
	}
	engine.Init(types.EngineOpts{
		Using:           1,
		SegmenterDict:   "./testdata/test_dict.txt",
		DefaultRankOpts: &rankOpts,
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)
	engine.RemoveDoc(5)
	engine.FlushIndex()

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true

	request := types.SearchReq{
		Text:   "中国人口",
		DocIds: docIds}

	tokens := engine.Tokens(request)
	// 建立排序器返回的通信通道
	rankerReturnChan := make(
		chan rankerReturnReq, engine.initOptions.NumShards)

	// 生成查找请求
	lookupRequest := indexerLookupReq{
		countDocsOnly:    request.CountDocsOnly,
		tokens:           tokens,
		labels:           request.Labels,
		docIds:           request.DocIds,
		options:          rankOpts,
		rankerReturnChan: rankerReturnChan,
		orderless:        request.Orderless,
		logic:            request.Logic,
	}

	// 向索引器发送查找请求
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerLookupChans[shard] <- lookupRequest
	}

	outputs := engine.Ranks(request, rankOpts, tokens, rankerReturnChan)

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredDocs)
		utils.Expect(t, "1", len(outDocs))
	}
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "1", outputs.NumDocs)
}

func TestSearchWithin(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	outputs := engine.Search(types.SearchReq{
		Text:   "中国人口",
		DocIds: docIds,
	})
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "中国", outputs.Tokens[0])
	utils.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	utils.Expect(t, "2", len(outDocs))

	utils.Expect(t, "1", outDocs[0].DocId)
	utils.Expect(t, "76", int(outDocs[0].Scores[0]*1000))
	utils.Expect(t, "[0 18]", outDocs[0].TokenSnippetLocs)

	utils.Expect(t, "5", outDocs[1].DocId)
	utils.Expect(t, "100", int(outDocs[1].Scores[0]*1000))
	utils.Expect(t, "[0 15]", outDocs[1].TokenSnippetLocs)
}

func TestSearchJp(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict_jp.txt",
		DefaultRankOpts: &types.RankOpts{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)

	engine.IndexDoc(6, types.DocIndexData{
		Content: "こんにちは世界, こんにちは",
		Fields:  ScoringFields{1, 2, 3},
	})
	engine.FlushIndex()

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	docIds[6] = true
	outputs := engine.Search(types.SearchReq{
		Text:   "こんにちは世界",
		DocIds: docIds,
	})

	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "こんにちは", outputs.Tokens[0])
	utils.Expect(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	utils.Expect(t, "1", len(outDocs))

	utils.Expect(t, "6", outDocs[0].DocId)
	utils.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	utils.Expect(t, "[0 15]", outDocs[0].TokenSnippetLocs)
}

func TestSearchGse(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		// Using:         1,
		SegmenterDict: "./testdata/test_dict_jp.txt",
		DefaultRankOpts: &types.RankOpts{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)

	engine.IndexDoc(6, types.DocIndexData{
		Content: "こんにちは世界, こんにちは",
		Fields:  ScoringFields{1, 2, 3},
	})

	tokenData := types.TokenData{Text: "こんにちは"}
	tokenDatas := []types.TokenData{tokenData}
	engine.IndexDoc(7, types.DocIndexData{
		Content: "你好世界, hello world!",
		Tokens:  tokenDatas,
		Fields:  ScoringFields{1, 2, 3},
	})
	engine.FlushIndex()

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	docIds[6] = true
	docIds[7] = true
	outputs := engine.Search(types.SearchReq{
		Text:   "こんにちは世界",
		DocIds: docIds,
	})

	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "こんにちは", outputs.Tokens[0])
	utils.Expect(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	utils.Expect(t, "2", len(outDocs))

	utils.Expect(t, "7", outDocs[0].DocId)
	utils.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	utils.Expect(t, "[]", outDocs[0].TokenSnippetLocs)

	utils.Expect(t, "6", outDocs[1].DocId)
	utils.Expect(t, "1000", int(outDocs[1].Scores[0]*1000))
	utils.Expect(t, "[0 15]", outDocs[1].TokenSnippetLocs)
}

func TestSearchLogic(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		SegmenterDict: "./testdata/test_dict_jp.txt",
		DefaultRankOpts: &types.RankOpts{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)

	engine.IndexDoc(6, types.DocIndexData{
		Content: "こんにちは世界, こんにちは",
		Fields:  ScoringFields{1, 2, 3},
	})

	tokenData := types.TokenData{Text: "こんにちは"}
	tokenDatas := []types.TokenData{tokenData}
	engine.IndexDoc(7, types.DocIndexData{
		Content: "你好世界, hello world!",
		Tokens:  tokenDatas,
		Fields:  ScoringFields{1, 2, 3},
	})

	engine.IndexDoc(8, types.DocIndexData{
		Content: "你好世界, hello world!",
		Fields:  ScoringFields{1, 2, 3},
	})

	engine.IndexDoc(9, types.DocIndexData{
		Content: "你好世界, hello!",
		Fields:  ScoringFields{1, 2, 3},
	})

	engine.FlushIndex()

	docIds := make(map[uint64]bool)
	for index := 0; index < 10; index++ {
		docIds[uint64(index)] = true
	}

	strArr := []string{"こんにちは"}
	outputs := engine.Search(types.SearchReq{
		Text:   "こんにちは世界",
		DocIds: docIds,
		Logic: types.Logic{
			Should: true,
			LogicExpr: types.LogicExpr{
				NotInLabels: strArr,
			},
		},
	})

	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "こんにちは", outputs.Tokens[0])
	utils.Expect(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	utils.Expect(t, "2", len(outDocs))

	utils.Expect(t, "9", outDocs[0].DocId)
	utils.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	utils.Expect(t, "[]", outDocs[0].TokenSnippetLocs)

	utils.Expect(t, "8", outDocs[1].DocId)
	utils.Expect(t, "1000", int(outDocs[1].Scores[0]*1000))
	utils.Expect(t, "[]", outDocs[1].TokenSnippetLocs)
}

func TestDocGetAllID(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		UseStorage:    true,
		StorageFolder: "riot.id",
		IDOnly:        true,
		SegmenterDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      1,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	AddDocs(&engine)
	engine.RemoveDoc(5)
	engine.FlushIndex()

	allIds := engine.GetAllDocIds()
	fmt.Println("all id", allIds)
	utils.Expect(t, "4", len(allIds))
	utils.Expect(t, "[3 4 1 2]", allIds)

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	outputs := engine.Search(types.SearchReq{
		Text:   "中国人口",
		DocIds: docIds})

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		utils.Expect(t, "1", len(outDocs))
	}
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "1", outputs.NumDocs)

	os.RemoveAll("riot.id")
}
