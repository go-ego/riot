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
	engine.Flush()
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
	engine.Flush()
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

func TestGetVer(t *testing.T) {
	ver := GetVersion()
	utils.Expect(t, version, ver)
}

func TestTry(t *testing.T) {
	var arr []int

	Try(func() {
		fmt.Println(arr[2])
	}, func(err interface{}) {
		log.Println("err", err)
		utils.Expect(t, "runtime error: index out of range", err)
	})
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
	engine.Flush()
	engine.IndexDoc(6, types.DocIndexData{
		Content: "中国人口有十三亿",
		Fields:  ScoringFields{0, 9, 1},
	})
	engine.Flush()

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

func TestIndexDocWithLabelsStopTokenFile(t *testing.T) {
	var engine1 Engine

	engine1.Init(types.EngineOpts{
		SegmenterDict: "./data/dict/dictionary.txt",
		StopTokenFile: "./testdata/test_stop_dict.txt",
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	addDocsWithLabels(&engine1)

	outputs1 := engine1.Search(types.SearchReq{Text: "百度"})
	utils.Expect(t, "0", len(outputs1.Tokens))
	// utils.Expect(t, "百度", outputs1.Tokens[0])

	outDocs := outputs1.Docs.(types.ScoredDocs)
	utils.Expect(t, "0", len(outDocs))
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
	engine1.Flush()

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
	engine1.Flush()
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
	engine.Flush()

	outputs := engine.Search(types.SearchReq{Text: "中国人口",
		CountDocsOnly: true})
	// utils.Expect(t, "0", len(outputs.Docs))
	if outputs.Docs == nil {
		utils.Expect(t, "0", 0)
	}
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "2", outputs.NumDocs)
}

func TestDocOrderless(t *testing.T) {
	var engine, engine1 Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		SegmenterDict: "./testdata/test_dict.txt",
	})

	AddDocs(&engine)
	engine.RemoveDoc(5)
	engine.Flush()

	outputs := engine.Search(types.SearchReq{Text: "中国人口",
		Orderless: true})
	// utils.Expect(t, "0", len(outputs.Docs))
	if outputs.Docs == nil {
		utils.Expect(t, "0", 0)
	}
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "2", outputs.NumDocs)

	engine1.Init(types.EngineOpts{
		Using:         1,
		IDOnly:        true,
		SegmenterDict: "./testdata/test_dict.txt",
	})

	AddDocs(&engine1)
	engine1.RemoveDoc(5)
	engine1.Flush()

	outputs1 := engine1.Search(types.SearchReq{Text: "中国人口",
		Orderless: true})
	if outputs1.Docs == nil {
		utils.Expect(t, "0", 0)
	}

	utils.Expect(t, "2", len(outputs1.Tokens))
	utils.Expect(t, "2", outputs1.NumDocs)
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
	engine.Flush()

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

	outputs1 := engine.Search(types.SearchReq{
		Text:    "中国人口",
		Timeout: 10,
		DocIds:  docIds})

	if outputs1.Docs != nil {
		outDocs1 := outputs.Docs.(types.ScoredIDs)
		utils.Expect(t, "1", len(outDocs1))
	}
	utils.Expect(t, "2", len(outputs1.Tokens))
	utils.Expect(t, "1", outputs1.NumDocs)
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
	engine.Flush()

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
	engine.Flush()

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

	outputs1 := engine.Search(types.SearchReq{
		Text:    "中国人口",
		Timeout: 10,
		DocIds:  docIds})

	if outputs1.Docs != nil {
		outDocs1 := outputs.Docs.(types.ScoredDocs)
		utils.Expect(t, "1", len(outDocs1))
	}
	utils.Expect(t, "2", len(outputs1.Tokens))
	utils.Expect(t, "1", outputs1.NumDocs)
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
	engine.Flush()

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
	engine.Flush()

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

