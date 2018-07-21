package riot

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/go-ego/gse"
	"github.com/go-ego/riot/types"
	"github.com/vcaesar/tt"
)

type ScoringFields struct {
	A, B, C float32
}

func TestGetVer(t *testing.T) {
	fmt.Println("go version: ", runtime.Version())
	ver := GetVersion()
	tt.Expect(t, Version, ver)
	tt.Equal(t, Version, ver)
}

func AddDocs(engine *Engine) {
	docId := uint64(1)
	engine.Index(docId, types.DocData{
		Content: "中国有十三亿人口人口",
		Fields:  ScoringFields{1, 2, 3},
	})
	docId++
	engine.IndexDoc(docId, types.DocIndexData{
		Content: "中国人口",
		Fields:  nil,
	})
	docId++
	engine.Index(docId, types.DocData{
		Content: "有人口",
		Fields:  ScoringFields{2, 3, 1},
	})
	docId++
	engine.Index(docId, types.DocData{
		Content: "有十三亿人口",
		Fields:  ScoringFields{2, 3, 3},
	})
	docId++
	engine.Index(docId, types.DocData{
		Content: "中国十三亿人口",
		Fields:  ScoringFields{0, 9, 1},
	})

	engine.Flush()
}

func addDocsWithLabels(engine *Engine) {
	docId := uint64(1)
	engine.Index(docId, types.DocData{
		Content: "此次百度收购将成中国互联网最大并购",
		Labels:  []string{"百度", "中国"},
	})
	log.Println("engine.Segment(): ", engine.Segment("此次百度收购将成中国互联网最大并购"))

	docId++
	engine.Index(docId, types.DocData{
		Content: "百度宣布拟全资收购91无线业务",
		Labels:  []string{"百度"},
	})
	docId++
	engine.Index(docId, types.DocData{
		Content: "百度是中国最大的搜索引擎",
		Labels:  []string{"百度"},
	})
	docId++
	engine.Index(docId, types.DocData{
		Content: "百度在研制无人汽车",
		Labels:  []string{"百度"},
	})
	docId++
	engine.Index(docId, types.DocData{
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

func TestTry(t *testing.T) {
	var arr []int

	Try(func() {
		fmt.Println(arr[2])
	}, func(err interface{}) {
		log.Println("err", err)
		tt.Expect(t, "runtime error: index out of range", err)
	})
}

func TestEngineIndexDoc(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
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
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "中国", outputs.Tokens[0])
	tt.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "3", len(outDocs))

	log.Println("TestEngineIndexDoc:", outDocs)
	tt.Expect(t, "2", outDocs[0].DocId)
	tt.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[0 6]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "5", outDocs[1].DocId)
	tt.Expect(t, "100", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[0 15]", outDocs[1].TokenSnippetLocs)

	tt.Expect(t, "1", outDocs[2].DocId)
	tt.Expect(t, "76", int(outDocs[2].Scores[0]*1000))
	tt.Expect(t, "[0 18]", outDocs[2].TokenSnippetLocs)

	engine.Close()
}

func TestReverseOrder(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
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
	tt.Expect(t, "3", len(outDocs))

	tt.Expect(t, "1", outDocs[0].DocId)
	tt.Expect(t, "5", outDocs[1].DocId)
	tt.Expect(t, "2", outDocs[2].DocId)

	engine.Close()
}

func TestOffsetAndMaxOutputs(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
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
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "5", outDocs[0].DocId)
	tt.Expect(t, "2", outDocs[1].DocId)

	engine.Close()
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
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
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
	tt.Expect(t, "2", len(outDocs))

	log.Println(outDocs)
	tt.Expect(t, "1", outDocs[0].DocId)
	tt.Expect(t, "18000", int(outDocs[0].Scores[0]*1000))

	tt.Expect(t, "5", outDocs[1].DocId)
	tt.Expect(t, "9000", int(outDocs[1].Scores[0]*1000))

	engine.Close()
}

func TestCompactIndex(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ScoringCriteria: TestScoringCriteria{},
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchReq{Text: "中国人口"})

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "5", outDocs[0].DocId)
	tt.Expect(t, "9000", int(outDocs[0].Scores[0]*1000))

	tt.Expect(t, "1", outDocs[1].DocId)
	tt.Expect(t, "6000", int(outDocs[1].Scores[0]*1000))

	engine.Close()
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
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
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
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "1", outDocs[0].DocId)
	tt.Expect(t, "2500", int(outDocs[0].Scores[0]*1000))

	tt.Expect(t, "5", outDocs[1].DocId)
	tt.Expect(t, "1818", int(outDocs[1].Scores[0]*1000))

	engine.Close()
}

