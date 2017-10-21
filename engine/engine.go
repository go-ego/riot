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

Package engine is riot engine
*/
package engine

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	// "reflect"

	"github.com/go-ego/riot/core"
	"github.com/go-ego/riot/storage"
	"github.com/go-ego/riot/types"
	"github.com/go-ego/riot/utils"

	"github.com/go-ego/gse"
	"github.com/go-ego/murmur"
	"github.com/shirou/gopsutil/mem"
)

const (
	version string = "v0.10.0.103, Mount Qomolangma!"

	// NumNanosecondsInAMillisecond nano-seconds in a milli-second num
	NumNanosecondsInAMillisecond = 1000000
	// PersistentStorageFilePrefix persistent storage file prefix
	PersistentStorageFilePrefix = "riot"
)

// GetVersion get version
func GetVersion() string {
	return version
}

// Engine initialize the engine
type Engine struct {
	// 计数器，用来统计有多少文档被索引等信息
	numDocumentsIndexed      uint64
	numDocumentsRemoved      uint64
	numDocumentsForceUpdated uint64
	numIndexingRequests      uint64
	numRemovingRequests      uint64
	numForceUpdatingRequests uint64
	numTokenIndexAdded       uint64
	numDocumentsStored       uint64

	// 记录初始化参数
	initOptions types.EngineInitOptions
	initialized bool

	indexers   []core.Indexer
	rankers    []core.Ranker
	segmenter  gse.Segmenter
	stopTokens StopTokens
	dbs        []storage.Storage

	// 建立索引器使用的通信通道
	segmenterChannel         chan segmenterRequest
	indexerAddDocChannels    []chan indexerAddDocumentRequest
	indexerRemoveDocChannels []chan indexerRemoveDocRequest
	rankerAddDocChannels     []chan rankerAddDocRequest

	// 建立排序器使用的通信通道
	indexerLookupChannels   []chan indexerLookupRequest
	rankerRankChannels      []chan rankerRankRequest
	rankerRemoveDocChannels []chan rankerRemoveDocRequest

	// 建立持久存储使用的通信通道
	storageIndexDocChannels      []chan storageIndexDocRequest
	persistentStorageInitChannel chan bool
}

// Indexer initialize the indexer channel
func (engine *Engine) Indexer(options types.EngineInitOptions) {
	engine.indexerAddDocChannels = make(
		[]chan indexerAddDocumentRequest, options.NumShards)
	engine.indexerRemoveDocChannels = make(
		[]chan indexerRemoveDocRequest, options.NumShards)
	engine.indexerLookupChannels = make(
		[]chan indexerLookupRequest, options.NumShards)
	for shard := 0; shard < options.NumShards; shard++ {
		engine.indexerAddDocChannels[shard] = make(
			chan indexerAddDocumentRequest,
			options.IndexerBufferLength)
		engine.indexerRemoveDocChannels[shard] = make(
			chan indexerRemoveDocRequest,
			options.IndexerBufferLength)
		engine.indexerLookupChannels[shard] = make(
			chan indexerLookupRequest,
			options.IndexerBufferLength)
	}
}

// Ranker initialize the ranker channel
func (engine *Engine) Ranker(options types.EngineInitOptions) {
	engine.rankerAddDocChannels = make(
		[]chan rankerAddDocRequest, options.NumShards)
	engine.rankerRankChannels = make(
		[]chan rankerRankRequest, options.NumShards)
	engine.rankerRemoveDocChannels = make(
		[]chan rankerRemoveDocRequest, options.NumShards)
	for shard := 0; shard < options.NumShards; shard++ {
		engine.rankerAddDocChannels[shard] = make(
			chan rankerAddDocRequest,
			options.RankerBufferLength)
		engine.rankerRankChannels[shard] = make(
			chan rankerRankRequest,
			options.RankerBufferLength)
		engine.rankerRemoveDocChannels[shard] = make(
			chan rankerRemoveDocRequest,
			options.RankerBufferLength)
	}
}

// InitStorage initialize the persistent storage channel
func (engine *Engine) InitStorage() {
	if engine.initOptions.UseStorage {
		engine.storageIndexDocChannels =
			make([]chan storageIndexDocRequest,
				engine.initOptions.StorageShards)
		for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
			engine.storageIndexDocChannels[shard] = make(
				chan storageIndexDocRequest)
		}
		engine.persistentStorageInitChannel = make(
			chan bool, engine.initOptions.StorageShards)
	}
}

