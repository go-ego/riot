// Copyright 2013 Hui Chen
// Copyright 2016 ego authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package engine

import (
	// "fmt"
	"strings"

	"github.com/go-ego/gpy"
	"github.com/go-ego/riot/types"
)

type segmenterRequest struct {
	docId uint64
	hash  uint32
	data  types.DocIndexData
	// data        types.DocumentIndexData
	forceUpdate bool
}

// Map defines the type map[string][]int
type Map map[string][]int

// Segspl split seg
func (engine *Engine) Segspl(splitData []string, num int) (Map, int) {
	var (
		numTokens int
		sqlitStr  string
	)
	tokensMap := make(map[string][]int)
	for i := 0; i < num; i++ {
		if splitData[i] != "" {
			if !engine.stopTokens.IsStopToken(splitData[i]) {
				numTokens++
				tokensMap[splitData[i]] = append(tokensMap[splitData[i]], numTokens)
			}

			sqlitStr += splitData[i]
			if !engine.stopTokens.IsStopToken(sqlitStr) {
				numTokens++
				tokensMap[sqlitStr] = append(tokensMap[sqlitStr], numTokens)
			}

			if engine.initOptions.Using == 6 {
				// more combination
				var sqlitsStr string
				for s := i + 1; s < len(splitData); s++ {
					sqlitsStr += splitData[s]

					if !engine.stopTokens.IsStopToken(sqlitsStr) {
						numTokens++
						tokensMap[sqlitsStr] = append(tokensMap[sqlitsStr], numTokens)
					}
				}
			}

		}
	}

	return tokensMap, numTokens
}

func (engine *Engine) splitData(request segmenterRequest) (Map, int) {
	tokensMap := make(map[string][]int)
	var (
		num       int
		numTokens int
	)

	if request.data.Content != "" {
		request.data.Content = strings.ToLower(request.data.Content)
		if engine.initOptions.Using == 4 {
			// use segmenter
			segments := engine.segmenter.Segment([]byte(request.data.Content))
			for _, segment := range segments {
				token := segment.Token().Text()
				if !engine.stopTokens.IsStopToken(token) {
					tokensMap[token] = append(tokensMap[token], segment.Start())
				}
			}
			numTokens += len(segments)
		}

		if engine.initOptions.Using == 5 {
			// use segmenter
			splitSpaData := strings.Split(request.data.Content, " ")
			num := len(splitSpaData)
			tokenMap, numToken := engine.Segspl(splitSpaData, num)
			numTokens += numToken
			for key, val := range tokenMap {
				tokensMap[key] = val
			}
		}

		if engine.initOptions.Using != 5 {
			splitData := strings.Split(request.data.Content, "")
			num = len(splitData)
			tokenMap, numToken := engine.Segspl(splitData, num)
			numTokens += numToken
			for key, val := range tokenMap {
				tokensMap[key] = val
			}
		}
	}

	for _, t := range request.data.Tokens {
		if !engine.stopTokens.IsStopToken(t.Text) {
			tokensMap[t.Text] = t.Locations
		}
	}

	numTokens += len(request.data.Tokens)

	return tokensMap, numTokens
}

func (engine *Engine) segmenterData(request segmenterRequest) (Map, int) {
	tokensMap := make(map[string][]int)
	numTokens := 0

	if engine.initOptions.Using == 1 && request.data.Content != "" {
		// Content分词, 当文档正文不为空时，优先从内容分词中得到关键词
		segments := engine.segmenter.Segment([]byte(request.data.Content))
		for _, segment := range segments {
			token := segment.Token().Text()
			if !engine.stopTokens.IsStopToken(token) {
				tokensMap[token] = append(tokensMap[token], segment.Start())
			}
		}
		numTokens = len(segments)

		return tokensMap, numTokens
	}

	if engine.initOptions.Using == 2 || ((engine.initOptions.Using == 1 || engine.initOptions.Using == 3) && request.data.Content == "") {
		for _, t := range request.data.Tokens {
			if !engine.stopTokens.IsStopToken(t.Text) {
				tokensMap[t.Text] = t.Locations
			}
		}

		numTokens = len(request.data.Tokens)

		return tokensMap, numTokens
	}

	if engine.initOptions.Using == 3 && request.data.Content != "" {
		// Content分词, 当文档正文不为空时，优先从内容分词中得到关键词
		segments := engine.segmenter.Segment([]byte(request.data.Content))
		for _, segment := range segments {
			token := segment.Token().Text()
			if !engine.stopTokens.IsStopToken(token) {
				tokensMap[token] = append(tokensMap[token], segment.Start())
			}
		}

		for _, t := range request.data.Tokens {
			if !engine.stopTokens.IsStopToken(t.Text) {
				tokensMap[t.Text] = t.Locations
			}
		}

		numTokens = len(segments) + len(request.data.Tokens)

		return tokensMap, numTokens
	}

	tokenMap, lenSplitData := engine.splitData(request)

	return tokenMap, lenSplitData
}

