// 一个微博搜索的例子。
package main

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-ego/riot/engine"
	"github.com/go-ego/riot/types"
)

const (
	SecondsInADay     = 86400
	MaxTokenProximity = 2
)

var (
	searcher      = engine.Engine{}
	wbs           = map[uint64]Weibo{}
	weiboData     = flag.String("weibo_data", "../../testdata/weibo_data.txt", "微博数据文件")
	dictFile      = flag.String("dict_file", "../../data/dictionary.txt", "词典文件")
	stopTokenFile = flag.String("stop_token_file", "../../data/stop_tokens.txt", "停用词文件")
	staticFolder  = flag.String("static_folder", "static", "静态文件目录")
)

type Weibo struct {
	Id           uint64 `json:"id"`
	Timestamp    uint64 `json:"timestamp"`
	UserName     string `json:"user_name"`
	RepostsCount uint64 `json:"reposts_count"`
	Text         string `json:"text"`
}

/*******************************************************************************
    索引
*******************************************************************************/
func indexWeibo() {
	// 读入微博数据
	file, err := os.Open(*weiboData)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := strings.Split(scanner.Text(), "||||")
		if len(data) != 10 {
			continue
		}
		wb := Weibo{}
		wb.Id, _ = strconv.ParseUint(data[0], 10, 64)
		wb.Timestamp, _ = strconv.ParseUint(data[1], 10, 64)
		wb.UserName = data[3]
		wb.RepostsCount, _ = strconv.ParseUint(data[4], 10, 64)
		wb.Text = data[9]
		wbs[wb.Id] = wb
	}

	log.Print("添加索引")
	for docId, weibo := range wbs {
		searcher.IndexDocument(docId, types.DocIndexData{
			Content: weibo.Text,
			Fields: WeiboScoringFields{
				Timestamp:    weibo.Timestamp,
				RepostsCount: weibo.RepostsCount,
			},
		}, false)
	}

	searcher.FlushIndex()
	log.Printf("索引了%d条微博\n", len(wbs))
}

/*******************************************************************************
    评分
*******************************************************************************/
type WeiboScoringFields struct {
	Timestamp    uint64
	RepostsCount uint64
}

type WeiboScoringCriteria struct {
}

func (criteria WeiboScoringCriteria) Score(
	doc types.IndexedDocument, fields interface{}) []float32 {
	if reflect.TypeOf(fields) != reflect.TypeOf(WeiboScoringFields{}) {
		return []float32{}
	}
	wsf := fields.(WeiboScoringFields)
	output := make([]float32, 3)
	if doc.TokenProximity > MaxTokenProximity {
		output[0] = 1.0 / float32(doc.TokenProximity)
	} else {
		output[0] = 1.0
	}
	output[1] = float32(wsf.Timestamp / (SecondsInADay * 3))
	output[2] = float32(doc.BM25 * (1 + float32(wsf.RepostsCount)/10000))
	return output
}

/*******************************************************************************
    JSON-RPC
*******************************************************************************/
type JsonResponse struct {
	Docs []*Weibo `json:"docs"`
}

func JsonRpcServer(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query().Get("query")
	output := searcher.Search(types.SearchRequest{
		Text: query,
		RankOptions: &types.RankOptions{
			ScoringCriteria: &WeiboScoringCriteria{},
			OutputOffset:    0,
			MaxOutputs:      100,
		},
	})

	// 整理为输出格式
	docs := []*Weibo{}
	for _, doc := range output.Docs {
		wb := wbs[doc.DocId]
		for _, t := range output.Tokens {
			wb.Text = strings.Replace(wb.Text, t, "<font color=red>"+t+"</font>", -1)
		}
		docs = append(docs, &wb)
	}
	response, _ := json.Marshal(&JsonResponse{Docs: docs})

	// fmt.Println("response...", response)

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(response))
}

/*******************************************************************************
	主函数
*******************************************************************************/
func main() {
	// 解析命令行参数
	flag.Parse()

	// 初始化
	gob.Register(WeiboScoringFields{})
	log.Print("引擎开始初始化")
	searcher.Init(types.EngineInitOptions{
		SegmenterDictionaries: *dictFile,
		StopTokenFile:         *stopTokenFile,
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
		// 如果你希望使用持久存储，启用下面的选项
		// 默认使用boltdb持久化，如果你希望修改数据库类型
		// 请修改 GWK_STORAGE_ENGINE 环境变量
		// UseStorage: true,
		// PersistentStorageFolder: "weibo_search",
	})
	log.Print("引擎初始化完毕")
	wbs = make(map[uint64]Weibo)

	// 索引
	log.Print("建索引开始")
	go indexWeibo()
	log.Print("建索引完毕")

	// 捕获ctrl-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			log.Print("捕获Ctrl-c，退出服务器")
			searcher.Close()
			os.Exit(0)
		}
	}()

	http.HandleFunc("/json", JsonRpcServer)
	http.Handle("/", http.FileServer(http.Dir(*staticFolder)))
	log.Print("服务器启动")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
