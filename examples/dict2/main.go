package main

import (
	"log"

	"github.com/go-ego/gse"
	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

var (
	searcher  = riot.Engine{}
	searcher1 = riot.Engine{}
	searcher2 = riot.Engine{}
)

func main() {
	searcher2.Init(types.EngineOpts{
		Using: 1,
	})
	defer searcher2.Close()
	log.Println("searcher2------------------...")

	gseSegmenter := gse.Segmenter{}
	gseSegmenter.LoadDict("zh")

	searcher.WithGse(gseSegmenter).Init(
		types.EngineOpts{
			Using: 1,
		})

	log.Println("searcher----------------...")

	searcher1.WithGse(gseSegmenter).Init(
		types.EngineOpts{
			Using: 1,
		})

	data := types.DocData{Content: `I wonder how, I wonder why
		, I wonder where they are`}
	data1 := types.DocData{Content: "所以, 你好, 再见"}
	data2 := types.DocData{Content: "没有理由"}

	searcher.Index(1, data)
	searcher.Index(2, data1)
	searcher.Index(4, data1)
	searcher.Index(3, data2)
	searcher.Flush()

	req := types.SearchReq{Text: "你好"}
	search := searcher.Search(req)
	log.Println("search...", search)

	searcher1.Index(1, data)
	searcher1.Index(2, data1)
	searcher1.Index(4, data1)
	searcher1.Index(3, data2)
	searcher1.Flush()

	search1 := searcher1.Search(req)
	log.Println("search1...", search1)
}
