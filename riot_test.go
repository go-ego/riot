package riot

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/go-ego/riot/types"
	"github.com/vcaesar/tt"
)

func TestEngineIndexWithNewStore(t *testing.T) {
	gob.Register(ScoringFields{})
	var engine = New("./testdata/test_dict.txt", "./riot.new", 8)
	log.Println("new engine start...")
	// engine = engine.New()
	AddDocs(engine)

	engine.RemoveDoc(5, true)
	engine.Flush()

	engine.Close()
	// os.RemoveAll("riot.new")

	// var engine1 = New("./testdata/test_dict.txt", "./riot.new")
	var engine1 = New("./testdata/test_new.toml")
	// engine1 = engine1.New()
	log.Println("test...")
	engine1.Flush()
	log.Println("new engine1 start...")

	outputs := engine1.Search(types.SearchReq{Text: "中国人口"})
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "中国", outputs.Tokens[0])
	tt.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))

	// tt.Expect(t, "2", outDocs[0].DocId)
	tt.Expect(t, "2500", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[]", outDocs[0].TokenSnippetLocs)

	// tt.Expect(t, "1", outDocs[1].DocId)
	tt.Expect(t, "2000", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[]", outDocs[1].TokenSnippetLocs)

	engine1.Close()
	os.RemoveAll("riot.new")
	// os.RemoveAll("riot-index")
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
		GseDict:         "./testdata/test_dict.txt",
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
		tt.Expect(t, "1", len(outDocs))
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "1", outputs.NumDocs)

	engine.Close()
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
		GseDict:         "./testdata/test_dict.txt",
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
		tt.Expect(t, "1", len(outDocs))
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "1", outputs.NumDocs)

	outputs1 := engine.Search(types.SearchReq{
		Text:    "中国人口",
		Timeout: 10,
		DocIds:  docIds})

	if outputs1.Docs != nil {
		outDocs1 := outputs.Docs.(types.ScoredDocs)
		tt.Expect(t, "1", len(outDocs1))
	}
	tt.Expect(t, "2", len(outputs1.Tokens))
	tt.Expect(t, "1", outputs1.NumDocs)

	engine.Close()
}

func TestDocGetAllDocAndID(t *testing.T) {
	gob.Register(ScoringFields{})

	var engine Engine
	engine.Init(types.EngineOpts{
		Using:     1,
		NumShards: 5,
		UseStore:  true,
		// StoreEngine: "bg",
		StoreFolder: "riot.id",
		IDOnly:      true,
		GseDict:     "./testdata/test_dict.txt",
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
	tt.Expect(t, "4", len(allIds))
	tt.Expect(t, "[3 4 1 2]", allIds)

	allIds = engine.GetAllDocIds()
	fmt.Println("all doc id", allIds)
	tt.Expect(t, "4", len(allIds))
	tt.Expect(t, "[3 4 1 2]", allIds)

	ids, docs := engine.GetDBAllDocs()
	fmt.Println("all id and doc", allIds, docs)
	tt.Expect(t, "4", len(ids))
	tt.Expect(t, "4", len(docs))
	tt.Expect(t, "[3 4 1 2]", ids)
	allDoc := `[{有人口 <nil> [] [] {2 3 1}} {有十三亿人口 <nil> [] [] {2 3 3}} {中国有十三亿人口人口 <nil> [] [] {1 2 3}} {中国人口 <nil> [] [] <nil>}]`
	tt.Expect(t, allDoc, docs)

	has := engine.HasDoc(5)
	tt.Expect(t, "false", has)

	has = engine.HasDoc(2)
	tt.Equal(t, true, has)
	has = engine.HasDoc(3)
	tt.Equal(t, true, has)
	has = engine.HasDoc(4)
	tt.Expect(t, "true", has)

	dbhas := engine.HasDocDB(5)
	tt.Expect(t, "false", dbhas)

	dbhas = engine.HasDocDB(2)
	tt.Equal(t, true, dbhas)
	dbhas = engine.HasDocDB(3)
	tt.Equal(t, true, dbhas)
	dbhas = engine.HasDocDB(4)
	tt.Expect(t, "true", dbhas)

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true

	outputs := engine.Search(types.SearchReq{
		Text:   "中国人口",
		DocIds: docIds})

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		fmt.Println("output docs: ", outputs)
		tt.Expect(t, "1", len(outDocs))
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "1", outputs.NumDocs)

	engine.Close()
	os.RemoveAll("riot.id")
}

func TestDocPinYin(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		// Using:         1,
		UseStore:    true,
		StoreFolder: "riot.py",
		IDOnly:      true,
		GseDict:     "./testdata/test_dict.txt",
	})

	// AddDocs(&engine)
	// engine.RemoveDoc(5)

	tokens := engine.PinYin("在路上, in the way")

	fmt.Println("tokens...", tokens)
	tt.Expect(t, "52", len(tokens))

	var tokenDatas []types.TokenData
	// tokens := []string{"z", "zl"}
	for i := 0; i < len(tokens); i++ {
		tokenData := types.TokenData{Text: tokens[i]}
		tokenDatas = append(tokenDatas, tokenData)
	}

	index1 := types.DocData{Tokens: tokenDatas, Fields: "在路上"}
	index2 := types.DocData{Content: "在路上, in the way",
		Tokens: tokenDatas}

	engine.Index(10, index1)
	engine.Index(11, index2)

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
		tt.Expect(t, "2", len(outDocs))
		// tt.Expect(t, "11", outDocs[0].DocId)
		// tt.Expect(t, "10", outDocs[1].DocId)
	}
	tt.Expect(t, "1", len(outputs.Tokens))
	tt.Expect(t, "2", outputs.NumDocs)

	engine.Close()
	os.RemoveAll("riot.py")
}

func TestForSplitData(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:       4,
		UseStore:    true,
		StoreFolder: "riot.data",
		IDOnly:      true,
		GseDict:     "./testdata/test_dict.txt",
	})

	AddDocs(&engine)

	engine.RemoveDoc(5)
	engine.Flush()

	tokenDatas := engine.PinYin("在路上, in the way")
	tokens, num := engine.ForSplitData(tokenDatas, 52)
	tt.Expect(t, "93", len(tokens))
	tt.Expect(t, "104", num)

	index1 := types.DocData{Content: "在路上"}
	engine.Index(10, index1, true)

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	outputs := engine.Search(types.SearchReq{
		Text:   "中国人口",
		DocIds: docIds})

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		tt.Expect(t, "0", len(outDocs))
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "0", outputs.NumDocs)

	engine.Close()
	os.RemoveAll("riot.data")
}

func TestDocCounters(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineOpts{
		Using:       1,
		UseStore:    true,
		StoreFolder: "riot.doc",
		IDOnly:      true,
		GseDict:     "./testdata/test_dict.txt",
	})

	AddDocs(&engine)
	engine.RemoveDoc(5)
	engine.Flush()

	numAdd := engine.NumTokenIndexAdded()
	tt.Expect(t, "14", numAdd)
	numInx := engine.NumDocsIndexed()
	tt.Expect(t, "5", numInx)
	numRm := engine.NumDocsRemoved()
	tt.Expect(t, "8", numRm)

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

	engine.Close()
	os.RemoveAll("riot.doc")
}
