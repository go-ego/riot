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

var (
	text1   = "Hello world, 你好世界!"
	textJP  = "こんにちは世界"
	textJP1 = "こんにちは世界, こんにちは"
	reqText = "World人口"
	reqG    = types.SearchReq{Text: "Google"}
	req1    = types.SearchReq{Text: reqText}

	score1   = ScoringFields{1, 2, 3}
	score091 = ScoringFields{0, 9, 1}
	inxOpts  = &types.IndexerOpts{
		IndexType: types.LocsIndex,
	}
)

func TestGetVer(t *testing.T) {
	fmt.Println("go version: ", runtime.Version())
	ver := GetVersion()
	tt.Expect(t, Version, ver)
	tt.Equal(t, Version, ver)
}

func AddDocs(engine *Engine) {
	// docId := uint64(1)
	engine.Index(1, types.DocData{
		Content: "The world, 有七十亿人口人口",
		Fields:  score1,
	})

	// docId++
	engine.IndexDoc(2, types.DocIndexData{
		Content: "The world, 人口",
		Fields:  nil,
	})

	engine.Index(3, types.DocData{
		Content: "The world",
		Fields:  nil,
	})

	engine.Index(4, types.DocData{
		Content: "有人口",
		Fields:  ScoringFields{2, 3, 1},
	})

	engine.Index(5, types.DocData{
		Content: "The world, 七十亿人口",
		Fields:  score091,
	})

	engine.Index(6, types.DocData{
		Content: "有七十亿人口",
		Fields:  ScoringFields{2, 3, 3},
	})

	engine.Flush()
}

func addDocsWithLabels(engine *Engine) {
	// docId := uint64(1)
	engine.Index(1, types.DocData{
		Content: "《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄",
		Labels:  []string{"复仇者", "战争"},
	})
	log.Println("engine.Segment(): ",
		engine.Segment("《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄"))

	// docId++
	engine.Index(2, types.DocData{
		Content: "在IMAX影院放映时",
		Labels:  []string{"影院"},
	})

	engine.Index(3, types.DocData{
		Content: " Google 是世界最大搜索引擎, baidu 是最大中文的搜索引擎",
		Labels:  []string{"Google"},
	})

	engine.Index(4, types.DocData{
		Content: "Google 在研制无人汽车",
		Labels:  []string{"Google"},
	})

	engine.Index(5, types.DocData{
		Content: " GAMAF 世界五大互联网巨头, BAT 是中国互联网三巨头",
		Labels:  []string{"互联网"},
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

func rankOptsOrder(order bool) types.RankOpts {
	return types.RankOpts{
		ReverseOrder:    order,
		OutputOffset:    0,
		MaxOutputs:      10,
		ScoringCriteria: &RankByTokenProximity{},
	}
}

func rankOptsMax(output, max int) types.RankOpts {
	return types.RankOpts{
		ReverseOrder:    true,
		OutputOffset:    output,
		MaxOutputs:      max,
		ScoringCriteria: &RankByTokenProximity{},
	}
}

var (
	rankOptsMax1 = rankOptsMax(0, 1)

	rankOptsMax10      = rankOptsOrder(false)
	rankOptsMax10Order = rankOptsOrder(true)

	rankOptsMax3 = rankOptsMax(1, 3)
)

func rankEngineOpts(rankOpts types.RankOpts) types.EngineOpts {
	return types.EngineOpts{
		Using:       1,
		GseDict:     "./testdata/test_dict.txt",
		DefRankOpts: &rankOpts,
		IndexerOpts: inxOpts,
	}
}

var (
	testIndexOpts = rankEngineOpts(rankOptsMax10)

	orderOpts = rankEngineOpts(rankOptsMax10Order)
)

func TestEngineIndexDoc(t *testing.T) {
	var engine Engine
	engine.Init(testIndexOpts)

	AddDocs(&engine)

	outputs := engine.Search(req1)
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "world", outputs.Tokens[0])
	tt.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "3", len(outDocs))

	log.Println("TestEngineIndexDoc:", outDocs)
	tt.Expect(t, "2", outDocs[0].DocId)
	tt.Expect(t, "333", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[4 11]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "5", outDocs[1].DocId)
	tt.Expect(t, "83", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[4 20]", outDocs[1].TokenSnippetLocs)

	tt.Expect(t, "1", outDocs[2].DocId)
	tt.Expect(t, "66", int(outDocs[2].Scores[0]*1000))
	tt.Expect(t, "[4 23]", outDocs[2].TokenSnippetLocs)

	engine.Close()
}
func TestReverseOrder(t *testing.T) {
	var engine Engine
	engine.Init(orderOpts)

	AddDocs(&engine)

	outputs := engine.Search(req1)

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
		Using:       1,
		GseDict:     "./testdata/test_dict.txt",
		DefRankOpts: &rankOptsMax3,
		IndexerOpts: inxOpts,
	})

	AddDocs(&engine)

	outputs := engine.Search(req1)

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

var (
	engOpts = types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
		DefRankOpts: &types.RankOpts{
			ScoringCriteria: TestScoringCriteria{},
		},
		IndexerOpts: inxOpts,
	}
)

