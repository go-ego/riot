// 一个使用自定义评分规则搜索微博数据的例子
//
// 微博数据文件每行的格式是"<id>||||<timestamp>||||<uid>||||<reposts count>||||<text>"
// <timestamp>, <reposts count>和<text>的文本长度做评分数据
//
// 自定义评分规则为：
//	1. 首先排除关键词紧邻距离大于150个字节(五十个汉字)的微博
// 	2. 按照帖子距当前时间评分，精度为天，越晚的帖子评分越高
// 	3. 按照帖子BM25的整数部分排名
//	4. 同一天的微博再按照转发数评分，转发越多的帖子评分越高
//	5. 最后按照帖子长度评分，越长的帖子评分越高

package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-ego/gwk/engine"
	"github.com/go-ego/gwk/types"
)

const (
	SecondsInADay     = 86400
	MaxTokenProximity = 150
)

var (
	weibo_data = flag.String(
		"weibo_data",
		"../testdata/weibo_data.txt",
		"索引的微博帖子，每行当作一个文档")
	query = flag.String(
		"query",
		"chinajoy游戏",
		"待搜索的短语")
	dictionaries = flag.String(
		"dictionaries",
		"../data/dictionary.txt",
		"分词字典文件")
	stop_token_file = flag.String(
		"stop_token_file",
		"../data/stop_tokens.txt",
		"停用词文件")

	searcher = engine.Engine{}
	options  = types.RankOptions{
		ScoringCriteria: WeiboScoringCriteria{},
		OutputOffset:    0,
		MaxOutputs:      100,
	}
	searchQueries = []string{}
)

// 微博评分字段
type WeiboScoringFields struct {
	// 帖子的时间戳
	Timestamp uint32

	// 帖子的转发数
	RepostsCount uint32

	// 帖子的长度
	TextLength int
}

// 自定义的微博评分规则
type WeiboScoringCriteria struct {
}

func (criteria WeiboScoringCriteria) Score(
	doc types.IndexedDocument, fields interface{}) []float32 {
	if doc.TokenProximity > MaxTokenProximity { // 评分第一步
		return []float32{}
	}
	if reflect.TypeOf(fields) != reflect.TypeOf(WeiboScoringFields{}) {
		return []float32{}
	}
	output := make([]float32, 4)
	wsf := fields.(WeiboScoringFields)
	output[0] = float32(wsf.Timestamp / SecondsInADay) // 评分第二步
	output[1] = float32(int(doc.BM25))                 // 评分第三步
	output[2] = float32(wsf.RepostsCount)              // 评分第四步
	output[3] = float32(wsf.TextLength)                // 评分第五步
	return output
}

func main() {
	// 解析命令行参数
	flag.Parse()
	log.Printf("待搜索的短语为\"%s\"", *query)

	// 初始化
	gob.Register(WeiboScoringFields{})
	searcher.Init(types.EngineInitOptions{
		SegmenterDictionaries: *dictionaries,
		StopTokenFile:         *stop_token_file,
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
		DefaultRankOptions: &options,
	})
	defer searcher.Close()

	// 读入微博数据
	file, err := os.Open(*weibo_data)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.Printf("读入文本 %s", *weibo_data)
	scanner := bufio.NewScanner(file)
	lines := []string{}
	fieldsSlice := []WeiboScoringFields{}
	for scanner.Scan() {
		data := strings.Split(scanner.Text(), "||||")
		if len(data) != 10 {
			continue
		}
		timestamp, _ := strconv.ParseUint(data[1], 10, 32)
		repostsCount, _ := strconv.ParseUint(data[4], 10, 32)
		text := data[9]
		if text != "" {
			lines = append(lines, text)
			fields := WeiboScoringFields{
				Timestamp:    uint32(timestamp),
				RepostsCount: uint32(repostsCount),
				TextLength:   len(text),
			}
			fieldsSlice = append(fieldsSlice, fields)
		}
	}
	log.Printf("读入%d条微博\n", len(lines))

	// 建立索引
	log.Print("建立索引")
	for i, text := range lines {
		searcher.IndexDocument(uint64(i),
			types.DocIndexData{Content: text, Fields: fieldsSlice[i]})
	}
	searcher.FlushIndex()
	log.Print("索引建立完毕")

	// 搜索
	log.Printf("开始查询")
	output := searcher.Search(types.SearchRequest{Text: *query})

	// 显示
	fmt.Println()
	for _, doc := range output.Docs {
		fmt.Printf("%v %s\n\n", doc.Scores, lines[doc.DocId])
	}
	log.Printf("查询完毕")
}
