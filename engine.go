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
	// "reflect"

	"sync/atomic"

	"github.com/go-ego/riot/core"
	"github.com/go-ego/riot/storage"
	"github.com/go-ego/riot/types"
	"github.com/go-ego/riot/utils"

	"github.com/go-ego/gse"
	"github.com/go-ego/murmur"
	"github.com/shirou/gopsutil/mem"
)

const (
	version string = "v0.10.0.297, Danube River!"

	// NumNanosecondsInAMillisecond nano-seconds in a milli-second num
	NumNanosecondsInAMillisecond = 1000000
	// StorageFilePrefix persistent storage file prefix
	StorageFilePrefix = "riot"
)

// GetVersion get version
func GetVersion() string {
	return version
}

// Engine initialize the engine
type Engine struct {
	loc sync.RWMutex

	// 计数器，用来统计有多少文档被索引等信息
	numDocsIndexed       uint64
	numDocsRemoved       uint64
	numDocsForceUpdated  uint64
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
	stopTokens StopTokens
	dbs        []storage.Storage

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
	storageIndexDocChans []chan storageIndexDocReq
	storageInitChan      chan bool
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
			chan indexerAddDocReq,
			options.IndexerBufLen)
		engine.indexerRemoveDocChans[shard] = make(
			chan indexerRemoveDocReq,
			options.IndexerBufLen)
		engine.indexerLookupChans[shard] = make(
			chan indexerLookupReq,
			options.IndexerBufLen)
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
			chan rankerAddDocReq,
			options.RankerBufLen)
		engine.rankerRankChans[shard] = make(
			chan rankerRankReq,
			options.RankerBufLen)
		engine.rankerRemoveDocChans[shard] = make(
			chan rankerRemoveDocReq,
			options.RankerBufLen)
	}
}

// InitStorage initialize the persistent storage channel
func (engine *Engine) InitStorage() {
	engine.storageIndexDocChans =
		make([]chan storageIndexDocReq,
			engine.initOptions.StorageShards)
	for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
		engine.storageIndexDocChans[shard] = make(
			chan storageIndexDocReq)
	}
	engine.storageInitChan = make(
		chan bool, engine.initOptions.StorageShards)
}

// CheckMem check the memory when the memory is larger
// than 99.99% using the storage
func (engine *Engine) CheckMem() {
	// Todo test
	if !engine.initOptions.UseStorage {
		log.Println("Check virtualMemory...")
		vmem, _ := mem.VirtualMemory()
		log.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n",
			vmem.Total, vmem.Free, vmem.UsedPercent)

		useMem := fmt.Sprintf("%.2f", vmem.UsedPercent)
		if useMem == "99.99" {
			engine.initOptions.UseStorage = true
			engine.initOptions.StorageFolder = "./riot-index"
			os.MkdirAll("./riot-index", 0777)
		}
	}
}