func TestSearchWithCriteria(t *testing.T) {
	var engine Engine
	engine.Init(engOpts)

	AddDocs(&engine)

	outputs := engine.Search(req1)

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))

	log.Println(outDocs)
	tt.Expect(t, "1", outDocs[0].DocId)
	tt.Expect(t, "20000", int(outDocs[0].Scores[0]*1000))

	tt.Expect(t, "5", outDocs[1].DocId)
	tt.Expect(t, "9000", int(outDocs[1].Scores[0]*1000))

	engine.Close()
}

func TestCompactIndex(t *testing.T) {
	var engine Engine
	engine.Init(useOpts)

	AddDocs(&engine)

	outputs := engine.Search(req1)

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
		DefRankOpts: &types.RankOpts{
			ScoringCriteria: BM25ScoringCriteria{},
		},
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.FrequenciesIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(req1)

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "1", outDocs[0].DocId)
	tt.Expect(t, "2374", int(outDocs[0].Scores[0]*1000))

	tt.Expect(t, "5", outDocs[1].DocId)
	tt.Expect(t, "2133", int(outDocs[1].Scores[0]*1000))

	engine.Close()
}

var (
	useOpts = types.EngineOpts{
		Using:   1,
		GseDict: "./testdata/test_dict.txt",
		DefRankOpts: &types.RankOpts{
			ScoringCriteria: TestScoringCriteria{},
		},
	}
)

func TestRemoveDoc(t *testing.T) {
	var engine Engine
	engine.Init(useOpts)

	AddDocs(&engine)

	engine.RemoveDoc(5)
	engine.RemoveDoc(6)
	engine.Flush()

	engine.Index(6, types.DocData{
		Content: "World, 人口有七十亿",
		Fields:  score091,
	})
	engine.Flush()

	outputs := engine.Search(req1)

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
	engine.Init(testIndexOpts)

	data1 := types.TokenData{
		Text:      "world",
		Locations: []int{0},
	}
	// docId := uint64(1)
	engine.Index(1, types.DocData{
		Content: "",
		Tokens: []types.TokenData{
			data1,
			{"人口", []int{18, 24}},
		},
		Fields: score1,
	})

	// docId++
	engine.Index(2, types.DocData{
		Content: "",
		Tokens: []types.TokenData{
			data1,
			{"人口", []int{6}},
		},
		Fields: score1,
	})

	engine.Index(3, types.DocData{
		Content: "The world, 七十亿人口",
		Fields:  score091,
	})
	engine.FlushIndex()

	outputs := engine.Search(req1)
	log.Println("TestEngineIndexWithTokens: ", outputs)
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "world", outputs.Tokens[0])
	tt.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "3", len(outDocs))

	tt.Expect(t, "2", outDocs[0].DocId)
	tt.Expect(t, "500", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[0 6]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "3", outDocs[1].DocId)
	tt.Expect(t, "83", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[4 20]", outDocs[1].TokenSnippetLocs)

	tt.Expect(t, "1", outDocs[2].DocId)
	tt.Expect(t, "71", int(outDocs[2].Scores[0]*1000))
	tt.Expect(t, "[0 18]", outDocs[2].TokenSnippetLocs)

	engine.Close()
}

func testLabelsOpts(indexType int) types.EngineOpts {
	return types.EngineOpts{
		GseDict: "./data/dict/dictionary.txt",
		IndexerOpts: &types.IndexerOpts{
			IndexType: indexType,
		},
	}
}

