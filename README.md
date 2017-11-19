# riot full text search engine

<!--<img align="right" src="https://raw.githubusercontent.com/go-ego/ego/master/logo.jpg">-->
<!--<a href="https://circleci.com/gh/go-ego/ego/tree/dev"><img src="https://img.shields.io/circleci/project/go-ego/ego/dev.svg" alt="Build Status"></a>-->
[![CircleCI Status](https://circleci.com/gh/go-ego/riot.svg?style=shield)](https://circleci.com/gh/go-ego/riot)
[![codecov](https://codecov.io/gh/go-ego/riot/branch/master/graph/badge.svg)](https://codecov.io/gh/go-ego/riot)
[![Build Status](https://travis-ci.org/go-ego/riot.svg)](https://travis-ci.org/go-ego/riot)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-ego/riot)](https://goreportcard.com/report/github.com/go-ego/riot)
[![GoDoc](https://godoc.org/github.com/go-ego/riot?status.svg)](https://godoc.org/github.com/go-ego/riot)
[![Release](https://github-release-version.herokuapp.com/github/go-ego/riot/release.svg?style=flat)](https://github.com/go-ego/riot/releases/latest)
[![Join the chat at https://gitter.im/go-ego/ego](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/go-ego/ego?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
<!--<a href="https://github.com/go-ego/ego/releases"><img src="https://img.shields.io/badge/%20version%20-%206.0.0%20-blue.svg?style=flat-square" alt="Releases"></a>-->

[简体中文](https://github.com/go-ego/riot/blob/master/README_zh.md)


* [Efficient indexing and search](/docs/en/benchmarking.md) (1M blog 500M data 28 seconds index finished, 1.65 ms search response time, 19K search QPS）
* Support for logical search
* Support Chinese word segmentation (use [gse word segmentation package](https://github.com/go-ego/gse) concurrent word, speed 27MB / s）
* Support the calculation of the keyword in the text [close to the distance](/docs/zh/token_proximity.md)（token proximity）
* Support calculation [BM25 correlation](/docs/zh/bm25.md)
* Support [custom scoring field and scoring rules](/docs/zh/custom_scoring_criteria.md)
* Support [add online, delete index](/docs/zh/realtime_indexing.md)
* Support heartbeat
* Support multiple [persistent storage](/docs/zh/persistent_storage.md)
* Support distributed index and search
* Can be achieved [distributed index and search](/docs/zh/distributed_indexing_and_search.md)

* [Look at Word segmentation rules](https://github.com/go-ego/riot/blob/master/docs/en/segmenter.md)

## Requirements
Go version >= 1.8

## Installation/Update

```
go get -u github.com/go-ego/riot
```

## [Build-tools](https://github.com/go-ego/re)
```
go get -u github.com/go-ego/re 
```
### re riot
To create a new riot application

```
$ re riot my-riotapp
```

### re run

To run the application we just created, you can navigate to the application folder and execute:
```
$ cd my-riotapp && re run
```

## Usage:

#### [Look at an example](/examples/simple/main.go)

```go
package main

import (
	"log"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

var (
	// searcher is coroutine safe
	searcher = riot.Engine{}
)

func main() {
	// Init
	searcher.Init(types.EngineInitOptions{
		Using:             4,
		NotUsingSegmenter: true})
	defer searcher.Close()

	// Add the document to the index, docId starts at 1
	searcher.IndexDocument(1, types.DocIndexData{Content: "Google Is Experimenting With Virtual Reality Advertising"}, false)
	searcher.IndexDocument(2, types.DocIndexData{Content: "Google accidentally pushed Bluetooth update for Home speaker early"}, false)
	searcher.IndexDocument(3, types.DocIndexData{Content: "Google is testing another Search results layout with rounded cards, new colors, and the 4 mysterious colored dots again"}, false)

	// Wait for the index to refresh
	searcher.FlushIndex()

	// The search output format is found in the types.SearchResponse structure
	log.Print(searcher.Search(types.SearchRequest{Text:"google testing"}))
}
```

It is very simple!

#### [Look at more Examples](https://github.com/go-ego/riot/tree/master/examples)

#### [Look at Store example](https://github.com/go-ego/riot/blob/master/examples/store/main.go)
#### [Look at Logic search example](https://github.com/go-ego/riot/blob/master/examples/logic/main.go)

#### [Look at Pinyin search example](https://github.com/go-ego/riot/blob/master/examples/pinyin/main.go)

#### [Look at different dict and language search example](https://github.com/go-ego/riot/blob/master/examples/dict/main.go)

#### [Look at benchmark example](https://github.com/go-ego/riot/blob/master/examples/benchmark/benchmark.go)

#### [Riot search engine templates, client and dictionaries](https://github.com/go-ego/riot/tree/master/data)

## Donate

Supporting riot, [buy me a coffee](https://github.com/go-vgo/buy-me-a-coffee).

#### Paypal

Donate money by [paypal](https://www.paypal.me/veni0/25) to my account [vzvway@gmail.com](vzvway@gmail.com)

## License

Riot is primarily distributed under the terms of the Apache License (Version 2.0), base on [wukong](https://github.com/huichen/wukong).
