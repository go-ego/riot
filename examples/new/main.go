package main

import (
	"log"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

var (
	searcher = riot.New("../../data/dict/dictionary.txt")
)

func main() {
	data := types.DocIndexData{Content: `I wonder how, I wonder why
		, I wonder where they are`}
	data1 := types.DocIndexData{Content: "留给真爱你的人"}
	data2 := types.DocIndexData{Content: "也没有理由"}
	searcher.IndexDoc(1, data)
	searcher.IndexDoc(2, data1)
	searcher.IndexDoc(3, data2)
	searcher.FlushIndex()

	req := types.SearchReq{Text: "真爱"}
	search := searcher.Search(req)
	log.Println("search...", search)
}