func TestEngineIndexWithContentAndLabels(t *testing.T) {
	var engine1, engine2 Engine
	engine1.Init(testLabelsOpts(types.LocsIndex))
	engine2.Init(testLabelsOpts(types.DocIdsIndex))

	addDocsWithLabels(&engine1)
	addDocsWithLabels(&engine2)

	outputs1 := engine1.Search(reqG)
	outputs2 := engine2.Search(reqG)
	tt.Expect(t, "1", len(outputs1.Tokens))
	tt.Expect(t, "1", len(outputs2.Tokens))
	tt.Expect(t, "google", outputs1.Tokens[0])
	tt.Expect(t, "google", outputs2.Tokens[0])

	outDocs := outputs1.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))
	tt.Expect(t, "2", len(outputs2.Docs.(types.ScoredDocs)))

	engine1.Close()
	engine2.Close()
}

func TestIndexWithLabelsStopTokenFile(t *testing.T) {
	var engine1 Engine

	engine1.Init(types.EngineOpts{
		GseDict:       "./data/dict/dictionary.txt",
		StopTokenFile: "./testdata/test_stop_dict.txt",
		IndexerOpts:   inxOpts,
	})

	addDocsWithLabels(&engine1)

	outputs1 := engine1.Search(reqG)
	outputsDoc := engine1.SearchDoc(reqG)
	tt.Expect(t, "1", len(outputs1.Tokens))
	// tt.Expect(t, "Google", outputs1.Tokens[0])

	outDocs := outputs1.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))
	tt.Expect(t, "2", len(outputsDoc.Docs))
}

func TestEngineIndexWithStore(t *testing.T) {
	gob.Register(ScoringFields{})

	var opts = types.EngineOpts{
		Using:       1,
		GseDict:     "./testdata/test_dict.txt",
		DefRankOpts: &rankOptsMax10,
		IndexerOpts: inxOpts,
		UseStore:    true,
		StoreFolder: "riot.persistent",
		StoreShards: 2,
	}

	var engine Engine
	engine.Init(opts)

	AddDocs(&engine)

	engine.RemoveDoc(5, true)
	engine.Flush()

	engine.Close()

	var engine1 Engine
	engine1.Init(opts)
	engine1.Flush()

	outputs := engine1.Search(req1)
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "world", outputs.Tokens[0])
	tt.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "2", outDocs[0].DocId)
	tt.Expect(t, "333", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[4 11]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "1", outDocs[1].DocId)
	tt.Expect(t, "66", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[4 23]", outDocs[1].TokenSnippetLocs)

	engine1.Close()
	os.RemoveAll("riot.persistent")
}

func TestCountDocsOnly(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:       1,
		GseDict:     "./testdata/test_dict.txt",
		DefRankOpts: &rankOptsMax1,
		IndexerOpts: inxOpts,
	})

	AddDocs(&engine)

	engine.RemoveDoc(5)
	engine.Flush()

	outputs := engine.Search(
		types.SearchReq{Text: reqText, CountDocsOnly: true},
	)
	// tt.Expect(t, "0", len(outputs.Docs))
	if outputs.Docs == nil {
		tt.Expect(t, "0", 0)
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "2", outputs.NumDocs)

	engine.Close()
}

func orderlessOpts(idOnly bool) types.EngineOpts {
	return types.EngineOpts{
		Using:   1,
		IDOnly:  idOnly,
		GseDict: "./testdata/test_dict.txt",
	}
}

func TestDocOrderless(t *testing.T) {
	var engine, engine1 Engine
	engine.Init(orderlessOpts(false))

	AddDocs(&engine)

	engine.RemoveDoc(5)
	engine.Flush()

	orderReq := types.SearchReq{Text: reqText, Orderless: true}
	outputs := engine.Search(orderReq)
	// tt.Expect(t, "0", len(outputs.Docs))
	if outputs.Docs == nil {
		tt.Expect(t, "0", 0)
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "2", outputs.NumDocs)

	engine1.Init(orderlessOpts(true))

	AddDocs(&engine1)

	engine1.RemoveDoc(5)
	engine1.Flush()

	outputs1 := engine1.Search(orderReq)
	if outputs1.Docs == nil {
		tt.Expect(t, "0", 0)
	}

	tt.Expect(t, "2", len(outputs1.Tokens))
	tt.Expect(t, "2", outputs1.NumDocs)

	engine.Close()
}