func TestRemoveDoc(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			ScoringCriteria: TestScoringCriteria{},
		},
	})

	AddDocs(&engine)

	engine.RemoveDoc(5)
	engine.RemoveDoc(6)
	engine.Flush()

	engine.Index(6, types.DocData{
		Content: "中国人口有十三亿",
		Fields:  ScoringFields{0, 9, 1},
	})
	engine.Flush()

	outputs := engine.Search(types.SearchReq{Text: "中国人口"})

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "6", outDocs[0].DocId)
	tt.Expect(t, "9000", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "1", outDocs[1].DocId)
	tt.Expect(t, "6000", int(outDocs[1].Scores[0]*1000))

	engine.Close()
}

func TestEngineIndexWithTokens(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
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
	engine.Index(docId, types.DocData{
		Content: "",
		Tokens: []types.TokenData{
			{"中国", []int{0}},
			{"人口", []int{18, 24}},
		},
		Fields: ScoringFields{1, 2, 3},
	})

	docId++
	engine.Index(docId, types.DocData{
		Content: "",
		Tokens: []types.TokenData{
			{"中国", []int{0}},
			{"人口", []int{6}},
		},
		Fields: ScoringFields{1, 2, 3},
	})

	docId++
	engine.Index(docId, types.DocData{
		Content: "中国十三亿人口",
		Fields:  ScoringFields{0, 9, 1},
	})
	engine.FlushIndex()

	outputs := engine.Search(types.SearchReq{Text: "中国人口"})
	log.Println("TestEngineIndexWithTokens: ", outputs)
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "中国", outputs.Tokens[0])
	tt.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "3", len(outDocs))

	tt.Expect(t, "2", outDocs[0].DocId)
	tt.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[0 6]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "3", outDocs[1].DocId)
	tt.Expect(t, "100", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[0 15]", outDocs[1].TokenSnippetLocs)

	tt.Expect(t, "1", outDocs[2].DocId)
	tt.Expect(t, "76", int(outDocs[2].Scores[0]*1000))
	tt.Expect(t, "[0 18]", outDocs[2].TokenSnippetLocs)

	engine.Close()
}

func TestEngineIndexWithContentAndLabels(t *testing.T) {
	var engine1, engine2 Engine
	engine1.Init(types.EngineOpts{
		GseDict: "./data/dict/dictionary.txt",
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	engine2.Init(types.EngineOpts{
		GseDict: "./data/dict/dictionary.txt",
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.DocIdsIndex,
		},
	})

	addDocsWithLabels(&engine1)
	addDocsWithLabels(&engine2)

	outputs1 := engine1.Search(types.SearchReq{Text: "百度"})
	outputs2 := engine2.Search(types.SearchReq{Text: "百度"})
	tt.Expect(t, "1", len(outputs1.Tokens))
	tt.Expect(t, "1", len(outputs2.Tokens))
	tt.Expect(t, "百度", outputs1.Tokens[0])
	tt.Expect(t, "百度", outputs2.Tokens[0])

	outDocs := outputs1.Docs.(types.ScoredDocs)
	tt.Expect(t, "5", len(outDocs))
	tt.Expect(t, "5", len(outputs2.Docs.(types.ScoredDocs)))

	engine1.Close()
	engine2.Close()
}

func TestIndexWithLabelsStopTokenFile(t *testing.T) {
	var engine1 Engine

	engine1.Init(types.EngineOpts{
		GseDict:       "./data/dict/dictionary.txt",
		StopTokenFile: "./testdata/test_stop_dict.txt",
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	addDocsWithLabels(&engine1)

	outputs1 := engine1.Search(types.SearchReq{Text: "百度"})
	tt.Expect(t, "0", len(outputs1.Tokens))
	// tt.Expect(t, "百度", outputs1.Tokens[0])

	outDocs := outputs1.Docs.(types.ScoredDocs)
	tt.Expect(t, "0", len(outDocs))
}

func TestEngineIndexWithStore(t *testing.T) {
	gob.Register(ScoringFields{})

	var engine Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
		UseStore:    true,
		StoreFolder: "riot.persistent",
		StoreShards: 2,
	})

	AddDocs(&engine)

	engine.RemoveDoc(5, true)
	engine.Flush()

	engine.Close()

	var engine1 Engine
	engine1.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
		DefaultRankOpts: &types.RankOpts{
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
		UseStore:    true,
		StoreFolder: "riot.persistent",
		StoreShards: 2,
	})
	engine1.Flush()

	outputs := engine1.Search(types.SearchReq{Text: "中国人口"})
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "中国", outputs.Tokens[0])
	tt.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "2", outDocs[0].DocId)
	tt.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[0 6]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "1", outDocs[1].DocId)
	tt.Expect(t, "76", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[0 18]", outDocs[1].TokenSnippetLocs)

	engine1.Close()
	os.RemoveAll("riot.persistent")
}

