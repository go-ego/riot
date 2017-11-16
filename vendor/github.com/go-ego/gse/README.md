# gse

Go efficient text segmentation; support english, chinese, japanese and other.

<!--<img align="right" src="https://raw.githubusercontent.com/go-ego/ego/master/logo.jpg">-->
<!--<a href="https://circleci.com/gh/go-ego/ego/tree/dev"><img src="https://img.shields.io/circleci/project/go-ego/ego/dev.svg" alt="Build Status"></a>-->
[![CircleCI Status](https://circleci.com/gh/go-ego/gse.svg?style=shield)](https://circleci.com/gh/go-ego/gse)
[![codecov](https://codecov.io/gh/go-ego/gse/branch/master/graph/badge.svg)](https://codecov.io/gh/go-ego/gse)
[![Build Status](https://travis-ci.org/go-ego/gse.svg)](https://travis-ci.org/go-ego/gse)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-ego/gse)](https://goreportcard.com/report/github.com/go-ego/gse)
[![GoDoc](https://godoc.org/github.com/go-ego/gse?status.svg)](https://godoc.org/github.com/go-ego/gse)
[![Release](https://github-release-version.herokuapp.com/github/go-ego/gse/release.svg?style=flat)](https://github.com/go-ego/gse/releases/latest)
[![Join the chat at https://gitter.im/go-ego/ego](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/go-ego/ego?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
<!--<a href="https://github.com/go-ego/ego/releases"><img src="https://img.shields.io/badge/%20version%20-%206.0.0%20-blue.svg?style=flat-square" alt="Releases"></a>-->

[简体中文](https://github.com/go-ego/gse/blob/master/README_zh.md)

<a href="https://github.com/go-ego/gse/blob/master/dictionary.go">Dictionary </a> with double array trie (Double-Array Trie) to achieve,
<a href="https://github.com/go-ego/gse/blob/master/segmenter.go">Sender </a> algorithm is the shortest path based on word frequency plus dynamic programming.

Support common and search engine two participle mode, support user dictionary, POS tagging, run<a href="https://github.com/go-ego/gse/blob/master/server/server.go"> JSON RPC service</a>.

Text Segmentation speed<a href="https://github.com/go-ego/gse/blob/master/tools/benchmark.go"> single thread</a> 9MB/s，<a href="https://github.com/go-ego/gse/blob/master/tools/goroutines.go">goroutines concurrent</a> 42MB/s (8 nuclear Macbook Pro).

## Install / update

```
go get -u github.com/go-ego/gse
```

## [Build-tools](https://github.com/go-ego/re)
```
go get -u github.com/go-ego/re 
```
### re gse
To create a new gse application

```
$ re gse my-gse
```

### re run

To run the application we just created, you can navigate to the application folder and execute:
```
$ cd my-gse && re run
```


## Use


```go
package main

import (
	"fmt"

	"github.com/go-ego/gse"
)

func main() {
	// Load the dictionary
	var segmenter gse.Segmenter
	segmenter.LoadDict()
	// segmenter.LoadDict("your gopath"+"/src/github.com/go-ego/gse/data/dict/dictionary.txt")

	// Text Segmentation
	text := []byte("你好世界, Hello world.")
	segments := segmenter.Segment(text)
  
	// Handle word segmentation results
	// Support for normal mode and search mode two participle, see the comments in the code ToString function.
	fmt.Println(gse.ToString(segments, false)) 
}
```
## License

Gse is primarily distributed under the terms of both the MIT license and the Apache License (Version 2.0), base on [sego](https://github.com/huichen/sego)
