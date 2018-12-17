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

	data = types.DocData{Content: `I wonder how, I wonder why
		, I wonder where they are`}
	data1 = types.DocData{Content: "所以, 你好, 再见"}
	data2 = types.DocData{Content: "没有理由"}

	req = types.SearchReq{Text: "你好"}
)

func searchWithGseFn1(seg gse.Segmenter) {
	searcher.WithGse(seg).Init(
		types.EngineOpts{
			Using: 1,
		})

	log.Println("searcher----------------...")

	searcher.Index("1", data)
	searcher.Index("2", data1)
	searcher.Index("4", data1)
	searcher.Index("3", data2)
	searcher.Flush()

	search := searcher.Search(req)
	log.Println("search...", search)
}

func searchWithGse2(seg gse.Segmenter) {
	searcher1.WithGse(seg).Init(
		types.EngineOpts{
			Using: 1,
		})

	searcher1.Index("1", data)
	searcher1.Index("2", data1)
	searcher1.Index("4", data1)
	searcher1.Index("3", data2)
	searcher1.Flush()

	search1 := searcher1.Search(req)
	log.Println("search1...", search1)
}

func main() {
	searcher2.Init(types.EngineOpts{
		Using: 1,
	})
	defer searcher2.Close()
	log.Println("searcher2------------------...")

	seg := gse.Segmenter{}
	seg.LoadDict("zh")

	searchWithGseFn1(seg)

	searchWithGse2(seg)
}
