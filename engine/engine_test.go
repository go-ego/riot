package engine

import (
	"encoding/gob"
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
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "中国有十三亿人口人口",
		Fields:  ScoringFields{1, 2, 3},
	}, false)
	docId++
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "中国人口",
		Fields:  nil,
	}, false)
	docId++
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "有人口",
		Fields:  ScoringFields{2, 3, 1},
	}, false)
	docId++
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "有十三亿人口",
		Fields:  ScoringFields{2, 3, 3},
	}, false)
	docId++
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "中国十三亿人口",
		Fields:  ScoringFields{0, 9, 1},
	}, false)
	engine.FlushIndex()
}

func addDocsWithLabels(engine *Engine) {
	docId := uint64(1)
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "此次百度收购将成中国互联网最大并购",
		Labels:  []string{"百度", "中国"},
	}, false)
	docId++
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "百度宣布拟全资收购91无线业务",
		Labels:  []string{"百度"},
	}, false)
	docId++
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "百度是中国最大的搜索引擎",
		Labels:  []string{"百度"},
	}, false)
	docId++
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "百度在研制无人汽车",
		Labels:  []string{"百度"},
	}, false)
	docId++
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "BAT是中国互联网三巨头",
		Labels:  []string{"百度"},
	}, false)
	engine.FlushIndex()
}

type RankByTokenProximity struct {
}

func (rule RankByTokenProximity) Score(
	doc types.IndexedDocument, fields interface{}) []float32 {
	if doc.TokenProximity < 0 {
		return []float32{}
	}
	return []float32{1.0 / (float32(doc.TokenProximity) + 1)}
}

func TestEngineIndexDocument(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "中国", outputs.Tokens[0])
	utils.Expect(t, "人口", outputs.Tokens[1])
	utils.Expect(t, "3", len(outputs.Docs))

	log.Println("TestEngineIndexDocument:", outputs.Docs)
	utils.Expect(t, "2", outputs.Docs[0].DocId)
	utils.Expect(t, "1000", int(outputs.Docs[0].Scores[0]*1000))
	utils.Expect(t, "[0 6]", outputs.Docs[0].TokenSnippetLocations)

	utils.Expect(t, "5", outputs.Docs[1].DocId)
	utils.Expect(t, "100", int(outputs.Docs[1].Scores[0]*1000))
	utils.Expect(t, "[0 15]", outputs.Docs[1].TokenSnippetLocations)

	utils.Expect(t, "1", outputs.Docs[2].DocId)
	utils.Expect(t, "76", int(outputs.Docs[2].Scores[0]*1000))
	utils.Expect(t, "[0 18]", outputs.Docs[2].TokenSnippetLocations)
}

func TestReverseOrder(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "3", len(outputs.Docs))

	utils.Expect(t, "1", outputs.Docs[0].DocId)
	utils.Expect(t, "5", outputs.Docs[1].DocId)
	utils.Expect(t, "2", outputs.Docs[2].DocId)
}

func TestOffsetAndMaxOutputs(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ReverseOrder:    true,
			OutputOffset:    1,
			MaxOutputs:      3,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Docs))

	utils.Expect(t, "5", outputs.Docs[0].DocId)
	utils.Expect(t, "2", outputs.Docs[1].DocId)
}

type TestScoringCriteria struct {
}

func (criteria TestScoringCriteria) Score(
	doc types.IndexedDocument, fields interface{}) []float32 {
	if reflect.TypeOf(fields) != reflect.TypeOf(ScoringFields{}) {
		return []float32{}
	}
	fs := fields.(ScoringFields)
	return []float32{float32(doc.TokenProximity)*fs.A + fs.B*fs.C}
}

func TestSearchWithCriteria(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ScoringCriteria: TestScoringCriteria{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Docs))

	log.Println(outputs.Docs)
	utils.Expect(t, "1", outputs.Docs[0].DocId)
	utils.Expect(t, "18000", int(outputs.Docs[0].Scores[0]*1000))

	utils.Expect(t, "5", outputs.Docs[1].DocId)
	utils.Expect(t, "9000", int(outputs.Docs[1].Scores[0]*1000))
}

func TestCompactIndex(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ScoringCriteria: TestScoringCriteria{},
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Docs))

	utils.Expect(t, "5", outputs.Docs[0].DocId)
	utils.Expect(t, "9000", int(outputs.Docs[0].Scores[0]*1000))

	utils.Expect(t, "1", outputs.Docs[1].DocId)
	utils.Expect(t, "6000", int(outputs.Docs[1].Scores[0]*1000))
}

type BM25ScoringCriteria struct {
}

func (criteria BM25ScoringCriteria) Score(
	doc types.IndexedDocument, fields interface{}) []float32 {
	if reflect.TypeOf(fields) != reflect.TypeOf(ScoringFields{}) {
		return []float32{}
	}
	return []float32{doc.BM25}
}

func TestFrequenciesIndex(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ScoringCriteria: BM25ScoringCriteria{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.FrequenciesIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Docs))

	utils.Expect(t, "1", outputs.Docs[0].DocId)
	utils.Expect(t, "2500", int(outputs.Docs[0].Scores[0]*1000))

	utils.Expect(t, "5", outputs.Docs[1].DocId)
	utils.Expect(t, "1818", int(outputs.Docs[1].Scores[0]*1000))
}

func TestRemoveDocument(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ScoringCriteria: TestScoringCriteria{},
		},
	})

	AddDocs(&engine)
	engine.RemoveDocument(5, false)
	engine.RemoveDocument(6, false)
	engine.FlushIndex()
	engine.IndexDocument(6, types.DocIndexData{
		Content: "中国人口有十三亿",
		Fields:  ScoringFields{0, 9, 1},
	}, false)
	engine.FlushIndex()

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Docs))

	utils.Expect(t, "6", outputs.Docs[0].DocId)
	utils.Expect(t, "9000", int(outputs.Docs[0].Scores[0]*1000))
	utils.Expect(t, "1", outputs.Docs[1].DocId)
	utils.Expect(t, "6000", int(outputs.Docs[1].Scores[0]*1000))
}

