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

/*

Package riot is riot engine
*/
package riot

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"sync/atomic"

	"github.com/go-ego/riot/core"
	"github.com/go-ego/riot/store"
	"github.com/go-ego/riot/types"
	"github.com/go-ego/riot/utils"

	"github.com/go-ego/gse"
	"github.com/go-ego/murmur"
	"github.com/shirou/gopsutil/mem"
)

const (
	// Version get the riot version
	Version string = "v0.10.0.425, Danube River!"

	// NumNanosecondsInAMillisecond nano-seconds in a milli-second num
	NumNanosecondsInAMillisecond = 1000000
	// StoreFilePrefix persistent store file prefix
	StoreFilePrefix = "riot"

	// DefaultPath default db path
	DefaultPath = "./riot-index"
)

// GetVersion get the riot version
func GetVersion() string {
	return Version
}

// Engine initialize the engine
type Engine struct {
	loc sync.RWMutex

	// 计数器，用来统计有多少文档被索引等信息
	numDocsIndexed      uint64
	numDocsRemoved      uint64
	numDocsForceUpdated uint64

	numIndexingReqs      uint64
	numRemovingReqs      uint64
	numForceUpdatingReqs uint64
	numTokenIndexAdded   uint64
	numDocsStored        uint64

	// 记录初始化参数
	initOptions types.EngineOpts
	initialized bool

	indexers   []core.Indexer
	rankers    []core.Ranker
	segmenter  gse.Segmenter
	loaded     bool
	stopTokens StopTokens
	dbs        []store.Store

	// 建立索引器使用的通信通道
	segmenterChan         chan segmenterReq
	indexerAddDocChans    []chan indexerAddDocReq
	indexerRemoveDocChans []chan indexerRemoveDocReq
	rankerAddDocChans     []chan rankerAddDocReq

	// 建立排序器使用的通信通道
	indexerLookupChans   []chan indexerLookupReq
	rankerRankChans      []chan rankerRankReq
	rankerRemoveDocChans []chan rankerRemoveDocReq

	// 建立持久存储使用的通信通道
	storeIndexDocChans []chan storeIndexDocReq
	storeInitChan      chan bool
}

// Indexer initialize the indexer channel
func (engine *Engine) Indexer(options types.EngineOpts) {
	engine.indexerAddDocChans = make(
		[]chan indexerAddDocReq, options.NumShards)

	engine.indexerRemoveDocChans = make(
		[]chan indexerRemoveDocReq, options.NumShards)

	engine.indexerLookupChans = make(
		[]chan indexerLookupReq, options.NumShards)

	for shard := 0; shard < options.NumShards; shard++ {
		engine.indexerAddDocChans[shard] = make(
			chan indexerAddDocReq, options.IndexerBufLen)

		engine.indexerRemoveDocChans[shard] = make(
			chan indexerRemoveDocReq, options.IndexerBufLen)

		engine.indexerLookupChans[shard] = make(
			chan indexerLookupReq, options.IndexerBufLen)
	}
}

// Ranker initialize the ranker channel
func (engine *Engine) Ranker(options types.EngineOpts) {
	engine.rankerAddDocChans = make(
		[]chan rankerAddDocReq, options.NumShards)

	engine.rankerRankChans = make(
		[]chan rankerRankReq, options.NumShards)

	engine.rankerRemoveDocChans = make(
		[]chan rankerRemoveDocReq, options.NumShards)

	for shard := 0; shard < options.NumShards; shard++ {
		engine.rankerAddDocChans[shard] = make(
			chan rankerAddDocReq, options.RankerBufLen)

		engine.rankerRankChans[shard] = make(
			chan rankerRankReq, options.RankerBufLen)

		engine.rankerRemoveDocChans[shard] = make(
			chan rankerRemoveDocReq, options.RankerBufLen)
	}
}

// InitStore initialize the persistent store channel
func (engine *Engine) InitStore() {
	engine.storeIndexDocChans = make(
		[]chan storeIndexDocReq, engine.initOptions.StoreShards)

	for shard := 0; shard < engine.initOptions.StoreShards; shard++ {
		engine.storeIndexDocChans[shard] = make(
			chan storeIndexDocReq)
	}
	engine.storeInitChan = make(
		chan bool, engine.initOptions.StoreShards)
}

