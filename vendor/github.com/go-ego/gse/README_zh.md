# [gse](https://github.com/go-ego/gse)

Go 语言高效分词, 支持英文、中文、日文等

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

<a href="https://github.com/go-ego/gse/blob/master/dictionary.go">词典</a>用双数组trie（Double-Array Trie）实现，
<a href="https://github.com/go-ego/gse/blob/master/segmenter.go">分词器</a>算法为基于词频的最短路径加动态规划。

支持普通和搜索引擎两种分词模式，支持用户词典、词性标注，可运行<a href="https://github.com/go-ego/gse/blob/master/server/server.go">JSON RPC服务</a>。

分词速度<a href="https://github.com/go-ego/gse/blob/master/tools/benchmark.go">单线程</a>9MB/s，<a href="https://github.com/go-ego/gse/blob/master/tools/goroutines.go">goroutines并发</a>42MB/s（8核Macbook Pro）。

## 安装/更新

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


## 使用


```go
package main

import (
	"fmt"

	"github.com/go-ego/gse"
)

func main() {
	// 载入词典
	var segmenter gse.Segmenter
	segmenter.LoadDict()
	// segmenter.LoadDict("your gopath"+"/src/github.com/go-ego/gse/data/dict/dictionary.txt")

	// 分词
	text := []byte("中华人民共和国中央人民政府")
	segments := segmenter.Segment(text)
  
	// 处理分词结果
	// 支持普通模式和搜索模式两种分词，见代码中 ToString 函数的注释。
	fmt.Println(gse.ToString(segments, false)) 

	text1 := []byte("深圳地标建筑, 深圳地王大厦")
	segments1 := seg.Segment([]byte(text1))
	fmt.Println(gse.ToString(segments1, false)) 
}
```
## License

Gse is primarily distributed under the terms of both the MIT license and the Apache License (Version 2.0), base on [sego](https://github.com/huichen/sego)
