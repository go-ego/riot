package main

import (
	"log"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

var (
	// searcher 是协程安全的
	searcher = riot.Engine{}

	text  = "《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄"
	text1 = "在IMAX影院放映时"
	text2 = "全片以上下扩展至IMAX 1.9：1的宽高比来呈现"
)

func dictZh() {
	// 初始化
	searcher.Init(types.EngineOpts{
		// Using:         3,
		GseDict: "zh",
		// GseDict: "your gopath"+"/src/github.com/go-ego/riot/data/dict/dictionary.txt",
	})
	defer searcher.Close()

	// 将文档加入索引，docId 从1开始
	searcher.Index(1, types.DocData{Content: text})
	searcher.Index(2, types.DocData{Content: text1})
	searcher.Index(3, types.DocData{Content: text2})

	// 等待索引刷新完毕
	searcher.Flush()

	// 搜索输出格式见types.SearchResp结构体
	log.Print(searcher.Search(types.SearchReq{Text: "复仇者"}))
}

// TODO
func dictJp() {
	var searcher2 = riot.Engine{}
	// 初始化
	searcher2.Init(types.EngineOpts{
		// Using:         3,
		GseDict: "jp",
		// GseDict: "your gopath"+"/src/github.com/go-ego/riot/data/dict/jp/dict.txt",
	})
	defer searcher2.Close()

	text3 := "こんにちは世界"

	// 将文档加入索引，docId 从1开始
	searcher2.Index(1, types.DocData{Content: text})
	searcher2.Index(2, types.DocData{Content: text1})
	searcher2.Index(3, types.DocData{Content: text3})

	// 等待索引刷新完毕
	searcher2.Flush()

	// 搜索输出格式见 types.SearchResp 结构体
	log.Print(searcher2.Search(types.SearchReq{Text: "こんにちは世界"}))
}

func main() {
	dictZh()
	dictJp()
}