// CheckMem check the memory when the memory is larger
// than 99.99% using the store
func (engine *Engine) CheckMem() {
	// Todo test
	if !engine.initOptions.UseStore {
		log.Println("Check virtualMemory...")

		vmem, _ := mem.VirtualMemory()
		log.Printf("Total: %v, Free: %v, UsedPercent: %f%%\n",
			vmem.Total, vmem.Free, vmem.UsedPercent)

		useMem := fmt.Sprintf("%.2f", vmem.UsedPercent)
		if useMem == "99.99" {
			engine.initOptions.UseStore = true
			engine.initOptions.StoreFolder = DefaultPath
			// os.MkdirAll(DefaultPath, 0777)
		}
	}
}

// Store start the persistent store work connection
func (engine *Engine) Store() {
	// if engine.initOptions.UseStore {
	err := os.MkdirAll(engine.initOptions.StoreFolder, 0700)
	if err != nil {
		log.Fatalf("Can not create directory: %s ; %v",
			engine.initOptions.StoreFolder, err)
	}

	// 打开或者创建数据库
	engine.dbs = make([]store.Store, engine.initOptions.StoreShards)
	for shard := 0; shard < engine.initOptions.StoreShards; shard++ {
		dbPath := engine.initOptions.StoreFolder + "/" +
			StoreFilePrefix + "." + strconv.Itoa(shard)

		db, err := store.OpenStore(dbPath, engine.initOptions.StoreEngine)
		if db == nil || err != nil {
			log.Fatal("Unable to open database ", dbPath, ": ", err)
		}
		engine.dbs[shard] = db
	}

	// 从数据库中恢复
	for shard := 0; shard < engine.initOptions.StoreShards; shard++ {
		go engine.storeInit(shard)
	}

	// 等待恢复完成
	for shard := 0; shard < engine.initOptions.StoreShards; shard++ {
		<-engine.storeInitChan
	}

	for {
		runtime.Gosched()

		inx := atomic.LoadUint64(&engine.numDocsIndexed)
		numDoced := engine.numIndexingReqs == inx

		if numDoced {
			break
		}

	}

	// 关闭并重新打开数据库
	for shard := 0; shard < engine.initOptions.StoreShards; shard++ {
		engine.dbs[shard].Close()
		dbPath := engine.initOptions.StoreFolder + "/" +
			StoreFilePrefix + "." + strconv.Itoa(shard)

		db, err := store.OpenStore(dbPath, engine.initOptions.StoreEngine)
		if db == nil || err != nil {
			log.Fatal("Unable to open database ", dbPath, ": ", err)
		}
		engine.dbs[shard] = db
	}

	for shard := 0; shard < engine.initOptions.StoreShards; shard++ {
		go engine.storeIndexDoc(shard)
	}
	// }
}

// WithGse Using user defined segmenter
// If using a not nil segmenter and the dictionary is loaded,
// the `opt.GseDict` will be ignore.
func (engine *Engine) WithGse(segmenter gse.Segmenter) *Engine {
	if engine.initialized {
		log.Fatal(`Do not re-initialize the engine, 
			WithGse should call before initialize the engine.`)
	}

	engine.segmenter = segmenter
	engine.loaded = true
	return engine
}

func (engine *Engine) initDef(options types.EngineOpts) types.EngineOpts {
	if options.GseDict == "" && !options.NotUseGse && !engine.loaded {
		log.Printf("Dictionary file path is empty, load the default dictionary file.")
		options.GseDict = "zh"
	}

	if options.UseStore == true && options.StoreFolder == "" {
		log.Printf("Store file path is empty, use default folder path.")
		options.StoreFolder = DefaultPath
		// os.MkdirAll(DefaultPath, 0777)
	}

	return options
}

