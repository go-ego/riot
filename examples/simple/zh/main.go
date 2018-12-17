/*

没有比这个更简单的例子了。

*/

package main

import (
	"log"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

var (
	// searcher 是线程安全的
	searcher = riot.Engine{}
)

func main() {
	// 初始化
	searcher.Init(types.EngineOpts{
		Using: 3,
		// IDOnly:        true,
		GseDict: "../../../data/dict/dictionary.txt",
	})
	defer searcher.Close()

	text := "《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄"
	text1 := "在IMAX影院放映时"
	text2 := "全片以上下扩展至IMAX 1.9：1的宽高比来呈现"

	// 将文档加入索引，docId 从1开始
	searcher.Index("1", types.DocData{Content: text})
	searcher.Index("2", types.DocData{Content: text1})
	searcher.Index("3", types.DocData{Content: text2})

	// 等待索引刷新完毕
	searcher.Flush()

	// 搜索输出格式见 types.SearchResp 结构体
	log.Print(searcher.Search(types.SearchReq{Text: "复仇者"}))
}
