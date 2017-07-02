悟空引擎入门
====

在本篇的结束，你将学会用悟空引擎写一个简单的全文本微博搜索。

在阅读本篇之前，你需要对Go语言有基本了解，如果你还不会用Go，[这里](http://go-tour-zh.appspot.com/#1)有个教程。

## 引擎的原理

引擎中处理用户请求、分词、索引和排序分别由不同的协程（goroutines）完成。

	1. 主协程，用于收发用户请求
	2. 分词器（segmenter）协程，负责分词
	3. 索引器（indexer）协程，负责建立和查找索引表
	4. 排序器（ranker）协程，负责对文档评分排序
    
![](https://raw.github.com/go-ego/gwk/master/docs/gwk.png)

**索引流程**

当一个将文档（document）加入索引的请求进来以后，主协程会通过一个信道（channel）将要分词的文本发送给某个分词协程，该协程将文本分词后通过另一个信道发送给一个索引器协程。索引器协程建立从搜索键（search keyword）到文档的反向索引（inverted index），反向索引表保存在内存中方便快速调用。

**搜索流程**

主协程接到用户请求，将请求短语在主协程内分词，然后通过信道发送给索引器，索引器查找每个搜索键对应的文档然后进行逻辑操作（归并求交集）得到一个精简的文档列表，此列表通过信道传递给排序器，排序器对文档进行评分（scoring）、筛选和排序，然后将排好序的文档通过指定的信道发送给主协程，主协程将结果返回给用户。

分词、索引和排序都有多个协程完成，中间结果保存在信道缓冲队列以避免阻塞。为了提高搜索的并发度降低延迟，悟空引擎将文档做了裂分（裂分数目可以由用户指定），索引和排序请求会发送到所有裂分上并行处理，结果在主协程进行二次归并排序。

上面就是悟空引擎的大致原理。任何完整的搜索系统都包括四个部分， **文档抓取** 、 **索引** 、 **搜索** 和 **显示** 。下面分别讲解这些部分是怎么实现的。

## 文档抓取

文档抓取的技术很多，多到可以单独拿出来写一篇文章。幸运的是微博抓取相对简单，可以通过新浪提供的API实现的，而且已经有[Go语言的SDK](http://github.com/huichen/gobo)可以并发抓取并且速度相当快。

我已经抓了大概十万篇微博放在了testdata/weibo_data.txt里(因为影响git clone的下载速度所以删除了，请从[这里](https://github.com/go-ego/gwk/blob/43f20b4c0921cc704cf41fe8653e66a3fcbb7e31/testdata/weibo_data.txt?raw=true)下载)，所以你就不需要自己做了。文件中每行存储了一篇微博，格式如下

    <微博id>||||<时间戳>||||<用户id>||||<用户名>||||<转贴数>||||<评论数>||||<喜欢数>||||<小图片网址>||||<大图片网址>||||<正文>

微博保存在下面的结构体中方便查询，只载入了我们需要的数据：

```go
type Weibo struct {
        Id           uint64
        Timestamp    uint64
        UserName     string
        RepostsCount uint64
        Text         string
}
```

如果你对抓取的细节感兴趣请见抓取程序[testdata/crawl_weibo_data.go](/testdata/crawl_weibo_data.go)。

## 索引

使用悟空引擎你需要import两个包

```go
import (
	"github.com/go-ego/gwk/engine"
	"github.com/go-ego/gwk/types"
)
```
第一个包定义了引擎功能，第二个包定义了常用结构体。在使用引擎之前需要初始化，例如

```go
var searcher engine.Engine
searcher.Init(types.EngineInitOptions{
	SegmenterDictionaries: "../../data/dictionary.txt",
	StopTokenFile:         "../../data/stop_tokens.txt",
	IndexerInitOptions: &types.IndexerInitOptions{
		IndexType: types.LocationsIndex,
	},
})
```
[types.EngineInitOptions](/types/engine_init_options.go)定义了初始化引擎需要设定的参数，比如从何处载入分词字典文件，停用词列表，索引器类型，BM25参数等，以及默认的评分规则（见“搜索”一节）和输出分页选项。具体细节请阅读代码中结构体的注释。

特别需要强调的是请慎重选择IndexerInitOptions.IndexType的类型，共有三种不同类型的索引表：

1. DocIdsIndex，提供了最基本的索引，仅仅记录搜索键出现的文档docid。
2. FrequenciesIndex，除了记录docid外，还保存了搜索键在每个文档中出现的频率，如果你需要BM25那么FrequenciesIndex是你需要的。
3. LocationsIndex，这个不仅包括上两种索引的内容，还额外存储了关键词在文档中的具体位置，这用来[计算紧邻距离](/docs/token_proximity.md)。

这三种索引由上到下在提供更多计算能力的同时也消耗了更多的内存，特别是LocationsIndex，当文档很长时会占用大量内存。请根据需要平衡选择。

初始化好了以后就可以添加索引了，下面的例子将一条微博加入引擎

```go
searcher.IndexDocument(docId, types.DocIndexData{
	Content: weibo.Text, // Weibo结构体见上文的定义。必须是UTF-8格式。
	Fields: WeiboScoringFields{
		Timestamp:    weibo.Timestamp,
		RepostsCount: weibo.RepostsCount,
	},
})
```

文档的docId必须大于0且唯一，对微博来说可以直接用微博的ID。悟空引擎允许你加入三种索引数据：

1. 文档的正文（content），会被分词为关键词（tokens）加入索引。
2. 文档的关键词（tokens）。当正文为空的时候，允许用户绕过悟空内置的分词器直接输入文档关键词，这使得在引擎外部进行文档分词成为可能。
3. 文档的属性标签（labels），比如微博的作者，类别等。标签并不出现在正文中。
4. 自定义评分字段（scoring fields），这允许你给文档添加 **任意类型** 、 **任意结构** 的数据用于排序。“搜索”一节会进一步介绍自定义评分字段的用法。

**特别注意的是** ，关键词（tokens）和标签（labels）组成了索引器中的搜索键（keywords），文档和代码中会反复出现这三个概念，请不要混淆。对正文的搜索就是在搜索键上的逻辑查询，比如一个文档正文中出现了“自行车”这个关键词，也有“健身”这样的分类标签，但“健身”这个词并不直接出现在正文中，当查询“自行车”+“健身”这样的搜索键组合时，这篇文章就会被查询到。设计标签的目的是为了方便从非字面意义的维度快速缩小查询范围。

引擎采用了非同步的索引方式，也就是说当IndexDocument返回时索引可能还没有加入索引表中，这方便你循环并发地加入索引。如果你需要等待索引添加完毕后再进行后续操作，请调用下面的函数

```go
searcher.FlushIndex()
```

## 搜索

搜索的过程分两步，第一步是在索引表中查找包含搜索键的文档，这在上一节已经介绍过。第二步是对所有索引到的文档进行排序。

排序的核心是对文档评分。悟空引擎允许你自定义任意的评分规则（scoring criteria）。在微博搜索例子中，我们定义的评分规则如下：

1. 首先按照关键词紧邻距离排序，比如搜索“自行车运动”，这个短语会被切分成两个关键词，“自行车”和“运动”，出现两个关键词紧邻的文章应该排在两个关键词分开的文章前面。
2. 然后按照微博的发布时间大致排序，每三天为一个梯队，较晚梯队的文章排在前面。
3. 最后给微博打分为 BM25*(1+转发数/10000)

这样的规则需要给每个文档保存一些评分数据，比如微博发布时间，微博的转发数等。这些数据保存在下面的结构体中
```go
type WeiboScoringFields struct {
        Timestamp    uint64
        RepostsCount uint64
}
```
你可能已经注意到了，这就是在上一节将文档加入索引时调用的IndexDocument函数传入的参数类型（实际上那个参数是interface{}类型的，因此可以传入任意类型的结构体）。这些数据保存在排序器的内存中等待调用。

有了这些数据，我们就可以评分了，代码如下：
```go
type WeiboScoringCriteria struct {
}

func (criteria WeiboScoringCriteria) Score {
        doc types.IndexedDocument, fields interface{}) []float32 {
        if reflect.TypeOf(fields) != reflect.TypeOf(WeiboScoringFields{}) {
                return []float32{}
        }
        wsf := fields.(WeiboScoringFields)
        output := make([]float32, 3)
        if doc.TokenProximity > MaxTokenProximity { // 第一步
                output[0] = 1.0 / float32(doc.TokenProximity)
        } else {
                output[0] = 1.0
        }
        output[1] = float32(wsf.Timestamp / (SecondsInADay * 3)) // 第二步
        output[2] = float32(doc.BM25 * (1 + float32(wsf.RepostsCount)/10000)) // 第三步
        return output
}
```
WeiboScoringCriteria实际上继承了types.ScoringCriteria接口，这个接口实现了Score函数。这个函数带有两个参数：

1. types.IndexedDocument参数传递了从索引器中得到的数据，比如词频，词的具体位置，BM25值，紧邻度等信息，具体见[types/index.go](/types/index.go)的注释。
2. 第二个参数是interface{}类型的，你可以把这个类型理解成C语言中的void指针，它可以指向任何数据类型。在我们的例子中指向的是WeiboScoringFields结构体，并通过反射机制检查是否是正确的类型。

有了自定义评分数据和自定义评分规则，我们就可以进行搜索了，见下面的代码

```go
response := searcher.Search(types.SearchRequest{
	Text: "自行车运动",
	RankOptions: &types.RankOptions{
		ScoringCriteria: &WeiboScoringCriteria{},
		OutputOffset: 0,
		MaxOutputs:   100,
	},
})
```

其中，Text是输入的搜索短语（必须是UTF-8格式），会被分词为关键词。和索引时相同，悟空引擎允许绕过内置的分词器直接输入关键词和文档标签，见types.SearchRequest结构体的注释。RankOptions定义了排序选项。WeiboScoringCriteria就是我们在上面定义的评分规则。另外你也可以通过OutputOffset和MaxOutputs参数控制分页输出。搜索结果保存在response变量中，具体内容见[types/search_response.go](/types/search_response.go)文件中定义的SearchResponse结构体，比如这个结构体返回了关键词出现在文档中的位置，可以用来生成文档的摘要。

## 显示

完成用户搜索的最后一步是将搜索结果呈现给用户。 通常的做法是将搜索引擎做成一个后台服务，然后让前端以JSON-RPC的方式调用它。前端并不属于悟空引擎本身因此就不多着墨了。

## 总结

读到这里，你应该对使用悟空引擎进行微博搜索有了基本了解，建议你自己动手将其完成。如果你没有耐心，可以看看已经完成的代码，见[examples/codelab/search_server.go](/examples/codelab/search_server.go)，总共不到200行。运行这个例子非常简单，进入examples/codelab目录后输入

	go run search_server.go

等待终端中出现“索引了xxx条微博”的输出后，在浏览器中打开[http://localhost:8080](http://localhost:8080) 即可进入搜索页面。

如果你想进一步了解悟空引擎，建议你直接阅读代码。代码的目录结构如下：

    /core       核心部件，包括索引器和排序器
    /data       字典文件和停用词文件
    /docs       文档
    /engine     引擎，包括主协程、分词协程、索引器协程和排序器协程的实现
    /examples   例子和性能测试程序
    /testdata   测试数据
    /types      常用结构体
    /utils      常用函数
