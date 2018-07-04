// riot 性能测试
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

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

const (
	numRepeatQuery = 1000
)

var (
	weiboData = flag.String(
		"weibo_data",
		"../../testdata/weibo_data.txt",
		"微博数据")
	queries = flag.String(
		"queries",
		"女人母亲, 你好中国, 网络草根, 热门微博, 红十字会,"+
			"鳄鱼表演, 星座歧视, chinajoy, 高帅富, 假期计划",
		"待搜索的关键词")
	dictionaries = flag.String(
		"dictionaries",
		"../../data/dict/dictionary.txt",
		"分词字典文件")
	stopTokenFile = flag.String(
		"stop_token_file",
		"../../data/dict/stop_tokens.txt",
		"停用词文件")
	cpuprofile              = flag.String("cpuprofile", "", "处理器profile文件")
	memprofile              = flag.String("memprofile", "", "内存profile文件")
	numRepeatText           = flag.Int("numRepeatText", 10, "文本重复加入多少次")
	numDeleteDocs           = flag.Int("numDeleteDocs", 1000, "测试删除文档的个数")
	indexType               = flag.Int("indexType", types.DocIdsIndex, "索引类型")
	usePersistent           = flag.Bool("usePersistent", false, "是否使用持久存储")
	persistentStorageFolder = flag.String("persistentStorageFolder", "benchmark.persistent", "持久存储数据库保存的目录")
	storageEngine           = flag.String("storageEngine", "lbd", "use StorageEngine")
	persistentStorageShards = flag.Int("persistentStorageShards", 0, "持久数据库存储裂分数目")

	searcher = riot.Engine{}
	options  = types.RankOpts{
		OutputOffset: 0,
		MaxOutputs:   100,
	}
	searchQueries = []string{}

	// NumShards shards number
	NumShards       = 2
	numQueryThreads = runtime.NumCPU() / NumShards
	t0              time.Time
)

func initEngine() {
	searcher.Init(types.EngineOpts{
		GseDict:       *dictionaries,
		StopTokenFile: *stopTokenFile,
		IndexerOpts: &types.IndexerOpts{
			IndexType: *indexType,
		},
		NumShards:       NumShards,
		DefaultRankOpts: &options,
		UseStorage:      *usePersistent,
		StorageFolder:   *persistentStorageFolder,
		StorageShards:   *persistentStorageShards,
	})
}

func openFile() {
	// 打开将要搜索的文件
	file, err := os.Open(*weiboData)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 逐行读入
	log.Printf("读入文本 %s", *weiboData)
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
			size += len(text) * (*numRepeatText)
			lines = append(lines, text)
		}
	}
	log.Println("size ...", size)
	log.Print("文件行数", len(lines))

	// 记录时间
	t0 = time.Now()

	// 打开处理器 profile 文件
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
	docIds := rand.Perm(*numRepeatText * len(lines))
	docIdx := 0
	for i := 0; i < *numRepeatText; i++ {
		for _, line := range lines {
			searcher.Index(
				uint64(docIds[docIdx]+1), types.DocData{
					Content: line})
			docIdx++
			if docIdx-docIdx/1000000*1000000 == 0 {
				log.Printf("已索引%d百万文档", docIdx/1000000)
				runtime.GC()
			}
		}
	}
}

func deleteDoc() {
	// 记录时间
	t1 := time.Now()
	log.Printf("建立索引花费时间 %v", t1.Sub(t0))
	log.Printf("建立索引速度每秒添加 %f 百万个索引",
		float64(searcher.NumTokenIndexAdded())/t1.Sub(t0).Seconds()/(1000000))

	// 记录时间并计算删除索引时间
	t2 := time.Now()
	for i := 1; i <= *numDeleteDocs; i++ {
		searcher.RemoveDoc(uint64(i))
	}
	searcher.Flush()

	t3 := time.Now()
	log.Printf("删除 %d 条索引花费时间 %v", *numDeleteDocs, t3.Sub(t2))
}

func searchQu() {
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
}

func useStore(tBeginInit, tEndInit time.Time) {
	searcher.Close()
	t6 := time.Now()
	searcher1 := riot.Engine{}
	searcher1.Init(types.EngineOpts{
		GseDict:       *dictionaries,
		StopTokenFile: *stopTokenFile,
		IndexerOpts: &types.IndexerOpts{
			IndexType: *indexType,
		},
		NumShards:       NumShards,
		DefaultRankOpts: &options,
		UseStorage:      *usePersistent,
		StorageFolder:   *persistentStorageFolder,
		StorageEngine:   *storageEngine,
		StorageShards:   *persistentStorageShards,
	})
	defer searcher1.Close()
	t7 := time.Now()
	t := t7.Sub(t6).Seconds() - tEndInit.Sub(tBeginInit).Seconds()
	log.Print("从持久存储加入的索引总数", searcher1.NumTokenIndexAdded())
	log.Printf("从持久存储建立索引花费时间 %v 秒", t)
	log.Printf("从持久存储建立索引速度每秒添加 %f 百万个索引",
		float64(searcher1.NumTokenIndexAdded())/t/(1000000))
}

func main() {
	// 解析命令行参数
	flag.Parse()
	searchQueries = strings.Split(*queries, ",")
	log.Printf("待搜索的关键词为\"%s\"", searchQueries)

	// 初始化
	tBeginInit := time.Now()

	initEngine()

	tEndInit := time.Now()
	defer searcher.Close()

	openFile()

	searcher.Flush()
	log.Print("加入的索引总数", searcher.NumTokenIndexAdded())

	deleteDoc()

	// 手动做 GC 防止影响性能测试
	time.Sleep(time.Second)
	runtime.GC()

	// 写入内存 profile 文件
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		defer f.Close()
	}

	searchQu()

	if *usePersistent {
		useStore(tBeginInit, tEndInit)
	}
	//os.RemoveAll(*persistentStorageFolder)

	log.Println("end...")
}

type recordResponseLock struct {
	sync.RWMutex
	count map[string]int
}

func search(ch chan bool, record *recordResponseLock) {
	for i := 0; i < numRepeatQuery; i++ {
		for _, query := range searchQueries {
			output := searcher.Search(types.SearchReq{Text: query})
			record.RLock()
			if _, found := record.count[query]; !found {
				record.RUnlock()
				record.Lock()
				record.count[query] = len(output.Docs.(types.ScoredDocs))
				record.Unlock()
			} else {
				record.RUnlock()
			}
		}
	}
	ch <- true
}