// Init initialize the engine
func (engine *Engine) Init(options types.EngineOpts) {
	// 将线程数设置为CPU数
	// runtime.GOMAXPROCS(runtime.NumCPU())
	// runtime.GOMAXPROCS(128)

	// 初始化初始参数
	if engine.initialized {
		log.Fatal("Do not re-initialize the engine.")
	}
	options = engine.initDef(options)

	options.Init()
	engine.initOptions = options
	engine.initialized = true

	if !options.NotUseGse {
		if !engine.loaded {
			// 载入分词器词典
			engine.segmenter.LoadDict(options.GseDict)
			engine.loaded = true
		}

		// 初始化停用词
		engine.stopTokens.Init(options.StopTokenFile)
	}

	// 初始化索引器和排序器
	for shard := 0; shard < options.NumShards; shard++ {
		engine.indexers = append(engine.indexers, core.Indexer{})
		engine.indexers[shard].Init(*options.IndexerOpts)

		engine.rankers = append(engine.rankers, core.Ranker{})
		engine.rankers[shard].Init(options.IDOnly)
	}

	// 初始化分词器通道
	engine.segmenterChan = make(
		chan segmenterReq, options.NumGseThreads)

	// 初始化索引器通道
	engine.Indexer(options)

	// 初始化排序器通道
	engine.Ranker(options)

	// engine.CheckMem(engine.initOptions.UseStore)
	engine.CheckMem()

	// 初始化持久化存储通道
	if engine.initOptions.UseStore {
		engine.InitStore()
	}

	// 启动分词器
	for iThread := 0; iThread < options.NumGseThreads; iThread++ {
		go engine.segmenterWorker()
	}

	// 启动索引器和排序器
	for shard := 0; shard < options.NumShards; shard++ {
		go engine.indexerAddDoc(shard)
		go engine.indexerRemoveDoc(shard)
		go engine.rankerAddDoc(shard)
		go engine.rankerRemoveDoc(shard)

		for i := 0; i < options.NumIndexerThreads; i++ {
			go engine.indexerLookup(shard)
		}
		for i := 0; i < options.NumRankerThreads; i++ {
			go engine.rankerRank(shard)
		}
	}

	// 启动持久化存储工作协程
	if engine.initOptions.UseStore {
		engine.Store()
	}

	atomic.AddUint64(&engine.numDocsStored, engine.numIndexingReqs)
}

// IndexDoc add the document to the index
// 将文档加入索引
//
// 输入参数：
//  docId	      标识文档编号，必须唯一，docId == 0 表示非法文档（用于强制刷新索引），[1, +oo) 表示合法文档
//  data	      见 DocIndexData 注释
//  forceUpdate 是否强制刷新 cache，如果设为 true，则尽快添加到索引，否则等待 cache 满之后一次全量添加
//
// 注意：
//      1. 这个函数是线程安全的，请尽可能并发调用以提高索引速度
//      2. 这个函数调用是非同步的，也就是说在函数返回时有可能文档还没有加入索引中，因此
//         如果立刻调用Search可能无法查询到这个文档。强制刷新索引请调用FlushIndex函数。
func (engine *Engine) IndexDoc(docId string, data types.DocData,
	forceUpdate ...bool) {
	engine.Index(docId, data, forceUpdate...)
}

// Index add the document to the index
func (engine *Engine) Index(docId string, data types.DocData,
	forceUpdate ...bool) {

	var force bool
	if len(forceUpdate) > 0 {
		force = forceUpdate[0]
	}

	// if engine.HasDoc(docId) {
	// 	engine.RemoveDoc(docId)
	// }

	// data.Tokens
	engine.internalIndexDoc(docId, data, force)

	hash := murmur.Sum32(docId) % uint32(engine.initOptions.StoreShards)

	if engine.initOptions.UseStore && docId != "0" {
		engine.storeIndexDocChans[hash] <- storeIndexDocReq{
			docId: docId, data: data}
	}
}

func (engine *Engine) internalIndexDoc(docId string, data types.DocData,
	forceUpdate bool) {

	if !engine.initialized {
		log.Fatal("The engine must be initialized first.")
	}

	if docId != "0" {
		atomic.AddUint64(&engine.numIndexingReqs, 1)
	}
	if forceUpdate {
		atomic.AddUint64(&engine.numForceUpdatingReqs, 1)
	}

	hash := murmur.Sum32(fmt.Sprintf("%s%s", docId, data.Content))
	engine.segmenterChan <- segmenterReq{
		docId: docId, hash: hash, data: data, forceUpdate: forceUpdate}
}

