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

var text2 = "在路上, in the way"

func makeDocIds() map[uint64]bool {
	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[3] = true
	docIds[1] = true
	docIds[2] = true

	return docIds
}

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

	outputs := engine1.Search(types.SearchReq{Text: reqText})
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "world", outputs.Tokens[0])
	tt.Expect(t, "人口", outputs.Tokens[1])

	outDocs := outputs.Docs.(types.ScoredDocs)
	tt.Expect(t, "2", len(outDocs))

	// tt.Expect(t, "2", outDocs[0].DocId)
	tt.Expect(t, "2500", int(outDocs[0].Scores[0]*1000))
	tt.Expect(t, "[]", outDocs[0].TokenSnippetLocs)

	// tt.Expect(t, "1", outDocs[1].DocId)
	tt.Expect(t, "2215", int(outDocs[1].Scores[0]*1000))
	tt.Expect(t, "[]", outDocs[1].TokenSnippetLocs)

	engine1.Close()
	os.RemoveAll("riot.new")
	// os.RemoveAll("riot-index")
}

var (
	rankTestOpts = rankOptsMax(0, 1)
)

func testRankOpt(idOnly bool) types.EngineOpts {
	return types.EngineOpts{
		Using:       1,
		IDOnly:      idOnly,
		GseDict:     "./testdata/test_dict.txt",
		DefRankOpts: &rankTestOpts,
		IndexerOpts: inxOpts,
	}
}

func lookupReq(engine *Engine) (types.SearchReq, []string, chan rankerReturnReq) {
	request := types.SearchReq{
		Text:   reqText,
		DocIds: makeDocIds(),
	}

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
		options:          rankTestOpts,
		rankerReturnChan: rankerReturnChan,
		orderless:        request.Orderless,
		logic:            request.Logic,
	}

	// 向索引器发送查找请求
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerLookupChans[shard] <- lookupRequest
	}

	return request, tokens, rankerReturnChan
}

func TestDocRankID(t *testing.T) {
	var engine Engine

	engine.Init(testRankOpt(true))
	AddDocs(&engine)

	engine.RemoveDoc(5)
	engine.Flush()

	request, tokens, rankerReturnChan := lookupReq(&engine)
	outputs := engine.RankID(request, rankTestOpts, tokens, rankerReturnChan)

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredIDs)
		tt.Expect(t, "1", len(outDocs))
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "2", outputs.NumDocs)

	engine.Close()
}

func TestDocRanks(t *testing.T) {
	var engine Engine

	engine.Init(testRankOpt(false))
	AddDocs(&engine)

	engine.RemoveDoc(5)
	engine.Flush()

	request, tokens, rankerReturnChan := lookupReq(&engine)
	outputs := engine.Ranks(request, rankTestOpts, tokens, rankerReturnChan)

	if outputs.Docs != nil {
		outDocs := outputs.Docs.(types.ScoredDocs)
		tt.Expect(t, "1", len(outDocs))
	}
	tt.Expect(t, "2", len(outputs.Tokens))
	tt.Expect(t, "2", outputs.NumDocs)

	// test search
	outputs1 := engine.Search(types.SearchReq{
		Text:    reqText,
		Timeout: 1000,
		DocIds:  makeDocIds()})

	if outputs1.Docs != nil {
		outDocs1 := outputs.Docs.(types.ScoredDocs)
		tt.Expect(t, "1", len(outDocs1))
	}
	tt.Expect(t, "2", len(outputs1.Tokens))
	tt.Expect(t, "2", outputs1.NumDocs)

	engine.Close()
}