var (
	testIDInlyOpts = types.EngineOpts{
		// Using:       1,
		IDOnly:      true,
		GseDict:     "./testdata/test_dict.txt",
		DefRankOpts: &rankOptsMax1,
		IndexerOpts: inxOpts,
	}
)

func TestDocOnlyID(t *testing.T) {
	var engine Engine
	engine.Init(testIDInlyOpts)
	AddDocs(&engine)

	engine.RemoveDoc(5)
	engine.Flush()

	req := types.SearchReq{
		Text:   reqText,
		DocIds: makeDocIds(),
	}
	outputs := engine.Search(req)
	outputsID := engine.SearchID(req)
	tt.Expect(t, "1", len(outputsID.Docs))

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		tt.Expect(t, "1", len(outDocs))
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "2", outputs.NumDocs)

	outputs1 := engine.Search(types.SearchReq{
		Text:    reqText,
		Timeout: 10,
		DocIds:  makeDocIds(),
	})

	if outputs1.Docs != nil {
		outDocs1 := outputs.Docs.(types.ScoredIDs)
		tt.Expect(t, "1", len(outDocs1))
	}
	tt.Expect(t, "2", len(outputs1.Tokens))
	tt.Expect(t, "2", outputs1.NumDocs)

	engine.Close()
}

func TestSearchWithin(t *testing.T) {
	var engine Engine
	engine.Init(orderOpts)

	AddDocs(&engine)

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true

	outputs := engine.Search(types.SearchReq{
		Text:   reqText,
		DocIds: docIds,
	})
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "world", outputs.Tokens[0])
	tt.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "1", outDocs[0].DocId)
	tt.Expect(t, "66", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[4 23]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "5", outDocs[1].DocId)
	tt.Expect(t, "83", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[4 20]", outDocs[1].TokenSnippetLocs)

	engine.Close()
}

func testJPOpts(use int) types.EngineOpts {
	return types.EngineOpts{
		// Using:           1,
		Using:       use,
		GseDict:     "./testdata/test_dict_jp.txt",
		DefRankOpts: &rankOptsMax10Order,
		IndexerOpts: inxOpts,
	}
}

func TestSearchJp(t *testing.T) {
	var engine Engine
	engine.Init(testJPOpts(1))

	AddDocs(&engine)

	engine.Index(7, types.DocData{
		Content: textJP1,
		Fields:  score1,
	})
	engine.Flush()

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	docIds[7] = true

	outputs := engine.Search(types.SearchReq{
		Text:   textJP,
		DocIds: docIds,
	})

	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "こんにちは", outputs.Tokens[0])
	tt.Expect(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	tt.Expect(t, "1", len(outDocs))

	tt.Expect(t, "7", outDocs[0].DocId)
	tt.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[0 15]", outDocs[0].TokenSnippetLocs)

	engine.Close()
}

func makeGseDocIds() map[uint64]bool {
	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	docIds[6] = true
	docIds[7] = true
	docIds[8] = true

	return docIds
}

func tokenData() []types.TokenData {
	tokenData := types.TokenData{
		Text:      "こんにちは",
		Locations: []int{10, 20},
	}

	return []types.TokenData{tokenData}
}

