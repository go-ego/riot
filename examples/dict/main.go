package main

import (
	"log"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

var (
	// searcher 是协程安全的
	searcher = riot.Engine{}
)

func dictZh() {
	// 初始化
	searcher.Init(types.EngineInitOptions{
		// Using:         3,
		SegmenterDict: "zh",
		// SegmenterDict: "your gopath"+"/src/github.com/go-ego/riot/data/dict/dictionary.txt",
	})
	defer searcher.Close()

	text := "此次百度收购将成中国互联网最大并购"
	text1 := "百度宣布拟全资收购91无线业务"
	text2 := "百度是中国最大的搜索引擎"

	// 将文档加入索引，docId 从1开始
	searcher.IndexDocument(1, types.DocIndexData{Content: text}, false)
	searcher.IndexDocument(2, types.DocIndexData{Content: text1}, false)
	searcher.IndexDocument(3, types.DocIndexData{Content: text2}, false)

	// 等待索引刷新完毕
	searcher.FlushIndex()

	// 搜索输出格式见types.SearchResponse结构体
	log.Print(searcher.Search(types.SearchRequest{Text: "百度中国"}))
}

// TODO
func dictJp() {
	var searcher2 = riot.Engine{}
	// 初始化
	searcher2.Init(types.EngineInitOptions{
		// Using:         3,
		SegmenterDict: "jp",
		// SegmenterDict: "your gopath"+"/src/github.com/go-ego/riot/data/dict/jp/dict.txt",
	})
	defer searcher2.Close()

	text := "此次百度收购将成中国互联网最大并购"
	text1 := "百度宣布拟全资收购91无线业务"
	text2 := "こんにちは世界"

	// 将文档加入索引，docId 从1开始
	searcher2.IndexDocument(1, types.DocIndexData{Content: text}, false)
	searcher2.IndexDocument(2, types.DocIndexData{Content: text1}, false)
	searcher2.IndexDocument(3, types.DocIndexData{Content: text2}, false)

	// 等待索引刷新完毕
	searcher2.FlushIndex()

	// 搜索输出格式见types.SearchResponse结构体
	log.Print(searcher2.Search(types.SearchRequest{Text: "こんにちは世界"}))
}

func main() {
	dictZh()
	dictJp()
}
