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

package types

import (
	"log"
	"runtime"
)

var (
	// EngineInitOptions的默认值
	defaultNumSegmenterThreads = runtime.NumCPU()
	// defaultNumShards                 = 2
	defaultNumShards                 = 8
	defaultIndexerBufferLength       = runtime.NumCPU()
	defaultNumIndexerThreadsPerShard = runtime.NumCPU()
	defaultRankerBufferLength        = runtime.NumCPU()
	defaultNumRankerThreadsPerShard  = runtime.NumCPU()
	defaultDefaultRankOptions        = RankOptions{
		ScoringCriteria: RankByBM25{},
	}
	defaultIndexerInitOptions = IndexerInitOptions{
		IndexType:      FrequenciesIndex,
		BM25Parameters: &defaultBM25Parameters,
	}
	defaultBM25Parameters = BM25Parameters{
		K1: 2.0,
		B:  0.75,
	}
	defaultStorageShards = 8
)

// EngineInitOptions init engine options
type EngineInitOptions struct {
	// 是否使用分词器
	// 默认使用，否则在启动阶段跳过SegmenterDict和StopTokenFile设置
	// 如果你不需要在引擎内分词，可以将这个选项设为true
	// 注意，如果你不用分词器，那么在调用IndexDocument时DocIndexData中的Content会被忽略
	NotUsingSegmenter bool

	// new
	Using int

	// 半角逗号分隔的字典文件，具体用法见
	// sego.Segmenter.LoadDict函数的注释
	SegmenterDict string

	// 停用词文件
	StopTokenFile string

	// 分词器线程数
	NumSegmenterThreads int

	// 索引器和排序器的shard数目
	// 被检索/排序的文档会被均匀分配到各个shard中
	NumShards int

	// 索引器的信道缓冲长度
	IndexerBufferLength int

	// 索引器每个shard分配的线程数
	NumIndexerThreadsPerShard int

	// 排序器的信道缓冲长度
	RankerBufferLength int

	// 排序器每个shard分配的线程数
	NumRankerThreadsPerShard int

	// 索引器初始化选项
	IndexerInitOptions *IndexerInitOptions

	// 默认的搜索选项
	DefaultRankOptions *RankOptions

	// 是否使用持久数据库，以及数据库文件保存的目录和裂分数目
	UseStorage    bool
	StorageFolder string
	StorageShards int
	StorageEngine string
}

// Init 初始化 EngineInitOptions，当用户未设定某个选项的值时用默认值取代
func (options *EngineInitOptions) Init() {
	if !options.NotUsingSegmenter {
		if options.SegmenterDict == "" {
			log.Fatal("字典文件不能为空")
		}
	}

	if options.NumSegmenterThreads == 0 {
		options.NumSegmenterThreads = defaultNumSegmenterThreads
	}

	if options.NumShards == 0 {
		options.NumShards = defaultNumShards
	}

	if options.IndexerBufferLength == 0 {
		options.IndexerBufferLength = defaultIndexerBufferLength
	}

	if options.NumIndexerThreadsPerShard == 0 {
		options.NumIndexerThreadsPerShard = defaultNumIndexerThreadsPerShard
	}

	if options.RankerBufferLength == 0 {
		options.RankerBufferLength = defaultRankerBufferLength
	}

	if options.NumRankerThreadsPerShard == 0 {
		options.NumRankerThreadsPerShard = defaultNumRankerThreadsPerShard
	}

	if options.IndexerInitOptions == nil {
		options.IndexerInitOptions = &defaultIndexerInitOptions
	}

	if options.IndexerInitOptions.BM25Parameters == nil {
		options.IndexerInitOptions.BM25Parameters = &defaultBM25Parameters
	}

	if options.DefaultRankOptions == nil {
		options.DefaultRankOptions = &defaultDefaultRankOptions
	}

	if options.DefaultRankOptions.ScoringCriteria == nil {
		options.DefaultRankOptions.ScoringCriteria = defaultDefaultRankOptions.ScoringCriteria
	}

	if options.StorageShards == 0 {
		options.StorageShards = defaultStorageShards
	}
}

// Try handler(err)
func Try(fun func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			handler(err)
		}
	}()
	fun()
}