// CheckMem check the memory when the memory is larger than 99.99% using the storage
func (engine *Engine) CheckMem() {
	// Todo test
	if !engine.initOptions.UseStorage {
		log.Println("Check virtualMemory...")
		vmem, _ := mem.VirtualMemory()
		fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", vmem.Total, vmem.Free, vmem.UsedPercent)
		useMem := fmt.Sprintf("%.2f", vmem.UsedPercent)
		if useMem == "99.99" {
			engine.initOptions.UseStorage = true
			engine.initOptions.StorageFolder = "./index"
			os.MkdirAll("./index", 0777)
		}
	}
}

// Storage start the persistent storage work connection
func (engine *Engine) Storage() {
	if engine.initOptions.UseStorage {
		err := os.MkdirAll(engine.initOptions.StorageFolder, 0700)
		if err != nil {
			log.Fatal("无法创建目录", engine.initOptions.StorageFolder)
		}

		// 打开或者创建数据库
		engine.dbs = make([]storage.Storage, engine.initOptions.StorageShards)
		for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
			dbPath := engine.initOptions.StorageFolder + "/" + PersistentStorageFilePrefix + "." + strconv.Itoa(shard)
			db, err := storage.OpenStorage(dbPath, engine.initOptions.StorageEngine)
			if db == nil || err != nil {
				log.Fatal("无法打开数据库", dbPath, ": ", err)
			}
			engine.dbs[shard] = db
		}

		// 从数据库中恢复
		for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
			go engine.storageInitWorker(shard)
		}

		// 等待恢复完成
		for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
			<-engine.persistentStorageInitChannel
		}
		for {
			runtime.Gosched()
			if engine.numIndexingRequests == engine.numDocumentsIndexed {
				break
			}
		}

		// 关闭并重新打开数据库
		for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
			engine.dbs[shard].Close()
			dbPath := engine.initOptions.StorageFolder + "/" + PersistentStorageFilePrefix + "." + strconv.Itoa(shard)
			db, err := storage.OpenStorage(dbPath, engine.initOptions.StorageEngine)
			if db == nil || err != nil {
				log.Fatal("无法打开数据库", dbPath, ": ", err)
			}
			engine.dbs[shard] = db
		}

		for shard := 0; shard < engine.initOptions.StorageShards; shard++ {
			go engine.storageIndexDocWorker(shard)
		}
	}
}

