package main

import (
	"log"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

var (
	searcher = riot.New("zh")
)

func main() {
	data := types.DocData{Content: `I wonder how, I wonder why
		, I wonder where they are`}
	data1 := types.DocData{Content: "所以, 你好, 再见"}
	data2 := types.DocData{Content: "没有理由"}
	searcher.Index(1, data)
	searcher.Index(2, data1)
	searcher.Index(3, data2)
	searcher.Flush()

	req := types.SearchReq{Text: "你好"}
	search := searcher.Search(req)
	log.Println("search response: ", search)

	usedMem, uerr := riot.MemUsed()
	log.Println("searcher mem used: ", usedMem, uerr)

	memPet, perr := riot.MemPercent()
	log.Println("searcher mem used percent: ", memPet, perr)

	riotMem, merr := searcher.UsedMem()
	log.Println("init mem: ", riot.InitMemUsed,
		"searcher mem: ", riotMem, "To MB: ", riot.ToMB(riotMem), merr)

	diskPet, err := riot.DiskPercent()
	log.Println("searcher use disk percent: ", diskPet, err)
	usedDisk, derr := riot.DiskUsed()
	log.Println("searcher disk used: ", usedDisk, derr)
}