func TestDocGetAllDocAndID(t *testing.T) {
	gob.Register(ScoringFields{})

	var engine Engine
	opts := types.EngineOpts{
		Using:     1,
		NumShards: 5,
		UseStore:  true,
		// StoreEngine: "bg",
		StoreFolder: "riot.id",
		IDOnly:      true,
		GseDict:     "./testdata/test_dict.txt",
		DefRankOpts: &rankTestOpts,
		IndexerOpts: inxOpts,
	}
	engine.Init(opts)

	AddDocs(&engine)

	engine.RemoveDoc(5)
	engine.Flush()

	allIds := engine.GetDBAllIds()
	fmt.Println("all id", allIds)
	tt.Expect(t, "5", len(allIds))
	tt.Expect(t, "[3 4 1 6 2]", allIds)

	allIds = engine.GetAllDocIds()
	fmt.Println("all doc id", allIds)
	tt.Expect(t, "5", len(allIds))
	tt.Expect(t, "[3 4 1 6 2]", allIds)

	ids, docs := engine.GetDBAllDocs()
	fmt.Println("all id and doc", allIds, docs)
	tt.Expect(t, "5", len(ids))
	tt.Expect(t, "5", len(docs))
	tt.Expect(t, "[3 4 1 6 2]", ids)
	allDoc := `[{The world <nil> [] [] <nil>} {有人口 <nil> [] [] {2 3 1}} {The world, 有七十亿人口人口 <nil> [] [] {1 2 3}} {有七十亿人口 <nil> [] [] {2 3 3}} {The world, 人口 <nil> [] [] <nil>}]`
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
		Text:   reqText,
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

func testOpts(use int, store string, args ...bool) types.EngineOpts {
	var pinyin bool
	if len(args) > 0 {
		pinyin = args[0]
	}

	return types.EngineOpts{
		// Using:      1,
		Using:       use,
		UseStore:    true,
		StoreFolder: store,
		PinYin:      pinyin,
		IDOnly:      true,
		GseDict:     "./testdata/test_dict.txt",
	}
}

func TestDocPinYin(t *testing.T) {
	var engine, pinyinOpt Engine
	engine.Init(testOpts(0, "riot.py"))
	pinyinOpt.Init(testOpts(0, "riot.py.opt", true))

	// AddDocs(&engine)
	// engine.RemoveDoc(5)

	tokens := engine.PinYin(text2)
	fmt.Println("tokens...", tokens)
	tt.Expect(t, "52", len(tokens))

	var tokenDatas []types.TokenData
	// tokens := []string{"z", "zl"}
	for i := 0; i < len(tokens); i++ {
		tokenData := types.TokenData{Text: tokens[i]}
		tokenDatas = append(tokenDatas, tokenData)
	}

	index1 := types.DocData{Tokens: tokenDatas, Fields: "在路上"}
	index2 := types.DocData{Content: text2, Tokens: tokenDatas}

	engine.Index(10, index1)
	engine.Index(11, index2)
	engine.Flush()

	data := types.DocData{Content: text2}
	pinyinOpt.Index(10, data)
	pinyinOpt.Index(11, data)
	pinyinOpt.Flush()

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[10] = true
	docIds[11] = true

	pyOutputs := pinyinOpt.SearchID(types.SearchReq{
		Text:   "zl",
		DocIds: docIds,
	})

	tt.Expect(t, "2", len(pyOutputs.Docs))
	tt.Expect(t, "1", len(pyOutputs.Tokens))
	tt.Expect(t, "2", pyOutputs.NumDocs)

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
	pinyinOpt.Close()
	os.RemoveAll("riot.py")
	os.RemoveAll("riot.py.opt")
}

func TestForSplitData(t *testing.T) {
	var engine Engine
	engine.Init(testOpts(4, "riot.data"))

	AddDocs(&engine)

	engine.RemoveDoc(5)
	engine.Flush()

	tokenDatas := engine.PinYin(text2)
	tokens, num := engine.ForSplitData(tokenDatas, 52)
	tt.Expect(t, "93", len(tokens))
	tt.Expect(t, "104", num)

	index1 := types.DocData{Content: "在路上"}
	engine.Index(10, index1, true)

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true
	outputs := engine.Search(types.SearchReq{
		Text:   reqText,
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

func testNum(t *testing.T, numAdd, numInx, numRm uint64) {
	tt.Expect(t, "26", numAdd)
	tt.Expect(t, "6", numInx)
	tt.Expect(t, "8", numRm)
}
func TestDocCounters(t *testing.T) {
	var engine Engine
	engine.Init(testOpts(1, "riot.doc"))

	AddDocs(&engine)
	engine.RemoveDoc(5)
	engine.Flush()

	numAdd := engine.NumTokenAdded()
	numInx := engine.NumIndexed()
	numRm := engine.NumRemoved()
	testNum(t, numAdd, numInx, numRm)

	numAdd = engine.NumTokenIndexAdded()
	numInx = engine.NumDocsIndexed()
	numRm = engine.NumDocsRemoved()
	testNum(t, numAdd, numInx, numRm)

	docIds := make(map[uint64]bool)
	docIds[5] = true
	docIds[1] = true

	outputs := engine.Search(types.SearchReq{
		Text:   reqText,
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