// RemoveDoc remove the document from the index
// 将文档从索引中删除
//
// 输入参数：
//  docId	      标识文档编号，必须唯一，docId == 0 表示非法文档（用于强制刷新索引），[1, +oo) 表示合法文档
//  forceUpdate 是否强制刷新 cache，如果设为 true，则尽快删除索引，否则等待 cache 满之后一次全量删除
//
// 注意：
//      1. 这个函数是线程安全的，请尽可能并发调用以提高索引速度
//      2. 这个函数调用是非同步的，也就是说在函数返回时有可能文档还没有加入索引中，因此
//         如果立刻调用 Search 可能无法查询到这个文档。强制刷新索引请调用 FlushIndex 函数。
func (engine *Engine) RemoveDoc(docId string, forceUpdate ...bool) {
	var force bool
	if len(forceUpdate) > 0 {
		force = forceUpdate[0]
	}

	if !engine.initialized {
		log.Fatal("The engine must be initialized first.")
	}

	if docId != "0" {
		atomic.AddUint64(&engine.numRemovingReqs, 1)
	}

	if force {
		atomic.AddUint64(&engine.numForceUpdatingReqs, 1)
	}

	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerRemoveDocChans[shard] <- indexerRemoveDocReq{
			docId: docId, forceUpdate: force}

		if docId == "0" {
			continue
		}
		engine.rankerRemoveDocChans[shard] <- rankerRemoveDocReq{docId: docId}
	}

	if engine.initOptions.UseStore && docId != "0" {
		// 从数据库中删除
		hash := murmur.Sum32(docId) % uint32(engine.initOptions.StoreShards)

		go engine.storeRemoveDoc(docId, hash)
	}
}

// // 获取文本的分词结果
// func (engine *Engine) Tokens(text []byte) (tokens []string) {
// 	querySegments := engine.segmenter.Segment(text)
// 	for _, s := range querySegments {
// 		token := s.Token().Text()
// 		if !engine.stopTokens.IsStopToken(token) {
// 			tokens = append(tokens, token)
// 		}
// 	}
// 	return tokens
// }

// Segment get the word segmentation result of the text
// 获取文本的分词结果, 只分词与过滤弃用词
func (engine *Engine) Segment(content string) (keywords []string) {

	var segments []string
	hmm := engine.initOptions.Hmm

	if engine.initOptions.GseMode {
		segments = engine.segmenter.CutSearch(content, hmm)
	} else {
		segments = engine.segmenter.Cut(content, hmm)
	}

	for _, token := range segments {
		if !engine.stopTokens.IsStopToken(token) {
			keywords = append(keywords, token)
		}
	}

	return
}

// Tokens get the engine tokens
func (engine *Engine) Tokens(request types.SearchReq) (tokens []string) {
	// 收集关键词
	// tokens := []string{}
	if request.Text != "" {
		reqText := strings.ToLower(request.Text)
		if engine.initOptions.NotUseGse {
			tokens = strings.Split(reqText, " ")
		} else {
			// querySegments := engine.segmenter.Segment([]byte(reqText))
			// tokens = engine.Tokens([]byte(reqText))
			tokens = engine.Segment(reqText)
		}

		// 叠加 tokens
		for _, t := range request.Tokens {
			tokens = append(tokens, t)
		}

		return
	}

	for _, t := range request.Tokens {
		tokens = append(tokens, t)
	}
	return
}

func maxRankOutput(rankOpts types.RankOpts, rankLen int) (int, int) {
	var start, end int
	if rankOpts.MaxOutputs == 0 {
		start = utils.MinInt(rankOpts.OutputOffset, rankLen)
		end = rankLen
		return start, end
	}

	start = utils.MinInt(rankOpts.OutputOffset, rankLen)
	end = utils.MinInt(start+rankOpts.MaxOutputs, rankLen)
	return start, end
}

func (engine *Engine) rankOutID(rankerOutput rankerReturnReq,
	rankOutArr types.ScoredIDs) types.ScoredIDs {
	for _, doc := range rankerOutput.docs.(types.ScoredIDs) {
		rankOutArr = append(rankOutArr, doc)
	}
	return rankOutArr
}

func (engine *Engine) rankOutDocs(rankerOutput rankerReturnReq,
	rankOutArr types.ScoredDocs) types.ScoredDocs {
	for _, doc := range rankerOutput.docs.(types.ScoredDocs) {
		rankOutArr = append(rankOutArr, doc)
	}
	return rankOutArr
}