func TestSearchGse(t *testing.T) {
	log.Println("Test search gse ...")
	var engine Engine
	engine.Init(testJPOpts(0))

	AddDocs(&engine)

	engine.Index(7, types.DocData{
		Content: textJP1,
		Fields:  score1,
	})

	tokenDatas := tokenData()
	engine.Index(8, types.DocData{
		Content: text1,
		Tokens:  tokenDatas,
		Fields:  ScoringFields{4, 5, 6},
	})
	engine.Flush()

	docIds := makeGseDocIds()
	outputs := engine.Search(types.SearchReq{
		Text:   textJP,
		DocIds: docIds,
	})

	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "こんにちは", outputs.Tokens[0])
	tt.Expect(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "8", outDocs[0].DocId)
	tt.Expect(t, "142", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[10 19]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "7", outDocs[1].DocId)
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

	data := types.DocData{
		Content: "Google Is Experimenting With Virtual Reality Advertising",
		Fields:  score1,
		Tokens:  []types.TokenData{{Text: "test"}},
	}

	engine.Index(7, data)
	engine.Index(8, data)

	engine1.Index(7, data, true)
	engine1.Index(8, data, true)

	engine.Flush()
	engine1.Flush()

	docIds := makeGseDocIds()

	outputs := engine.Search(types.SearchReq{
		Text:   "google is",
		DocIds: docIds,
	})

	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "google", outputs.Tokens[0])
	tt.Expect(t, "is", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "8", outDocs[0].DocId)
	tt.Expect(t, "3736", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[]", outDocs[0].TokenSnippetLocs)

	outputs1 := engine1.Search(types.SearchReq{
		Text:   "google",
		DocIds: docIds})
	tt.Expect(t, "1", len(outputs1.Tokens))
	tt.Expect(t, "2", outputs1.NumDocs)

	engine.Close()
	engine1.Close()
}

func TestSearchWithGse(t *testing.T) {
	seg := gse.Segmenter{}
	seg.LoadDict("zh") // ./data/dict/dictionary.txt

	var engine1, engine2, searcher2 Engine
	searcher2.Init(types.EngineOpts{
		Using: 1,
	})
	defer searcher2.Close()

	engine1.WithGse(seg).Init(types.EngineOpts{
		IndexerOpts: inxOpts,
	})

	engine2.WithGse(seg).Init(types.EngineOpts{
		// GseDict: "./data/dict/dictionary.txt",
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.DocIdsIndex,
		},
	})

	addDocsWithLabels(&engine1)
	addDocsWithLabels(&engine2)

	outputs1 := engine1.Search(reqG)
	outputs2 := engine2.Search(reqG)
	tt.Expect(t, "1", len(outputs1.Tokens))
	tt.Expect(t, "1", len(outputs2.Tokens))
	tt.Expect(t, "google", outputs1.Tokens[0])
	tt.Expect(t, "google", outputs2.Tokens[0])

	outDocs := outputs1.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))
	tt.Expect(t, "2", len(outputs2.Docs.(types.ScoredDocs)))

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
	tt.Equal(t, "[《 复仇者 联盟 3 ： 无限 战争 》 是 全片 使用 imax 摄影机 拍摄]",
		engine.Segment("《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄"))
	tt.Equal(t, "[此次 google 收购 将 成 世界 互联网 最大 并购]",
		engine1.Segment("此次Google收购将成世界互联网最大并购"))

	engine.Close()
	engine1.Close()
}

func TestSearchLogic(t *testing.T) {
	var engine Engine
	engine.Init(testJPOpts(0))

	AddDocs(&engine)

	engine.Index(7, types.DocData{
		Content: textJP1,
		Fields:  score1,
	})

	tokenDatas := tokenData()
	engine.Index(8, types.DocData{
		Content: text1,
		Tokens:  tokenDatas,
		Fields:  score1,
	})

	engine.Index(9, types.DocData{
		Content: text1,
		Fields:  score1,
	})

	engine.Index(10, types.DocData{
		Content: "Hello, 你好世界!",
		Tokens: []types.TokenData{{
			"世界",
			[]int{2, 3},
		}},
		Fields: score091,
	})

	engine.Flush()

	docIds := make(map[uint64]bool)
	for index := 0; index < 10; index++ {
		docIds[uint64(index)] = true
	}

	strArr := []string{"こんにちは"}
	logic := types.Logic{
		Should: true,
		LogicExpr: types.LogicExpr{
			NotInLabels: strArr,
		},
	}

	outputs := engine.Search(types.SearchReq{
		Text:   textJP,
		DocIds: docIds,
		Logic:  logic,
	})

	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "こんにちは", outputs.Tokens[0])
	tt.Expect(t, "世界", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	log.Println("outputs docs...", outDocs)
	tt.Expect(t, "2", len(outDocs))

	tt.Expect(t, "10", outDocs[0].DocId)
	tt.Expect(t, "1000", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[]", outDocs[0].TokenSnippetLocs)

	tt.Expect(t, "9", outDocs[1].DocId)
	tt.Expect(t, "1000", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[]", outDocs[1].TokenSnippetLocs)

	engine.Close()
}
