# riot 全文搜索引擎

<!--<img align="right" src="https://raw.githubusercontent.com/go-ego/ego/master/logo.jpg">-->
<!--[![Build Status](https://travis-ci.org/go-ego/ego.svg)](https://travis-ci.org/go-ego/ego)
[![codecov](https://codecov.io/gh/go-ego/ego/branch/master/graph/badge.svg)](https://codecov.io/gh/go-ego/ego)-->
<!--<a href="https://circleci.com/gh/go-ego/ego/tree/dev"><img src="https://img.shields.io/circleci/project/go-ego/ego/dev.svg" alt="Build Status"></a>-->
[![CircleCI Status](https://circleci.com/gh/go-ego/riot.svg?style=shield)](https://circleci.com/gh/go-ego/riot)
[![Build Status](https://travis-ci.org/go-ego/riot.svg)](https://travis-ci.org/go-ego/riot)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-ego/riot)](https://goreportcard.com/report/github.com/go-ego/riot)
[![GoDoc](https://godoc.org/github.com/go-ego/riot?status.svg)](https://godoc.org/github.com/go-ego/riot)
[![Release](https://github-release-version.herokuapp.com/github/go-ego/riot/release.svg?style=flat)](https://github.com/go-ego/riot/releases/latest)
[![Join the chat at https://gitter.im/go-ego/ego](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/go-ego/ego?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
<!--<a href="https://github.com/go-ego/ego/releases"><img src="https://img.shields.io/badge/%20version%20-%206.0.0%20-blue.svg?style=flat-square" alt="Releases"></a>-->

* [高效索引和搜索](/docs/benchmarking.md)（1M条微博500M数据28秒索引完，1.65毫秒搜索响应时间，19K搜索QPS）
* 支持中文分词（使用[gse分词包](https://github.com/go-ego/gse)并发分词，速度27MB/秒）
* 支持逻辑搜索
* 支持中文转拼音搜索
* 支持计算关键词在文本中的[紧邻距离](/docs/token_proximity.md)（token proximity）
* 支持计算[BM25相关度](/docs/bm25.md)
* 支持[自定义评分字段和评分规则](/docs/custom_scoring_criteria.md)
* 支持[在线添加、删除索引](/docs/realtime_indexing.md)
* 支持多种[持久存储](/docs/persistent_storage.md)
* 支持分布式索引和搜索
* 可实现[分布式索引和搜索](/docs/distributed_indexing_and_search.md)
* 采用对商业应用友好的[Apache License v2](/license.txt)发布

QQ 群: 120563750

## 安装/更新

```
go get -u github.com/go-ego/riot
```

## Requirements

需要 Go 版本至少 1.8

## [Build-tools](https://github.com/go-ego/re)
```
go get -u github.com/go-ego/re 
```
### re riot
创建 riot 项目

```
$ re riot my-riotapp
```

### re run

运行我们创建的 riot 项目, 你可以导航到应用程序文件夹并执行:
```
$ cd my-riotapp && re run
```

## 使用

先看一个例子（来自[simplest_example.go](/examples/simple/simplest_example.go)）
```go
package main

import (
	"log"

	"github.com/go-ego/riot/engine"
	"github.com/go-ego/riot/types"
)

var (
	// searcher是协程安全的
	searcher = engine.Engine{}
)

func main() {
	// 初始化
	searcher.Init(types.EngineInitOptions{
		SegmenterDict: "github.com/go-ego/riot/data/dict/dictionary.txt"})
	defer searcher.Close()

	// 将文档加入索引，docId 从1开始
	searcher.IndexDocument(1, types.DocIndexData{Content: "此次百度收购将成中国互联网最大并购"}, false)
	searcher.IndexDocument(2, types.DocIndexData{Content: "百度宣布拟全资收购91无线业务"}, false)
	searcher.IndexDocument(3, types.DocIndexData{Content: "百度是中国最大的搜索引擎"}, false)

	// 等待索引刷新完毕
	searcher.FlushIndex()

	// 搜索输出格式见types.SearchResponse结构体
	log.Print(searcher.Search(types.SearchRequest{Text:"百度中国"}))
}
```

是不是很简单！

然后看看一个[入门教程](/docs/codelab.md)，教你用不到200行Go代码实现一个微博搜索网站。

## Donate

支持 riot, [buy me a coffee](https://github.com/go-vgo/buy-me-a-coffee).

#### Paypal

Donate money by [paypal](https://www.paypal.me/veni0/25) to my account [vzvway@gmail.com](vzvway@gmail.com)

## 其它

* [为什么要有 riot 引擎](/docs/why_wukong.md)
* [联系方式](/docs/feedback.md)

## License

Riot is primarily distributed under the terms of the Apache License (Version 2.0), base on [wukong](https://github.com/huichen/wukong).