func TestEngineIndexDocumentWithTokens(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})

	docId := uint64(1)
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "",
		Tokens: []types.TokenData{
			{"中国", []int{0}},
			{"人口", []int{18, 24}},
		},
		Fields: ScoringFields{1, 2, 3},
	}, false)
	docId++
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "",
		Tokens: []types.TokenData{
			{"中国", []int{0}},
			{"人口", []int{6}},
		},
		Fields: ScoringFields{1, 2, 3},
	}, false)
	docId++
	engine.IndexDocument(docId, types.DocIndexData{
		Content: "中国十三亿人口",
		Fields:  ScoringFields{0, 9, 1},
	}, false)
	engine.FlushIndex()

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	log.Println("TestEngineIndexDocumentWithTokens", outputs)
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "中国", outputs.Tokens[0])
	utils.Expect(t, "人口", outputs.Tokens[1])
	utils.Expect(t, "3", len(outputs.Docs))

	utils.Expect(t, "2", outputs.Docs[0].DocId)
	utils.Expect(t, "1000", int(outputs.Docs[0].Scores[0]*1000))
	utils.Expect(t, "[0 6]", outputs.Docs[0].TokenSnippetLocations)

	utils.Expect(t, "3", outputs.Docs[1].DocId)
	utils.Expect(t, "100", int(outputs.Docs[1].Scores[0]*1000))
	utils.Expect(t, "[0 15]", outputs.Docs[1].TokenSnippetLocations)

	utils.Expect(t, "1", outputs.Docs[2].DocId)
	utils.Expect(t, "76", int(outputs.Docs[2].Scores[0]*1000))
	utils.Expect(t, "[0 18]", outputs.Docs[2].TokenSnippetLocations)
}

func TestEngineIndexDocumentWithContentAndLabels(t *testing.T) {
	var engine1, engine2 Engine
	engine1.Init(types.EngineInitOptions{
		SegmenterDict: "../data/dict/dictionary.txt",
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})
	engine2.Init(types.EngineInitOptions{
		SegmenterDict: "../data/dict/dictionary.txt",
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.DocIdsIndex,
		},
	})

	addDocsWithLabels(&engine1)
	addDocsWithLabels(&engine2)

	outputs1 := engine1.Search(types.SearchRequest{Text: "百度"})
	outputs2 := engine2.Search(types.SearchRequest{Text: "百度"})
	utils.Expect(t, "1", len(outputs1.Tokens))
	utils.Expect(t, "1", len(outputs2.Tokens))
	utils.Expect(t, "百度", outputs1.Tokens[0])
	utils.Expect(t, "百度", outputs2.Tokens[0])
	utils.Expect(t, "5", len(outputs1.Docs))
	utils.Expect(t, "5", len(outputs2.Docs))
}

func TestEngineIndexDocumentWithPersistentStorage(t *testing.T) {
	gob.Register(ScoringFields{})
	var engine Engine
	engine.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
		UseStorage:    true,
		StorageFolder: "riot.persistent",
		StorageShards: 2,
	})
	AddDocs(&engine)
	engine.RemoveDocument(5, true)
	engine.Close()

	var engine1 Engine
	engine1.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
		UseStorage:    true,
		StorageFolder: "riot.persistent",
		StorageShards: 2,
	})
	engine1.FlushIndex()

	outputs := engine1.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "中国", outputs.Tokens[0])
	utils.Expect(t, "人口", outputs.Tokens[1])
	utils.Expect(t, "2", len(outputs.Docs))

	utils.Expect(t, "2", outputs.Docs[0].DocId)
	utils.Expect(t, "1000", int(outputs.Docs[0].Scores[0]*1000))
	utils.Expect(t, "[0 6]", outputs.Docs[0].TokenSnippetLocations)

	utils.Expect(t, "1", outputs.Docs[1].DocId)
	utils.Expect(t, "76", int(outputs.Docs[1].Scores[0]*1000))
	utils.Expect(t, "[0 18]", outputs.Docs[1].TokenSnippetLocations)

	engine1.Close()
	os.RemoveAll("riot.persistent")
}

func TestCountDocsOnly(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      1,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})

	AddDocs(&engine)
	engine.RemoveDocument(5, false)
	engine.FlushIndex()

	outputs := engine.Search(types.SearchRequest{Text: "中国人口", CountDocsOnly: true})
	utils.Expect(t, "0", len(outputs.Docs))
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "2", outputs.NumDocs)
}

func TestSearchWithin(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		Using:         1,
		SegmenterDict: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})

	AddDocs(&engine)

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	outputs := engine.Search(types.SearchRequest{
		Text:   "中国人口",
		DocIds: docIds,
	})
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "中国", outputs.Tokens[0])
	utils.Expect(t, "人口", outputs.Tokens[1])
	utils.Expect(t, "2", len(outputs.Docs))

	utils.Expect(t, "1", outputs.Docs[0].DocId)
	utils.Expect(t, "76", int(outputs.Docs[0].Scores[0]*1000))
	utils.Expect(t, "[0 18]", outputs.Docs[0].TokenSnippetLocations)

	utils.Expect(t, "5", outputs.Docs[1].DocId)
	utils.Expect(t, "100", int(outputs.Docs[1].Scores[0]*1000))
	utils.Expect(t, "[0 15]", outputs.Docs[1].TokenSnippetLocations)
}