func TestCountDocsOnly(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
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
	// tt.Expect(t, "0", len(outputs.Docs))
	if outputs.Docs == nil {
		tt.Expect(t, "0", 0)
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "2", outputs.NumDocs)

	engine.Close()
}

func TestDocOrderless(t *testing.T) {
	var engine, engine1 Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
	})

	AddDocs(&engine)

	engine.RemoveDoc(5)
	engine.Flush()

	outputs := engine.Search(types.SearchReq{Text: "中国人口",
		Orderless: true})
	// tt.Expect(t, "0", len(outputs.Docs))
	if outputs.Docs == nil {
		tt.Expect(t, "0", 0)
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "2", outputs.NumDocs)

	engine1.Init(types.EngineOpts{
		Using:   1,
		IDOnly:  true,
		GseDict: "./testdata/test_dict.txt",
	})

	AddDocs(&engine1)

	engine1.RemoveDoc(5)
	engine1.Flush()

	outputs1 := engine1.Search(types.SearchReq{Text: "中国人口",
		Orderless: true})
	if outputs1.Docs == nil {
		tt.Expect(t, "0", 0)
	}

	tt.Expect(t, "2", len(outputs1.Tokens))
	tt.Expect(t, "2", outputs1.NumDocs)

	engine.Close()
}

func TestDocOnlyID(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		IDOnly:  true,
		GseDict: "./testdata/test_dict.txt",
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
		tt.Expect(t, "1", len(outDocs))
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "1", outputs.NumDocs)

	outputs1 := engine.Search(types.SearchReq{
		Text:    "中国人口",
		Timeout: 10,
		DocIds:  docIds})

	if outputs1.Docs != nil {
		outDocs1 := outputs.Docs.(types.ScoredIDs)
		tt.Expect(t, "1", len(outDocs1))
	}
	tt.Expect(t, "2", len(outputs1.Tokens))
	tt.Expect(t, "1", outputs1.NumDocs)

	engine.Close()
}

func TestSearchWithin(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
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
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "中国", outputs.Tokens[0])
	tt.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "1", outDocs[0].DocId)
	tt.Expect(t, "76", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[0 18]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "5", outDocs[1].DocId)
	tt.Expect(t, "100", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[0 15]", outDocs[1].TokenSnippetLocs)

	engine.Close()
}

