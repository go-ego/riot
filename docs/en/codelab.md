Riot engine getting started
====

At the end of this chapter, you will learn to write a simple full-text twitter search using the riot engine.

Before reading this article, you need to have a basic understanding of Go, and if you do not yet use Go, [here](http://go-tour-en.appspot.com/#1) have a tutorial.

## The principle of the engine

Engine processes user requests, word segmentation, indexing and sorting are done by different goroutines.

	1. The main goroutine, used to send and receive user requests
	2. Segmenter goroutine, responsible for word segmentation
	3. The indexer goroutine, responsible for creating and finding index tables
	4. Ranker goroutine, responsible for ranking documents
    
![](https://raw.github.com/go-ego/riot/master/docs/zh/gwk.png)

**Indexing process**

When a request to index a document comes in, the main corpus sends the text of the participle to a word segmentation goroutine over a channel, which sends the text to another word segmentation Indexer Association. The indexer goroutines build an inverted index from the search keyword to the document, and the inverted index table is kept in memory for quick and easy recall.

**Search process**

The main goroutine receives the user's request, divides the request phrase in the main goroutine, and then sends it to the indexer through the channel. The indexer looks for the document corresponding to each search key and then performs logical operations (merge intersection and intersection) to obtain a condensed document List, which is passed through the channel to the sequencer scoring, sorting and sorting the document and then sending the sorted document to the main goroutine over the designated channel, the main goroutine returning the result to the user .

There are multiple goroutines for word segmentation, indexing and sorting. The intermediate results are stored in the channel buffer queue to avoid blocking. In order to improve the search concurrency to reduce the delay riot engine splits the document (the number of splits can be specified by the user), indexing and sorting request will be sent to all splits on the parallel processing, the results of the secondary reduction in the main program Sort.

Above is the general principle of riot engine. Any complete search system includes four sections, ** Document Capture **, ** Index **, ** Search ** and ** Display **. The following sections explain how these parts are achieved.

## Document capture

Document capture a lot of technology, more to come up alone to write an article. Fortunately microblogging crawling is relatively simple and can be done through the API provided by Sina, and there is a [Go Language SDK] (http://github.com/huichen/gobo) that can be crawled concurrently and at a fraction of the speed.

I've already got about 100,000 tweets in testdata/weibo_data.txt (deleted because of the download speed of git clone, please click [here] (https://github.com/go-ego/riot/blob/43f20b4c0921cc704cf41fe8653e66a3fcbb7e31/testdata/weibo_data.txt?raw=true), so you do not need to do it yourself. Each line in the document stores a microblogging, the format is as follows

    <微博id>||||<时间戳>||||<用户id>||||<用户名>||||<转贴数>||||<评论数>||||<喜欢数>||||<小图片网址>||||<大图片网址>||||<正文>

Weibo saved in the following structure for easy query, only loaded the data we need:

```go
type Weibo struct {
        Id           uint64
        Timestamp    uint64
        UserName     string
        RepostsCount uint64
        Text         string
}
```

If you are interested in the details of the capture please see the crawler [testdata/crawl_weibo_data.go](/testdata/crawl_weibo_data.go)。

## Index

You need to import two packages using the riot engine

```go
import (
	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)
```
The first package defines the engine function, the second package defines the common structure. Need to initialize before using engine, for example:

```go
var searcher riot.Engine
searcher.Init(types.EngineOpts{
	SegmenterDict: "../../data/dict/dictionary.txt",
	StopTokenFile:         "../../data/dict/stop_tokens.txt",
	IndexerOpts: &types.IndexerOpts{
		IndexType: types.LocationsIndex,
	},
})
```
[types.EngineOpts](/types/engine_init_options.go) defines parameters that need to be set by the initialization engine, such as where to load the word dictionary file, stop word list, indexer type, BM25 parameters, etc., as well as the default scoring rules (see the "Search" section) and Output Paging Option. Please read the structure of the code for details.

In particular, it should be emphasized that please carefully choose IndexerOpts.IndexType types, there are three different types of index table:

1. DocIdsIndex, provides the most basic index, only record the document key docid appears.
2. FrequenciesIndex, in addition to record docid, but also save the search key appear in each document frequency, if you need BM25 then FrequenciesIndex is what you need.
3. LocationsIndex,  This includes not only the contents of the two indexes, but also an extra storage of the specific location of the keywords in the document, which is used [Calculate the immediate distance](/docs/en/token_proximity.md)。

These three indexes consume more memory from top to bottom while providing more computing power, especially LocationsIndex, which consumes a lot of memory when the document is long. Please choose the right balance.

Initialization can be added after the index, the following example will be a microblogging engine:

```go
searcher.IndexDoc(docId, types.DocIndexData{
	Content: weibo.Text, // Weibo structure See above definition. Must be UTF-8 format.
	Fields: WeiboScoringFields{
		Timestamp:    weibo.Timestamp,
		RepostsCount: weibo.RepostsCount,
	},
})
```

Document docId must be greater than 0 and unique, for Weibo can use Weibo ID. riot engine allows you to add three kinds of index data:

1. The content of the document is indexed by the word tokens.
2. Document tokens. When the body is empty, allowing the user to bypass the riot built-in tokenizer directly enter the document keywords, which makes it possible to document word outside the engine.
3. Document attributes labels (such as microblogging author, category, etc.). The label does not appear in the body.
4. Custom scoring fields, which allow you to add ** arbitrary type **, ** arbitrary structure ** data to the document for sorting. The "Search" section further describes the use of custom rating fields.

**Special attention** is given to the fact that keywords such as tokens and labels make up the search key in the indexer. Do not confuse these three concepts with documents and code. The search for the text is a logical query on the search key. For example, the keyword "bicycle" appears in the body of a document, and the category label such as "fitness", but the word "fitness" does not directly appear in the text. This article will be queried when searching for a combination of search keys such as "Bike" + "Fitness". The purpose of the label is to facilitate narrowing down the query quickly from a non-literal dimension.

The engine uses asynchronous indexing, which means that the index may not have been added to the indexed table when IndexDoc returns, which makes it easy for you to iterate through the index. If you need to wait for the index is added after the follow-up operation, please call the following function:

```go
searcher.FlushIndex()
```

## Search

The search process is a two-step process. The first step is to find the document containing the search key in the index table, as described in the previous section. The second step is to sort all the documents indexed.

The core of sorting is document rating. The riot engine allows you to customize any scoring criteria. In the Weibo search example, we define the scoring rules as follows:

1. First of all, according to the keyword close to the distance, for example, search for "bike", the phrase will be divided into two keywords, "bike" and "sports", two keywords appear next to the article should be ranked in two key words The front of the article.
2. And then roughly classified according to the time of publication of microblogging, every three days as a echelon, the echelon of the article in the front row.
3. Finally microblogging for the BM25 * (1 + forwarding number / 10000)

Such rules need to save some score data for each document, such as microblogging release time, the number of microblogging forwarding. The data is stored in the following structure:

```go
type WeiboScoringFields struct {
        Timestamp    uint64
        RepostsCount uint64
}
```
As you may have noticed, this is the type of argument passed in the IndexDoc function that was called when the document was indexed in the previous section (in fact that argument is of type interface {}, so you can pass in any type of structure). These data are saved in the sequencer's memory to be called.

With these data, we can score, the code is as follows:
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
        if doc.TokenProximity > MaxTokenProximity { // first step
                output[0] = 1.0 / float32(doc.TokenProximity)
        } else {
                output[0] = 1.0
        }
        output[1] = float32(wsf.Timestamp / (SecondsInADay * 3)) // The second step
        output[2] = float32(doc.BM25 * (1 + float32(wsf.RepostsCount)/10000)) // third step
        return output
}
```
WeiboScoringCriteria actually inherits the types.ScoringCriteria interface, which implements the Score function. This function takes two parameters:

1. types.IndexedDocument parameter passes data from the indexer, such as word frequency, the exact location of the word, the BM25 value, the nearest neighbor, and more, as specified in [types/index.go] (/types/index.go).
2. The second parameter is `interface {}` type, you can understand this type C language void pointer, it can point to any data type. In our case we point to the WeiboScoringFields structure and check if it is the correct type by the reflection mechanism.

With custom rating data and custom rating rules, we can search, see the code below:

```go
response := searcher.Search(types.SearchReq{
	Text: "自行车运动",
	RankOpts: &types.RankOpts{
		ScoringCriteria: &WeiboScoringCriteria{},
		OutputOffset: 0,
		MaxOutputs:   100,
	},
})
```

Among them, Text is the input search phrase (must be UTF-8 format), will be divided into keywords. As with indexing, the riot engine allows you to enter keywords and document labels directly by bypassing the built-in tokenizer, see the comments for the types.SearchReq structure. RankOpts defines the sorting options. WeiboScoringCriteria is the scoring rule we defined above. In addition, you can control the output of pagination through the OutputOffset and MaxOutputs parameters. The search results are stored in the response variable, as described in the SearchResp structure defined in the [types/search_response.go] (/types/search_response.go) file. For example, this structure returns the location of the keyword in the document, Used to generate a summary of the document.

## Output

The final step in completing a user search is to present the search results to the user. The usual practice is to make the search engine a background service, and have the front end invoke it as JSON-RPC. The front end does not belong to the riot engine itself so it is not much.

## Summarize

Read here, you should have a basic understanding of the use of riot engine microblogging search, it is recommended that you do it yourself. If you are impatient, take a look at the completed code, see [examples/codelab/search_server.go] (/examples/codelab/search_server.go) for a total of less than 200 lines. Run this example is very simple, enter the examples / codelab directory after input:

	go run search_server.go

After waiting for the output of "索引了xxx条微博" in the terminal, open [http://localhost:8080] (http://localhost:8080) in the browser to enter the search page.

If you want to learn more about riot engine, I suggest you read the code directly. The directory structure of the code is as follows:

    riot/          Engine, including the main goroutine, word segmentation goroutines, indexers goroutines and sequencer goroutines
        /core       Core components, including indexers and sorters
        /data       Dictionary file and stop word file
        /docs       Documents
        /engine     Engine, including the main goroutine, word segmentation goroutines, indexers goroutines and sequencer goroutines. Moved to home directory
        /examples   Examples and performance testing procedures
        /testdata   Test Data
        /types      Common structs
        /utils      Commonly used functions