func (engine *Engine) segmenterWorker() {
	for {
		request := <-engine.segmenterChannel
		if request.docId == 0 {
			if request.forceUpdate {
				for i := 0; i < engine.initOptions.NumShards; i++ {
					engine.indexerAddDocChannels[i] <- indexerAddDocumentRequest{forceUpdate: true}
				}
			}
			continue
		}

		shard := engine.getShard(request.hash)
		tokensMap, numTokens := engine.segmenterData(request)

		// 加入非分词的文档标签
		for _, label := range request.data.Labels {
			if !engine.initOptions.NotUsingSegmenter {
				if !engine.stopTokens.IsStopToken(label) {
					//当正文中已存在关键字时，若不判断，位置信息将会丢失
					if _, ok := tokensMap[label]; !ok {
						tokensMap[label] = []int{}
					}
				}
			} else {
				//当正文中已存在关键字时，若不判断，位置信息将会丢失
				if _, ok := tokensMap[label]; !ok {
					tokensMap[label] = []int{}
				}
			}
		}

		indexerRequest := indexerAddDocumentRequest{
			document: &types.DocumentIndex{
				DocId:       request.docId,
				TokenLength: float32(numTokens),
				Keywords:    make([]types.KeywordIndex, len(tokensMap)),
			},
			forceUpdate: request.forceUpdate,
		}
		iTokens := 0
		for k, v := range tokensMap {
			indexerRequest.document.Keywords[iTokens] = types.KeywordIndex{
				Text: k,
				// 非分词标注的词频设置为0，不参与tf-idf计算
				Frequency: float32(len(v)),
				Starts:    v}
			iTokens++
		}

		engine.indexerAddDocChannels[shard] <- indexerRequest
		if request.forceUpdate {
			for i := 0; i < engine.initOptions.NumShards; i++ {
				if i == shard {
					continue
				}
				engine.indexerAddDocChannels[i] <- indexerAddDocumentRequest{forceUpdate: true}
			}
		}
		rankerRequest := rankerAddDocRequest{
			// docId: request.docId, fields: request.data.Fields}
			docId: request.docId, fields: request.data.Fields, content: request.data.Content, attri: request.data.Attri}
		engine.rankerAddDocChannels[shard] <- rankerRequest
	}
}

// PinYin get the Chinese alphabet and abbreviation
func (engine *Engine) PinYin(hans string) []string {
	var (
		str      string
		pystr    string
		strArr   []string
		sqlitStr string
		// sqlitArr []string
	)

	//
	splitHans := strings.Split(hans, "")
	for i := 0; i < len(splitHans); i++ {
		if splitHans[i] != "" {
			strArr = append(strArr, splitHans[i])
			sqlitStr += splitHans[i]
		}
		strArr = append(strArr, sqlitStr)
	}

	// Segment 分词
	if engine.initOptions.NotUsingSegmenter {
		sehans := engine.Segment(hans)
		for h := 0; h < len(sehans); h++ {
			strArr = append(strArr, sehans[h])
		}
	}
	//
	// py := pinyin.LazyConvert(sehans[h], nil)
	py := gpy.LazyConvert(hans, nil)

	for i := 0; i < len(py); i++ {
		str += py[i][0:1]
		pystr += py[i]
		strArr = append(strArr, pystr)
		strArr = append(strArr, str)
	}
	// }

	return strArr
}