// NotTimeOut not set engine timeout
func (engine *Engine) NotTimeOut(request types.SearchReq,
	rankerReturnChan chan rankerReturnReq) (
	rankOutArr interface{}, numDocs int) {

	var (
		rankOutID  types.ScoredIDs
		rankOutDoc types.ScoredDocs
		idOnly     = engine.initOptions.IDOnly
	)

	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		rankerOutput := <-rankerReturnChan
		if !request.CountDocsOnly {
			if rankerOutput.docs != nil {
				if idOnly {
					rankOutID = engine.rankOutID(rankerOutput, rankOutID)
				} else {
					rankOutDoc = engine.rankOutDocs(rankerOutput, rankOutDoc)
				}
			}
		}
		numDocs += rankerOutput.numDocs
	}

	if idOnly {
		rankOutArr = rankOutID
		return
	}

	rankOutArr = rankOutDoc
	return
}

// TimeOut set engine timeout
func (engine *Engine) TimeOut(request types.SearchReq,
	rankerReturnChan chan rankerReturnReq) (
	rankOutArr interface{}, numDocs int, isTimeout bool) {

	deadline := time.Now().Add(time.Nanosecond *
		time.Duration(NumNanosecondsInAMillisecond*request.Timeout))

	var (
		rankOutID  types.ScoredIDs
		rankOutDoc types.ScoredDocs
		idOnly     = engine.initOptions.IDOnly
	)

	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		select {
		case rankerOutput := <-rankerReturnChan:
			if !request.CountDocsOnly {
				if rankerOutput.docs != nil {
					if idOnly {
						rankOutID = engine.rankOutID(rankerOutput, rankOutID)
					} else {
						rankOutDoc = engine.rankOutDocs(rankerOutput, rankOutDoc)
					}
				}
			}
			numDocs += rankerOutput.numDocs
		case <-time.After(deadline.Sub(time.Now())):
			isTimeout = true
			break
		}
	}

	if idOnly {
		rankOutArr = rankOutID
		return
	}

	rankOutArr = rankOutDoc
	return
}

// RankID rank docs by types.ScoredIDs
func (engine *Engine) RankID(request types.SearchReq, rankOpts types.RankOpts,
	tokens []string, rankerReturnChan chan rankerReturnReq) (output types.SearchResp) {
	// 从通信通道读取排序器的输出
	numDocs := 0
	rankOutput := types.ScoredIDs{}

	//**********/ begin
	timeout := request.Timeout
	isTimeout := false
	if timeout <= 0 {
		// 不设置超时
		rankOutArr, num := engine.NotTimeOut(request, rankerReturnChan)
		rankOutput = rankOutArr.(types.ScoredIDs)
		numDocs += num
	} else {
		// 设置超时
		rankOutArr, num, timeout := engine.TimeOut(request, rankerReturnChan)
		rankOutput = rankOutArr.(types.ScoredIDs)
		numDocs += num
		isTimeout = timeout
	}

	// 再排序
	if !request.CountDocsOnly && !request.Orderless {
		if rankOpts.ReverseOrder {
			sort.Sort(sort.Reverse(rankOutput))
		} else {
			sort.Sort(rankOutput)
		}
	}

	// 准备输出
	output.Tokens = tokens
	// 仅当 CountDocsOnly 为 false 时才充填 output.Docs
	if !request.CountDocsOnly {
		if request.Orderless {
			// 无序状态无需对 Offset 截断
			output.Docs = rankOutput
		} else {
			rankOutLen := len(rankOutput)
			start, end := maxRankOutput(rankOpts, rankOutLen)

			output.Docs = rankOutput[start:end]
		}
	}

	output.NumDocs = numDocs
	output.Timeout = isTimeout

	return
}

