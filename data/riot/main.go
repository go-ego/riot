// Copyright 2017 ego authors
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

package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/go-ego/riot/net/com"
	rpc "github.com/go-ego/riot/net/grpc"
	rhttp "github.com/go-ego/riot/net/http"

	"github.com/go-ego/riot"
	"github.com/go-vgo/gt/conf"
	"github.com/go-vgo/gt/zlog"
)

var (
	// searcher is coroutine safe
	searcher = riot.Engine{}

	// config rpc.Config
	config com.Config
	wg     sync.WaitGroup
)

func init() {
	zlog.Init("../conf/log.toml")
	conf.Init("../conf/riot.toml", &config)
	go conf.Watch("../conf/riot.toml", &config)

	rpc.InitEngine(config)

	grpcPort := config.Rpc.GrpcPort
	go rpc.InitGrpc(grpcPort[0])
	distPort := config.Rpc.DistPort
	go rpc.InitGrpc(distPort[0])

	go rpc.InitErpc(0)
}

func main() {

	if config.Engine.Mode == "dev" {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	http.HandleFunc("/search", rhttp.Search)
	http.HandleFunc("/dist", rhttp.WgDist)
	log.Println("listen and serve on 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