func TestSearchNotUseGse(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:       4,
		NotUsingGse: true,
	})

	AddDocs(&engine)

	engine.IndexDoc(6, types.DocIndexData{
		Content: "Google Is Experimenting With Virtual Reality Advertising",
		Fields:  ScoringFields{1, 2, 3},
	})

	engine.Flush()

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	docIds[6] = true
	outputs := engine.Search(types.SearchReq{
		Text:   "google is",
		DocIds: docIds,
	})

	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "google", outputs.Tokens[0])
	utils.Expect(t, "is", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	utils.Expect(t, "1", len(outDocs))

	utils.Expect(t, "6", outDocs[0].DocId)
	utils.Expect(t, "3200", int(outDocs[0].Scores[0]*1000))
	utils.Expect(t, "[]", outDocs[0].TokenSnippetLocs)

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

	engine.Flush()

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

func TestDocGetAllDocAndID(t *testing.T) {
	gob.Register(ScoringFields{})

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
	engine.Flush()

	allIds := engine.GetDBAllIds()
	fmt.Println("all id", allIds)
	utils.Expect(t, "4", len(allIds))
	utils.Expect(t, "[3 4 1 2]", allIds)

	allIds = engine.GetAllDocIds()
	fmt.Println("all doc id", allIds)
	utils.Expect(t, "4", len(allIds))
	utils.Expect(t, "[3 4 1 2]", allIds)

	ids, docs := engine.GetDBAllDocs()
	fmt.Println("all id and doc", allIds, docs)
	utils.Expect(t, "4", len(ids))
	utils.Expect(t, "4", len(docs))
	utils.Expect(t, "[3 4 1 2]", ids)
	allDoc := `[{有人口 <nil> [] [] {2 3 1}} {有十三亿人口 <nil> [] [] {2 3 3}} {中国有十三亿人口人口 <nil> [] [] {1 2 3}} {中国人口 <nil> [] [] <nil>}]`
	utils.Expect(t, allDoc, docs)

	has := engine.DBHasDoc(5)
	utils.Expect(t, "false", has)

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

	engine.Close()
	os.RemoveAll("riot.id")
}

func TestDocPinYin(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		// Using:         1,
		UseStorage:    true,
		StorageFolder: "riot.py",
		IDOnly:        true,
		SegmenterDict: "./testdata/test_dict.txt",
	})

	// AddDocs(&engine)
	// engine.RemoveDoc(5)

	tokens := engine.PinYin("在路上, in the way")

	fmt.Println("tokens...", tokens)
	utils.Expect(t, "52", len(tokens))

	var tokenDatas []types.TokenData
	// tokens := []string{"z", "zl"}
	for i := 0; i < len(tokens); i++ {
		tokenData := types.TokenData{Text: tokens[i]}
		tokenDatas = append(tokenDatas, tokenData)
	}

	index1 := types.DocIndexData{Tokens: tokenDatas, Fields: "在路上"}
	index2 := types.DocIndexData{Content: "在路上, in the way",
		Tokens: tokenDatas}

	engine.IndexDoc(10, index1)
	engine.IndexDoc(11, index2)

	engine.Flush()

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[10] = true
	docIds[11] = true
	outputs := engine.Search(types.SearchReq{
		Text:   "zl",
		DocIds: docIds,
	})

	fmt.Println("outputs", outputs.Docs)
	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		utils.Expect(t, "2", len(outDocs))
		// utils.Expect(t, "11", outDocs[0].DocId)
		// utils.Expect(t, "10", outDocs[1].DocId)
	}
	utils.Expect(t, "1", len(outputs.Tokens))
	utils.Expect(t, "2", outputs.NumDocs)

	engine.Close()
	os.RemoveAll("riot.py")
}

func TestForSplitData(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         4,
		UseStorage:    true,
		StorageFolder: "riot.data",
		IDOnly:        true,
		SegmenterDict: "./testdata/test_dict.txt",
	})

	AddDocs(&engine)
	engine.RemoveDoc(5)
	engine.Flush()

	tokenDatas := engine.PinYin("在路上, in the way")
	tokens, num := engine.ForSplitData(tokenDatas, 52)
	utils.Expect(t, "93", len(tokens))
	utils.Expect(t, "104", num)

	index1 := types.DocIndexData{Content: "在路上"}
	engine.IndexDoc(10, index1, true)

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	outputs := engine.Search(types.SearchReq{
		Text:   "中国人口",
		DocIds: docIds})

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		utils.Expect(t, "0", len(outDocs))
	}
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "0", outputs.NumDocs)

	engine.Close()
	os.RemoveAll("riot.data")
}

func TestDocCounters(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:         1,
		UseStorage:    true,
		StorageFolder: "riot.doc",
		IDOnly:        true,
		SegmenterDict: "./testdata/test_dict.txt",
	})

	AddDocs(&engine)
	engine.RemoveDoc(5)
	engine.Flush()

	numAdd := engine.NumTokenIndexAdded()
	utils.Expect(t, "14", numAdd)
	numInx := engine.NumDocsIndexed()
	utils.Expect(t, "5", numInx)
	numRm := engine.NumDocsRemoved()
	utils.Expect(t, "8", numRm)

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

	engine.Close()
	os.RemoveAll("riot.doc")
}
