package main

import (
	"log"
	"os"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

var (
	searcher = riot.New("../../data/dict/dictionary.txt")
	engine   = riot.New("../../testdata/test_new.toml")
)

func main() {
	data := types.DocData{Content: `I wonder how, I wonder why
		, I wonder where they are`}
	data1 := types.DocData{Content: "所以, 你好, 再见"}
	data2 := types.DocData{Content: "没有理由"}

	searcher.Index("1", data)
	searcher.Index("2", data1)
	searcher.Index("3", data2)
	searcher.Flush()

	engine.Index("1", data)
	engine.Index("2", data1)
	engine.Flush()

	req := types.SearchReq{Text: "你好"}
	search := searcher.Search(req)
	log.Println("search response: ", search)

	req = types.SearchReq{Text: "how"}
	search = engine.Search(req)
	log.Println("search response: ", search)

	searcher.Close()
	engine.Close()
	os.RemoveAll("./riot.new")
}
