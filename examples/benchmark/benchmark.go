// 悟空性能测试
package main

import (
	"bufio"
	"flag"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/go-ego/gwk/engine"
	"github.com/go-ego/gwk/types"
)

const (
	numRepeatQuery = 1000
)

var (
	weibo_data = flag.String(
		"weibo_data",
		"../testdata/weibo_data.txt",
		"微博数据")
	queries = flag.String(
		"queries",
		"女人母亲,你好中国,网络草根,热门微博,红十字会,"+
			"鳄鱼表演,星座歧视,chinajoy,高帅富,假期计划",
		"待搜索的关键词")
	dictionaries = flag.String(
		"dictionaries",
		"../data/dictionary.txt",
		"分词字典文件")
	stop_token_file = flag.String(
		"stop_token_file",
		"../data/stop_tokens.txt",
		"停用词文件")
	cpuprofile                = flag.String("cpuprofile", "", "处理器profile文件")
	memprofile                = flag.String("memprofile", "", "内存profile文件")
	num_repeat_text           = flag.Int("num_repeat_text", 10, "文本重复加入多少次")
	num_delete_docs           = flag.Int("num_delete_docs", 1000, "测试删除文档的个数")
	index_type                = flag.Int("index_type", types.DocIdsIndex, "索引类型")
	use_persistent            = flag.Bool("use_persistent", false, "是否使用持久存储")
	persistent_storage_folder = flag.String("persistent_storage_folder", "benchmark.persistent", "持久存储数据库保存的目录")
	persistent_storage_shards = flag.Int("persistent_storage_shards", 0, "持久数据库存储裂分数目")

	searcher = engine.Engine{}
	options  = types.RankOptions{
		OutputOffset: 0,
		MaxOutputs:   100,
	}
	searchQueries = []string{}

	NumShards       = 2
	numQueryThreads = runtime.NumCPU() / NumShards
)

func main() {
	// 解析命令行参数
	flag.Parse()
	searchQueries = strings.Split(*queries, ",")
	log.Printf("待搜索的关键词为\"%s\"", searchQueries)

	// 初始化
	tBeginInit := time.Now()
	searcher.Init(types.EngineInitOptions{
		SegmenterDictionaries: *dictionaries,
		StopTokenFile:         *stop_token_file,
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: *index_type,
		},
		NumShards:               NumShards,
		DefaultRankOptions:      &options,
		UsePersistentStorage:    *use_persistent,
		PersistentStorageFolder: *persistent_storage_folder,
		PersistentStorageShards: *persistent_storage_shards,
	})
	tEndInit := time.Now()
	defer searcher.Close()

	// 打开将要搜索的文件
	file, err := os.Open(*weibo_data)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 逐行读入
	log.Printf("读入文本 %s", *weibo_data)
	scanner := bufio.NewScanner(file)
	lines := []string{}
	size := 0
	for scanner.Scan() {
		var text string
		data := strings.Split(scanner.Text(), "||||")
		if len(data) != 10 {
			continue
		}
		text = data[9]
		if text != "" {
			size += len(text) * (*num_repeat_text)
			lines = append(lines, text)
		}
	}
	log.Print("文件行数", len(lines))

	// 记录时间
	t0 := time.Now()

	// 打开处理器profile文件
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// 建索引
	log.Print("建索引 ... ")
	// 打乱 docId 顺序进行测试，若 docId 最大值超 Int 则不能用 rand.Perm 方法
	docIds := rand.Perm(*num_repeat_text * len(lines))
	docIdx := 0
	for i := 0; i < *num_repeat_text; i++ {
		for _, line := range lines {
			searcher.IndexDocument(uint64(docIds[docIdx]+1), types.DocIndexData{
				Content: line}, false)
			docIdx++
			if docIdx-docIdx/1000000*1000000 == 0 {
				log.Printf("已索引%d百万文档", docIdx/1000000)
				runtime.GC()
			}
		}
	}
	searcher.FlushIndex()
	log.Print("加入的索引总数", searcher.NumTokenIndexAdded())

	// 记录时间
	t1 := time.Now()
	log.Printf("建立索引花费时间 %v", t1.Sub(t0))
	log.Printf("建立索引速度每秒添加 %f 百万个索引",
		float64(searcher.NumTokenIndexAdded())/t1.Sub(t0).Seconds()/(1000000))

	// 记录时间并计算删除索引时间
	t2 := time.Now()
	for i := 1; i <= *num_delete_docs; i++ {
		searcher.RemoveDocument(uint64(i), false)
	}
	searcher.FlushIndex()

	t3 := time.Now()
	log.Printf("删除 %d 条索引花费时间 %v", *num_delete_docs, t3.Sub(t2))

	// 手动做 GC 防止影响性能测试
	time.Sleep(time.Second)
	runtime.GC()

	// 写入内存profile文件
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		defer f.Close()
	}

	t4 := time.Now()
	done := make(chan bool)
	recordResponse := recordResponseLock{}
	recordResponse.count = make(map[string]int)
	for iThread := 0; iThread < numQueryThreads; iThread++ {
		go search(done, &recordResponse)
	}
	for iThread := 0; iThread < numQueryThreads; iThread++ {
		<-done
	}

	// 记录时间并计算分词速度
	t5 := time.Now()
	log.Printf("搜索平均响应时间 %v 毫秒",
		t5.Sub(t4).Seconds()*1000/float64(numRepeatQuery*len(searchQueries)))
	log.Printf("搜索吞吐量每秒 %v 次查询",
		float64(numRepeatQuery*numQueryThreads*len(searchQueries))/
			t5.Sub(t4).Seconds())

	// 测试搜索结果输出，因为不同 case 的 docId 对应不上，所以只测试总数
	recordResponse.RLock()
	for keyword, count := range recordResponse.count {
		log.Printf("关键词 [%s] 共搜索到 %d 个相关文档", keyword, count)
	}
	recordResponse.RUnlock()

	if *use_persistent {
		searcher.Close()
		t6 := time.Now()
		searcher1 := engine.Engine{}
		searcher1.Init(types.EngineInitOptions{
			SegmenterDictionaries: *dictionaries,
			StopTokenFile:         *stop_token_file,
			IndexerInitOptions: &types.IndexerInitOptions{
				IndexType: *index_type,
			},
			NumShards:               NumShards,
			DefaultRankOptions:      &options,
			UsePersistentStorage:    *use_persistent,
			PersistentStorageFolder: *persistent_storage_folder,
			PersistentStorageShards: *persistent_storage_shards,
		})
		defer searcher1.Close()
		t7 := time.Now()
		t := t7.Sub(t6).Seconds() - tEndInit.Sub(tBeginInit).Seconds()
		log.Print("从持久存储加入的索引总数", searcher1.NumTokenIndexAdded())
		log.Printf("从持久存储建立索引花费时间 %v 秒", t)
		log.Printf("从持久存储建立索引速度每秒添加 %f 百万个索引",
			float64(searcher1.NumTokenIndexAdded())/t/(1000000))

	}
	//os.RemoveAll(*persistent_storage_folder)
}

type recordResponseLock struct {
	sync.RWMutex
	count map[string]int
}

func search(ch chan bool, record *recordResponseLock) {
	for i := 0; i < numRepeatQuery; i++ {
		for _, query := range searchQueries {
			output := searcher.Search(types.SearchRequest{Text: query})
			record.RLock()
			if _, found := record.count[query]; !found {
				record.RUnlock()
				record.Lock()
				record.count[query] = len(output.Docs)
				record.Unlock()
			} else {
				record.RUnlock()
			}
		}
	}
	ch <- true
}