// Init initialize the engine
func (engine *Engine) Init(options types.EngineInitOptions) {
	// 将线程数设置为CPU数
	// runtime.GOMAXPROCS(runtime.NumCPU())
	// runtime.GOMAXPROCS(128)

	// 初始化初始参数
	if engine.initialized {
		log.Fatal("请勿重复初始化引擎")
	}
	options.Init()
	engine.initOptions = options
	engine.initialized = true

	if !options.NotUsingSegmenter {
		// 载入分词器词典
		engine.segmenter.LoadDict(options.SegmenterDict)

		// 初始化停用词
		engine.stopTokens.Init(options.StopTokenFile)
	}

	// 初始化索引器和排序器
	for shard := 0; shard < options.NumShards; shard++ {
		engine.indexers = append(engine.indexers, core.Indexer{})
		engine.indexers[shard].Init(*options.IndexerInitOptions)

		engine.rankers = append(engine.rankers, core.Ranker{})
		engine.rankers[shard].Init()
	}

	// 初始化分词器通道
	engine.segmenterChannel = make(
		chan segmenterRequest, options.NumSegmenterThreads)

	// 初始化索引器通道
	engine.Indexer(options)

	// 初始化排序器通道
	engine.Ranker(options)

	// engine.CheckMem(engine.initOptions.UseStorage)
	engine.CheckMem()

	// 初始化持久化存储通道
	engine.InitStorage()

	// 启动分词器
	for iThread := 0; iThread < options.NumSegmenterThreads; iThread++ {
		go engine.segmenterWorker()
	}

	// 启动索引器和排序器
	for shard := 0; shard < options.NumShards; shard++ {
		go engine.indexerAddDocumentWorker(shard)
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
	engine.Storage()

	atomic.AddUint64(&engine.numDocumentsStored, engine.numIndexingRequests)
}

// IndexDocument add the document to the index
// 将文档加入索引
//
// 输入参数：
//  docId	      标识文档编号，必须唯一，docId == 0 表示非法文档（用于强制刷新索引），[1, +oo) 表示合法文档
//  data	      见DocIndexData注释
//  forceUpdate 是否强制刷新 cache，如果设为 true，则尽快添加到索引，否则等待 cache 满之后一次全量添加
//
// 注意：
//      1. 这个函数是线程安全的，请尽可能并发调用以提高索引速度
//      2. 这个函数调用是非同步的，也就是说在函数返回时有可能文档还没有加入索引中，因此
//         如果立刻调用Search可能无法查询到这个文档。强制刷新索引请调用FlushIndex函数。
func (engine *Engine) IndexDocument(docId uint64, data types.DocIndexData, forceUpdate bool) {
	// data.Tokens
	engine.internalIndexDocument(docId, data, forceUpdate)

	hash := murmur.Murmur3([]byte(fmt.Sprintf("%d", docId))) % uint32(engine.initOptions.StorageShards)
	if engine.initOptions.UseStorage && docId != 0 {
		engine.storageIndexDocChannels[hash] <- storageIndexDocRequest{docId: docId, data: data}
	}
}

func (engine *Engine) internalIndexDocument(
	docId uint64, data types.DocIndexData, forceUpdate bool) {
	if !engine.initialized {
		log.Fatal("必须先初始化引擎")
	}

	if docId != 0 {
		atomic.AddUint64(&engine.numIndexingRequests, 1)
	}
	if forceUpdate {
		atomic.AddUint64(&engine.numForceUpdatingRequests, 1)
	}
	hash := murmur.Murmur3([]byte(fmt.Sprintf("%d%s", docId, data.Content)))
	engine.segmenterChannel <- segmenterRequest{
		docId: docId, hash: hash, data: data, forceUpdate: forceUpdate}
}

// RemoveDocument remove the document from the index
// 将文档从索引中删除
//
// 输入参数：
//  docId	      标识文档编号，必须唯一，docId == 0 表示非法文档（用于强制刷新索引），[1, +oo) 表示合法文档
//  forceUpdate 是否强制刷新 cache，如果设为 true，则尽快删除索引，否则等待 cache 满之后一次全量删除
//
// 注意：
//      1. 这个函数是线程安全的，请尽可能并发调用以提高索引速度
//      2. 这个函数调用是非同步的，也就是说在函数返回时有可能文档还没有加入索引中，因此
//         如果立刻调用Search可能无法查询到这个文档。强制刷新索引请调用FlushIndex函数。
func (engine *Engine) RemoveDocument(docId uint64, forceUpdate bool) {
	if !engine.initialized {
		log.Fatal("必须先初始化引擎")
	}

	if docId != 0 {
		atomic.AddUint64(&engine.numRemovingRequests, 1)
	}
	if forceUpdate {
		atomic.AddUint64(&engine.numForceUpdatingRequests, 1)
	}
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerRemoveDocChannels[shard] <- indexerRemoveDocRequest{docId: docId, forceUpdate: forceUpdate}
		if docId == 0 {
			continue
		}
		engine.rankerRemoveDocChannels[shard] <- rankerRemoveDocRequest{docId: docId}
	}

	if engine.initOptions.UseStorage && docId != 0 {
		// 从数据库中删除
		hash := murmur.Murmur3([]byte(fmt.Sprintf("%d", docId))) % uint32(engine.initOptions.StorageShards)
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

// Search find the document that satisfies the search criteria.
// This function is thread safe
// 查找满足搜索条件的文档，此函数线程安全
func (engine *Engine) Search(request types.SearchRequest) (output types.SearchResponse) {
	if !engine.initialized {
		log.Fatal("必须先初始化引擎")
	}

	var rankOptions types.RankOptions
	if request.RankOptions == nil {
		rankOptions = *engine.initOptions.DefaultRankOptions
	} else {
		rankOptions = *request.RankOptions
	}
	if rankOptions.ScoringCriteria == nil {
		rankOptions.ScoringCriteria = engine.initOptions.DefaultRankOptions.ScoringCriteria
	}

	// 收集关键词
	tokens := []string{}
	if request.Text != "" {
		request.Text = strings.ToLower(request.Text)
		if engine.initOptions.NotUsingSegmenter {
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

	// 建立排序器返回的通信通道
	rankerReturnChannel := make(
		chan rankerReturnRequest, engine.initOptions.NumShards)

	// 生成查找请求
	lookupRequest := indexerLookupRequest{
		countDocsOnly:       request.CountDocsOnly,
		tokens:              tokens,
		labels:              request.Labels,
		docIds:              request.DocIds,
		options:             rankOptions,
		rankerReturnChannel: rankerReturnChannel,
		orderless:           request.Orderless,
		logic:               request.Logic,
	}

	// 向索引器发送查找请求
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerLookupChannels[shard] <- lookupRequest
	}

	// 从通信通道读取排序器的输出
	numDocs := 0
	rankOutput := types.ScoredDocuments{}

	//**********/ begin
	timeout := request.Timeout
	isTimeout := false
	if timeout <= 0 {
		// 不设置超时
		for shard := 0; shard < engine.initOptions.NumShards; shard++ {
			rankerOutput := <-rankerReturnChannel
			if !request.CountDocsOnly {
				for _, doc := range rankerOutput.docs {
					rankOutput = append(rankOutput, doc)
				}
			}
			numDocs += rankerOutput.numDocs
		}
	} else {
		// 设置超时
		deadline := time.Now().Add(time.Nanosecond * time.Duration(NumNanosecondsInAMillisecond*request.Timeout))
		for shard := 0; shard < engine.initOptions.NumShards; shard++ {
			select {
			case rankerOutput := <-rankerReturnChannel:
				if !request.CountDocsOnly {
					for _, doc := range rankerOutput.docs {
						rankOutput = append(rankOutput, doc)
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
		if rankOptions.ReverseOrder {
			sort.Sort(sort.Reverse(rankOutput))
		} else {
			sort.Sort(rankOutput)
		}
	}

	// 准备输出
	output.Tokens = tokens
	// 仅当CountDocsOnly为false时才充填output.Docs
	if !request.CountDocsOnly {
		if request.Orderless {
			// 无序状态无需对Offset截断
			output.Docs = rankOutput
		} else {
			var start, end int
			if rankOptions.MaxOutputs == 0 {
				start = utils.MinInt(rankOptions.OutputOffset, len(rankOutput))
				end = len(rankOutput)
			} else {
				start = utils.MinInt(rankOptions.OutputOffset, len(rankOutput))
				end = utils.MinInt(start+rankOptions.MaxOutputs, len(rankOutput))
			}
			output.Docs = rankOutput[start:end]
		}
	}

	output.NumDocs = numDocs
	output.Timeout = isTimeout

	return
}

// FlushIndex block wait until all indexes are added
// 阻塞等待直到所有索引添加完毕
func (engine *Engine) FlushIndex() {
	for {
		runtime.Gosched()
		if engine.numIndexingRequests == engine.numDocumentsIndexed &&
			engine.numRemovingRequests*uint64(engine.initOptions.NumShards) == engine.numDocumentsRemoved &&
			(!engine.initOptions.UseStorage || engine.numIndexingRequests == engine.numDocumentsStored) {
			// 保证 CHANNEL 中 REQUESTS 全部被执行完
			break
		}
	}
	// 强制更新，保证其为最后的请求
	engine.IndexDocument(0, types.DocIndexData{}, true)
	for {
		runtime.Gosched()
		if engine.numForceUpdatingRequests*uint64(engine.initOptions.NumShards) == engine.numDocumentsForceUpdated {
			return
		}
	}
}

// Close close the engine
// 关闭引擎
func (engine *Engine) Close() {
	engine.FlushIndex()
	if engine.initOptions.UseStorage {
		for _, db := range engine.dbs {
			db.Close()
		}
	}
}

// 从文本hash得到要分配到的shard
func (engine *Engine) getShard(hash uint32) int {
	return int(hash - hash/uint32(engine.initOptions.NumShards)*uint32(engine.initOptions.NumShards))
}

// GetAllDocIds get all the DocId from the storage database and return
// 从数据库遍历所有的 DocId, 并返回
func (engine *Engine) GetAllDocIds() []uint64 {
	docsId := make([]uint64, 0)
	for i := range engine.dbs {
		engine.dbs[i].ForEach(func(k, v []byte) error {
			// fmt.Println(v)
			docsId = append(docsId, uint64(k[0]))
			return nil
		})
	}
	return docsId
}
