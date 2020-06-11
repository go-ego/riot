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
	"runtime"
)

var (
	// EngineOpts 的默认值
	// defaultNumGseThreads default Segmenter threads num
	defaultNumGseThreads = runtime.NumCPU()
	// defaultNumShards                 = 2
	defaultNumShards         = 8
	defaultIndexerBufLen     = runtime.NumCPU()
	defaultNumIndexerThreads = runtime.NumCPU()
	defaultRankerBufLen      = runtime.NumCPU()
	defaultNumRankerThreads  = runtime.NumCPU()
	defaultRankOpts          = RankOpts{
		ScoringCriteria: RankByBM25{},
	}
	defaultIndexerOpts = IndexerOpts{
		IndexType:      FrequenciesIndex,
		BM25Parameters: &defaultBM25Parameters,
	}
	defaultBM25Parameters = BM25Parameters{
		K1: 2.0,
		B:  0.75,
	}
	defaultStoreShards = 8
)

// EngineOpts init engine options
type EngineOpts struct {
	// 是否使用分词器
	// 默认使用，否则在启动阶段跳过 GseDict 和 StopTokenFile 设置
	// 如果你不需要在引擎内分词，可以将这个选项设为 true
	// 注意，如果你不用分词器，那么在调用 IndexDoc 时,
	// DocIndexData 中的 Content 会被忽略
	// Not use the gse segment
	NotUseGse bool `toml:"not_use_gse"`

	// new, 分词规则
	Using int `toml:"using"`

	// 半角逗号 "," 分隔的字典文件，具体用法见
	// gse.Segmenter.LoadDict 函数的注释
	GseDict   string `toml:"gse_dict"`
	PinYin    bool   `toml:"pin_yin"`
	UsePhrase bool   `toml:"use_phrase"`

	// 停用词文件
	StopTokenFile string `toml:"stop_file"`
	// Gse search mode
	GseMode bool   `toml:"gse_mode"`
	Hmm     bool   `toml:"hmm"`
	Model   string `toml:"model"`

	// 分词器线程数
	// NumSegmenterThreads int
	NumGseThreads int

	// 索引器和排序器的 shard 数目
	// 被检索/排序的文档会被均匀分配到各个 shard 中
	NumShards int

	// 索引器的信道缓冲长度
	IndexerBufLen int

	// 索引器每个shard分配的线程数
	NumIndexerThreads int

	// 排序器的信道缓冲长度
	RankerBufLen int

	// 排序器每个 shard 分配的线程数
	NumRankerThreads int

	// 索引器初始化选项
	IndexerOpts *IndexerOpts

	// 默认的搜索选项
	DefRankOpts *RankOpts

	// 是否使用持久数据库，以及数据库文件保存的目录和裂分数目
	StoreOnly bool `toml:"store_only"`
	UseStore  bool `toml:"use_store"`

	StoreFolder string `toml:"store_folder"`
	StoreShards int    `toml:"store_shards"`
	StoreEngine string `toml:"store_engine"`

	IDOnly bool `toml:"id_only"`
}

// Init init engine options
// 初始化 EngineOpts，当用户未设定某个选项的值时用默认值取代
func (options *EngineOpts) Init() {
	// if !options.NotUseGse && options.GseDict == "" {
	// 	log.Fatal("字典文件不能为空")
	//  options.GseDict = "zh"
	// }

	if options.NumGseThreads == 0 {
		options.NumGseThreads = defaultNumGseThreads
	}

	if options.NumShards == 0 {
		options.NumShards = defaultNumShards
	}

	if options.IndexerBufLen == 0 {
		options.IndexerBufLen = defaultIndexerBufLen
	}

	if options.NumIndexerThreads == 0 {
		options.NumIndexerThreads = defaultNumIndexerThreads
	}

	if options.RankerBufLen == 0 {
		options.RankerBufLen = defaultRankerBufLen
	}

	if options.NumRankerThreads == 0 {
		options.NumRankerThreads = defaultNumRankerThreads
	}

	if options.IndexerOpts == nil {
		options.IndexerOpts = &defaultIndexerOpts
	}

	if options.IndexerOpts.BM25Parameters == nil {
		options.IndexerOpts.BM25Parameters = &defaultBM25Parameters
	}

	if options.DefRankOpts == nil {
		options.DefRankOpts = &defaultRankOpts
	}

	if options.DefRankOpts.ScoringCriteria == nil {
		options.DefRankOpts.ScoringCriteria = defaultRankOpts.ScoringCriteria
	}

	if options.StoreShards == 0 {
		options.StoreShards = defaultStoreShards
	}
}