func TestSearchJp(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict_jp.txt",
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

	engine.Index(6, types.DocData{
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

	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "こんにちは", outputs.Tokens[0])
	tt.Expect(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	tt.Expect(t, "1", len(outDocs))

	tt.Expect(t, "6", outDocs[0].DocId)
	tt.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[0 15]", outDocs[0].TokenSnippetLocs)

	engine.Close()
}

func TestSearchGse(t *testing.T) {
	log.Println("Test search gse ...")
	var engine Engine
	engine.Init(types.EngineOpts{
		// Using:   1,
		GseDict: "./testdata/test_dict_jp.txt",
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

	engine.Index(6, types.DocData{
		Content: "こんにちは世界, こんにちは",
		Fields:  ScoringFields{1, 2, 3},
	})

	tokenData := types.TokenData{Text: "こんにちは"}
	tokenDatas := []types.TokenData{tokenData}
	engine.Index(7, types.DocData{
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

	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "こんにちは", outputs.Tokens[0])
	tt.Expect(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "7", outDocs[0].DocId)
	tt.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "6", outDocs[1].DocId)
	tt.Expect(t, "1000", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[0 15]", outDocs[1].TokenSnippetLocs)

	engine.Close()
}

func TestSearchNotUseGse(t *testing.T) {
	var engine, engine1 Engine
	engine.Init(types.EngineOpts{
		Using:     4,
		NotUseGse: true,
	})

	engine1.Init(types.EngineOpts{
		IDOnly:    true,
		NotUseGse: true,
	})

	AddDocs(&engine)
	AddDocs(&engine1)

	engine.Index(6, types.DocData{
		Content: "Google Is Experimenting With Virtual Reality Advertising",
		Fields:  ScoringFields{1, 2, 3},
		Tokens:  []types.TokenData{{Text: "test"}},
	})

	engine1.Index(6, types.DocData{
		Content: "Google Is Experimenting With Virtual Reality Advertising",
		Fields:  ScoringFields{1, 2, 3},
		Tokens:  []types.TokenData{{Text: "test"}},
	})

	engine.Flush()
	engine1.Flush()

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	docIds[6] = true

	outputs := engine.Search(types.SearchReq{
		Text:   "google is",
		DocIds: docIds,
	})

	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "google", outputs.Tokens[0])
	tt.Expect(t, "is", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	tt.Expect(t, "1", len(outDocs))

	tt.Expect(t, "6", outDocs[0].DocId)
	tt.Expect(t, "3170", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[]", outDocs[0].TokenSnippetLocs)

	outputs1 := engine1.Search(types.SearchReq{
		Text:   "google",
		DocIds: docIds})
	tt.Expect(t, "1", len(outputs1.Tokens))
	tt.Expect(t, "0", outputs1.NumDocs)

	engine.Close()
	engine1.Close()
}

func TestSearchWithGse(t *testing.T) {
	gseSegmenter := gse.Segmenter{}
	gseSegmenter.LoadDict("zh") // ./data/dict/dictionary.txt

	var engine1, engine2, searcher2 Engine
	searcher2.Init(types.EngineOpts{
		Using: 1,
	})
	defer searcher2.Close()

	engine1.WithGse(gseSegmenter).Init(types.EngineOpts{
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.LocsIndex,
		},
	})

	engine2.WithGse(gseSegmenter).Init(types.EngineOpts{
		// GseDict: "./data/dict/dictionary.txt",
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.DocIdsIndex,
		},
	})

	addDocsWithLabels(&engine1)
	addDocsWithLabels(&engine2)

	outputs1 := engine1.Search(types.SearchReq{Text: "百度"})
	outputs2 := engine2.Search(types.SearchReq{Text: "百度"})
	tt.Expect(t, "1", len(outputs1.Tokens))
	tt.Expect(t, "1", len(outputs2.Tokens))
	tt.Expect(t, "百度", outputs1.Tokens[0])
	tt.Expect(t, "百度", outputs2.Tokens[0])

	outDocs := outputs1.Docs.(types.ScoredDocs)
	tt.Expect(t, "5", len(outDocs))
	tt.Expect(t, "5", len(outputs2.Docs.(types.ScoredDocs)))

	engine1.Close()
	engine2.Close()
}

func TestRiotGse(t *testing.T) {
	var engine, engine1 Engine
	engine.Init(types.EngineOpts{
		Using: 1,
	})

	AddDocs(&engine)

	engine1.Init(types.EngineOpts{
		Using:   1,
		GseMode: true,
	})

	AddDocs(&engine1)
	tt.Equal(t, "[此次 百度 收购 将 成 中国 互联网 最大 并购]",
		engine.Segment("此次百度收购将成中国互联网最大并购"))
	tt.Equal(t, "[此次 百度 收购 将 成 中国 互联网 最大 并购]",
		engine1.Segment("此次百度收购将成中国互联网最大并购"))

	engine.Close()
	engine1.Close()
}

func TestSearchLogic(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		GseDict: "./testdata/test_dict_jp.txt",
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

	engine.Index(6, types.DocData{
		Content: "こんにちは世界, こんにちは",
		Fields:  ScoringFields{1, 2, 3},
	})

	tokenData := types.TokenData{Text: "こんにちは"}
	tokenDatas := []types.TokenData{tokenData}
	engine.Index(7, types.DocData{
		Content: "你好世界, hello world!",
		Tokens:  tokenDatas,
		Fields:  ScoringFields{1, 2, 3},
	})

	engine.Index(8, types.DocData{
		Content: "你好世界, hello world!",
		Fields:  ScoringFields{1, 2, 3},
	})

	engine.Index(9, types.DocData{
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

	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "こんにちは", outputs.Tokens[0])
	tt.Expect(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "9", outDocs[0].DocId)
	tt.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "8", outDocs[1].DocId)
	tt.Expect(t, "1000", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[]", outDocs[1].TokenSnippetLocs)

	engine.Close()
}