// Storage start the persistent storage work connection
func (engine *Engine) Storage() {
	// if engine.initOptions.UseStorage {
	err := os.MkdirAll(engine.initOptions.StorageFolder, 0700)
	if err != nil {
		log.Fatal("Can not create directory ", engine.initOptions.StorageFolder)
	}

	// 打开或者创建数据库
	engine.dbs = make([]storage.Storage, engine.initOptions.StorageShards)
	for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
		dbPath := engine.initOptions.StorageFolder + "/" +
			StorageFilePrefix + "." + strconv.Itoa(shard)

		db, err := storage.OpenStorage(dbPath, engine.initOptions.StorageEngine)
		if db == nil || err != nil {
			log.Fatal("Unable to open database ", dbPath, ": ", err)
		}
		engine.dbs[shard] = db
	}

	// 从数据库中恢复
	for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
		go engine.storageInitWorker(shard)
	}

	// 等待恢复完成
	for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
		<-engine.storageInitChan
	}

	for {
		runtime.Gosched()

		engine.loc.RLock()
		numDoced := engine.numIndexingReqs == engine.numDocsIndexed
		engine.loc.RUnlock()

		if numDoced {
			break
		}

	}

	// 关闭并重新打开数据库
	for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
		engine.dbs[shard].Close()
		dbPath := engine.initOptions.StorageFolder + "/" +
			StorageFilePrefix + "." + strconv.Itoa(shard)

		db, err := storage.OpenStorage(dbPath, engine.initOptions.StorageEngine)
		if db == nil || err != nil {
			log.Fatal("Unable to open database ", dbPath, ": ", err)
		}
		engine.dbs[shard] = db
	}

	for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
		go engine.storageIndexDocWorker(shard)
	}
	// }
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
	options.Init()
	engine.initOptions = options
	engine.initialized = true

	if !options.NotUsingGse {
		// 载入分词器词典
		engine.segmenter.LoadDict(options.SegmenterDict)

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
		chan segmenterReq, options.NumSegmenterThreads)

	// 初始化索引器通道
	engine.Indexer(options)

	// 初始化排序器通道
	engine.Ranker(options)

	// engine.CheckMem(engine.initOptions.UseStorage)
	engine.CheckMem()

	// 初始化持久化存储通道
	if engine.initOptions.UseStorage {
		engine.InitStorage()
	}

	// 启动分词器
	for iThread := 0; iThread < options.NumSegmenterThreads; iThread++ {
		go engine.segmenterWorker()
	}

	// 启动索引器和排序器
	for shard := 0; shard < options.NumShards; shard++ {
		go engine.indexerAddDocWorker(shard)
		go engine.indexerRemoveDocWorker(shard)
		go engine.rankerAddDocWorker(shard)
		go engine.rankerRemoveDocWorker(shard)

		for i := 0; i < options.NumIndexerThreadsPerShard; i++ {
			go engine.indexerLookupWorker(shard)
		}
		for i := 0; i < options.NumRankerThreadsPerShard; i++ {
			go engine.rankerRankWorker(shard)
		}
	}

	// 启动持久化存储工作协程
	if engine.initOptions.UseStorage {
		engine.Storage()
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
func (engine *Engine) IndexDoc(docId uint64, data types.DocIndexData,
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

	hash := murmur.Sum32(fmt.Sprintf("%d", docId)) %
		uint32(engine.initOptions.StorageShards)

	if engine.initOptions.UseStorage && docId != 0 {
		engine.storageIndexDocChans[hash] <- storageIndexDocReq{
			docId: docId, data: data}
	}
}

func (engine *Engine) internalIndexDoc(docId uint64, data types.DocIndexData,
	forceUpdate bool) {

	if !engine.initialized {
		log.Fatal("The engine must be initialized first.")
	}

	if docId != 0 {
		atomic.AddUint64(&engine.numIndexingReqs, 1)
	}
	if forceUpdate {
		atomic.AddUint64(&engine.numForceUpdatingReqs, 1)
	}

	hash := murmur.Sum32(fmt.Sprintf("%d%s", docId, data.Content))
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
func (engine *Engine) RemoveDoc(docId uint64, forceUpdate ...bool) {
	var force bool
	if len(forceUpdate) > 0 {
		force = forceUpdate[0]
	}

	if !engine.initialized {
		log.Fatal("The engine must be initialized first.")
	}

	if docId != 0 {
		atomic.AddUint64(&engine.numRemovingReqs, 1)
	}
	if force {
		atomic.AddUint64(&engine.numForceUpdatingReqs, 1)
	}
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerRemoveDocChans[shard] <- indexerRemoveDocReq{
			docId: docId, forceUpdate: force}

		if docId == 0 {
			continue
		}
		engine.rankerRemoveDocChans[shard] <- rankerRemoveDocReq{docId: docId}
	}

	if engine.initOptions.UseStorage && docId != 0 {
		// 从数据库中删除
		hash := murmur.Sum32(fmt.Sprintf("%d", docId)) %
			uint32(engine.initOptions.StorageShards)

		go engine.storageRemoveDocWorker(docId, hash)
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
	segments := engine.segmenter.Segment([]byte(content))
	for _, segment := range segments {
		token := segment.Token().Text()
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
		request.Text = strings.ToLower(request.Text)
		if engine.initOptions.NotUsingGse {
			tokens = strings.Split(request.Text, " ")
		} else {
			// querySegments := engine.segmenter.Segment([]byte(request.Text))
			// for _, s := range querySegments {
			// 	token := s.Token().Text()
			// 	if !engine.stopTokens.IsStopToken(token) {
			// 		tokens = append(tokens, s.Token().Text())
			// 	}
			// }

			// tokens = engine.Tokens([]byte(request.Text))
			tokens = engine.Segment(request.Text)
		}

		// 叠加 tokens
		for _, t := range request.Tokens {
			tokens = append(tokens, t)
		}

	} else {
		for _, t := range request.Tokens {
			tokens = append(tokens, t)
		}
	}

	return
}

// Rank rank docs by types.ScoredIDs
func (engine *Engine) Rank(request types.SearchReq,
	RankOpts types.RankOpts, tokens []string,
	rankerReturnChan chan rankerReturnReq) (output types.SearchResp) {
	// 从通信通道读取排序器的输出
	numDocs := 0
	var rankOutput types.ScoredIDs
	// var rankOutput interface{}

	//**********/ begin
	timeout := request.Timeout
	isTimeout := false
	if timeout <= 0 {
		// 不设置超时
		for shard := 0; shard < engine.initOptions.NumShards; shard++ {
			rankerOutput := <-rankerReturnChan
			if !request.CountDocsOnly {
				if rankerOutput.docs != nil {
					for _, doc := range rankerOutput.docs.(types.ScoredIDs) {
						rankOutput = append(rankOutput, doc)
					}
				}
			}
			numDocs += rankerOutput.numDocs
		}
	} else {
		// 设置超时
		deadline := time.Now().Add(time.Nanosecond *
			time.Duration(NumNanosecondsInAMillisecond*request.Timeout))

		for shard := 0; shard < engine.initOptions.NumShards; shard++ {
			select {
			case rankerOutput := <-rankerReturnChan:
				if !request.CountDocsOnly {
					if rankerOutput.docs != nil {
						for _, doc := range rankerOutput.docs.(types.ScoredIDs) {
							rankOutput = append(rankOutput, doc)
						}
					}
				}
				numDocs += rankerOutput.numDocs
			case <-time.After(deadline.Sub(time.Now())):
				isTimeout = true
				break
			}
		}
	}

	// 再排序
	if !request.CountDocsOnly && !request.Orderless {
		if RankOpts.ReverseOrder {
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
			var start, end int
			if RankOpts.MaxOutputs == 0 {
				start = utils.MinInt(RankOpts.OutputOffset, len(rankOutput))
				end = len(rankOutput)
			} else {
				start = utils.MinInt(RankOpts.OutputOffset, len(rankOutput))
				end = utils.MinInt(start+RankOpts.MaxOutputs, len(rankOutput))
			}
			output.Docs = rankOutput[start:end]
		}
	}

	output.NumDocs = numDocs
	output.Timeout = isTimeout

	return
}

// Ranks rank docs by types.ScoredDocs
func (engine *Engine) Ranks(request types.SearchReq,
	RankOpts types.RankOpts, tokens []string,
	rankerReturnChan chan rankerReturnReq) (output types.SearchResp) {
	// 从通信通道读取排序器的输出
	numDocs := 0
	rankOutput := types.ScoredDocs{}

	//**********/ begin
	timeout := request.Timeout
	isTimeout := false
	if timeout <= 0 {
		// 不设置超时
		for shard := 0; shard < engine.initOptions.NumShards; shard++ {
			rankerOutput := <-rankerReturnChan
			if !request.CountDocsOnly {
				if rankerOutput.docs != nil {
					for _, doc := range rankerOutput.docs.(types.ScoredDocs) {
						rankOutput = append(rankOutput, doc)
					}
				}
			}
			numDocs += rankerOutput.numDocs
		}
	} else {
		// 设置超时
		deadline := time.Now().Add(time.Nanosecond *
			time.Duration(NumNanosecondsInAMillisecond*request.Timeout))

		for shard := 0; shard < engine.initOptions.NumShards; shard++ {
			select {
			case rankerOutput := <-rankerReturnChan:
				if !request.CountDocsOnly {
					if rankerOutput.docs != nil {
						for _, doc := range rankerOutput.docs.(types.ScoredDocs) {
							rankOutput = append(rankOutput, doc)
						}
					}
				}
				numDocs += rankerOutput.numDocs
			case <-time.After(deadline.Sub(time.Now())):
				isTimeout = true
				break
			}
		}
	}

	// 再排序
	if !request.CountDocsOnly && !request.Orderless {
		if RankOpts.ReverseOrder {
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
			var start, end int
			if RankOpts.MaxOutputs == 0 {
				start = utils.MinInt(RankOpts.OutputOffset, len(rankOutput))
				end = len(rankOutput)
			} else {
				start = utils.MinInt(RankOpts.OutputOffset, len(rankOutput))
				end = utils.MinInt(start+RankOpts.MaxOutputs, len(rankOutput))
			}
			output.Docs = rankOutput[start:end]
		}
	}

	output.NumDocs = numDocs
	output.Timeout = isTimeout

	return
}

// Search find the document that satisfies the search criteria.
// This function is thread safe
// 查找满足搜索条件的文档，此函数线程安全
func (engine *Engine) Search(request types.SearchReq) (output types.SearchResp) {
	if !engine.initialized {
		log.Fatal("The engine must be initialized first.")
	}

	tokens := engine.Tokens(request)

	var RankOpts types.RankOpts
	if request.RankOpts == nil {
		RankOpts = *engine.initOptions.DefaultRankOpts
	} else {
		RankOpts = *request.RankOpts
	}
	if RankOpts.ScoringCriteria == nil {
		RankOpts.ScoringCriteria = engine.initOptions.DefaultRankOpts.ScoringCriteria
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
		options:          RankOpts,
		rankerReturnChan: rankerReturnChan,
		orderless:        request.Orderless,
		logic:            request.Logic,
	}

	// 向索引器发送查找请求
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerLookupChans[shard] <- lookupRequest
	}

	if engine.initOptions.IDOnly {
		output = engine.Rank(request, RankOpts, tokens, rankerReturnChan)
		return
	}

	output = engine.Ranks(request, RankOpts, tokens, rankerReturnChan)
	return
}

// Flush block wait until all indexes are added
// 阻塞等待直到所有索引添加完毕
func (engine *Engine) Flush() {
	for {
		runtime.Gosched()

		engine.loc.RLock()
		inxd := engine.numIndexingReqs == engine.numDocsIndexed
		rmd := engine.numRemovingReqs*uint64(engine.initOptions.NumShards) ==
			engine.numDocsRemoved
		stored := !engine.initOptions.UseStorage || engine.numIndexingReqs ==
			engine.numDocsStored
		engine.loc.RUnlock()

		if inxd && rmd && stored {
			// 保证 CHANNEL 中 REQUESTS 全部被执行完
			break
		}
	}

	// 强制更新，保证其为最后的请求
	engine.IndexDoc(0, types.DocIndexData{}, true)
	for {
		runtime.Gosched()

		engine.loc.RLock()
		forced := engine.numForceUpdatingReqs*uint64(engine.initOptions.NumShards) ==
			engine.numDocsForceUpdated
		engine.loc.RUnlock()

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
	if engine.initOptions.UseStorage {
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