// Ranks rank docs by types.ScoredDocs
func (engine *Engine) Ranks(request types.SearchReq, rankOpts types.RankOpts,
	tokens []string, rankerReturnChan chan rankerReturnReq) (output types.SearchResp) {
	// 从通信通道读取排序器的输出
	numDocs := 0
	rankOutput := types.ScoredDocs{}

	//**********/ begin
	timeout := request.Timeout
	isTimeout := false
	if timeout <= 0 {
		// 不设置超时
		rankOutArr, num := engine.NotTimeOut(request, rankerReturnChan)
		rankOutput = rankOutArr.(types.ScoredDocs)
		numDocs += num
	} else {
		// 设置超时
		rankOutArr, num, timeout := engine.TimeOut(request, rankerReturnChan)
		rankOutput = rankOutArr.(types.ScoredDocs)
		numDocs += num
		isTimeout = timeout
	}

	// 再排序
	if !request.CountDocsOnly && !request.Orderless {
		if rankOpts.ReverseOrder {
			sort.Sort(sort.Reverse(rankOutput))
		} else {
			sort.Sort(rankOutput)
		}
	}

	// 准备输出
	output.Tokens = tokens
	// 仅当 CountDocsOnly 为 false 时才充填 output.Docs
	if !request.CountDocsOnly {
		if request.Orderless {
			// 无序状态无需对 Offset 截断
			output.Docs = rankOutput
		} else {
			rankOutLen := len(rankOutput)
			start, end := maxRankOutput(rankOpts, rankOutLen)

			output.Docs = rankOutput[start:end]
		}
	}

	output.NumDocs = numDocs
	output.Timeout = isTimeout

	return
}

// SearchDoc find the document that satisfies the search criteria.
// This function is thread safe, return not IDonly
func (engine *Engine) SearchDoc(request types.SearchReq) (output types.SearchDoc) {
	resp := engine.Search(request)
	return types.SearchDoc{
		BaseResp: resp.BaseResp,
		Docs:     resp.Docs.(types.ScoredDocs),
	}
}

// SearchID find the document that satisfies the search criteria.
// This function is thread safe, return IDonly
func (engine *Engine) SearchID(request types.SearchReq) (output types.SearchID) {
	// return types.SearchID(engine.Search(request))
	resp := engine.Search(request)
	return types.SearchID{
		BaseResp: resp.BaseResp,
		Docs:     resp.Docs.(types.ScoredIDs),
	}
}

// Search find the document that satisfies the search criteria.
// This function is thread safe
// 查找满足搜索条件的文档，此函数线程安全
func (engine *Engine) Search(request types.SearchReq) (output types.SearchResp) {
	if !engine.initialized {
		log.Fatal("The engine must be initialized first.")
	}

	tokens := engine.Tokens(request)

	var rankOpts types.RankOpts
	if request.RankOpts == nil {
		rankOpts = *engine.initOptions.DefRankOpts
	} else {
		rankOpts = *request.RankOpts
	}

	if rankOpts.ScoringCriteria == nil {
		rankOpts.ScoringCriteria = engine.initOptions.DefRankOpts.ScoringCriteria
	}

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

	if engine.initOptions.IDOnly {
		output = engine.RankID(request, rankOpts, tokens, rankerReturnChan)
		return
	}

	output = engine.Ranks(request, rankOpts, tokens, rankerReturnChan)
	return
}

// Flush block wait until all indexes are added
// 阻塞等待直到所有索引添加完毕
func (engine *Engine) Flush() {
	for {
		runtime.Gosched()

		inxd := engine.numIndexingReqs == atomic.LoadUint64(&engine.numDocsIndexed)
		numRm := engine.numRemovingReqs * uint64(engine.initOptions.NumShards)
		rmd := numRm == atomic.LoadUint64(&engine.numDocsRemoved)

		nums := engine.numIndexingReqs == atomic.LoadUint64(&engine.numDocsStored)
		stored := !engine.initOptions.UseStore || nums

		if inxd && rmd && stored {
			// 保证 CHANNEL 中 REQUESTS 全部被执行完
			break
		}
	}

	// 强制更新，保证其为最后的请求
	engine.IndexDoc("0", types.DocData{}, true)
	for {
		runtime.Gosched()

		numf := engine.numForceUpdatingReqs * uint64(engine.initOptions.NumShards)
		forced := numf == atomic.LoadUint64(&engine.numDocsForceUpdated)

		if forced {
			return
		}

	}
}

// FlushIndex block wait until all indexes are added
// 阻塞等待直到所有索引添加完毕
func (engine *Engine) FlushIndex() {
	engine.Flush()
}

// Close close the engine
// 关闭引擎
func (engine *Engine) Close() {
	engine.Flush()
	if engine.initOptions.UseStore {
		for _, db := range engine.dbs {
			db.Close()
		}
	}
}

// 从文本hash得到要分配到的 shard
func (engine *Engine) getShard(hash uint32) int {
	return int(hash - hash/uint32(engine.initOptions.NumShards)*
		uint32(engine.initOptions.NumShards))
}
