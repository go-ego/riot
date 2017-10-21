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

package grpc

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/go-ego/riot/net/com"
	pb "github.com/go-ego/riot/net/grpc/riot-pb"
	"github.com/go-ego/riot/types"
)

var (
	// config   = com.Conf
	config   com.Config
	searcher = com.Searcher
)

// InitEngine init engine
func InitEngine(conf com.Config) {
	config = conf
	com.InitEngine(conf)
}

type rpcSlice []*pb.Text

func (s rpcSlice) Len() int      { return len(s) }
func (s rpcSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s rpcSlice) Less(i, j int) bool {

	// sid := strings.Split(s[i].Id, "-")
	// lenstr := len(sid) - 1
	// ival, err := strconv.Atoi(sid[lenstr])
	// if err != nil {
	// 	log.Println("strconv.Atoi(sid)...", err)
	// }

	// jid := strings.Split(s[j].Id, "-")
	// lenjid := len(jid) - 1
	// jval, err := strconv.Atoi(jid[lenjid])
	// if err != nil {
	// 	log.Println("strconv.Atoi(jid)...", err)
	// }

	if s[i].Attri.Ts == s[j].Attri.Ts {
		// return ival > jval
		return s[i].Id > s[j].Id
	}
	return s[i].Attri.Ts > s[j].Attri.Ts
}

// rpcSearch rpc search fn
func rpcSearch(sea com.SearchArgs) *pb.SearchReply {
	var (
		// outputOffset int = sea.OutputOffset
		maxOutputs int = sea.MaxOutputs
	)
	if maxOutputs == 0 {
		// outputOffset = config.Engine.OutputOffset
		maxOutputs = config.Engine.MaxOutputs
	}

	docs := com.Search(sea)
	var textArr []*pb.Text
	for i := 0; i < len(docs.Docs); i++ {
		attri := &pb.Attri{
			Time: docs.Docs[i].Attri.(types.Attri).Time,
			Ts:   docs.Docs[i].Attri.(types.Attri).Ts,
		}

		text := &pb.Text{
			Id:      docs.Docs[i].DocId,
			Content: docs.Docs[i].Content,
			Attri:   attri,
		}
		textArr = append(textArr, text)
	}

	sort.Sort(rpcSlice(textArr))

	if len(textArr) > maxOutputs {
		textArr = textArr[0:maxOutputs]
	}

	timestamp := time.Now().Unix()
	rep := &pb.SearchReply{
		Len:       int32(len(textArr)),
		Timestamp: timestamp,
		Docs:      textArr}

	return rep
}

var rpcdata []*pb.SearchReply
var rpcwg sync.WaitGroup

// WgRpc rpc
func WgRpc(address string, sea com.SearchArgs) {
	data, err := InitSearchRpc(address, sea)

	if data == nil || err != nil {
		log.Println("data is null ---")
		defer rpcwg.Done()
		return
	}

	rpcdata = append(rpcdata, data)

	defer rpcwg.Done()
}

func wgGrpc(sea com.SearchArgs) *pb.SearchReply {
	rpcdata = nil
	var (
		distData   *pb.SearchReply
		maxOutputs = sea.MaxOutputs
	)

	if maxOutputs == 0 {
		maxOutputs = config.Engine.MaxOutputs
	}

	// config.Grpc.GrpcPort
	for i := 0; i < len(config.Rpc.DistPort); i++ {
		rpcwg.Add(1)
		addr := config.Rpc.DistPort[i]
		go WgRpc(addr, sea)
	}

	rpcwg.Wait()

	if len(rpcdata) == 1 {
		distData = rpcdata[0]
	} else {
		var docs []*pb.Text
		for i := 0; i < len(rpcdata); i++ {
			for d := 0; d < len(rpcdata[i].Docs); d++ {
				docs = append(docs, rpcdata[i].Docs[d])
			}
		}
		sort.Sort(rpcSlice(docs))

		if len(docs) > maxOutputs {
			end := maxOutputs - 1
			docs = docs[0:end]
		}

		timestamp := time.Now().Unix()

		response := &pb.SearchReply{
			Len:       int32(len(docs)),
			Timestamp: timestamp,
			Docs:      docs}

		distData = response
	}

	return distData
}
