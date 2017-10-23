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

package hb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "github.com/go-ego/riot/net/grpc/riot-pb"
)

var (
	// address string = "localhost:50051"
	rpc int32
)

// Conf config
type Conf struct {
	Addr string
	Path string
	Cmd  string
}

var wg sync.WaitGroup

// Grpcc grpc heartbeat
func Grpcc(addr, dir, bra string) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	for {
		time.Sleep(1 * time.Second)
		heart(c)
		if rpc > 3 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := restart(dir, bra)
				if err != nil {
					log.Println("restart...", err)
				}
			}()
			wg.Wait()

		}
	}
}

func heart(c pb.GreeterClient) {
	r, err := c.HeartBeat(context.Background(), &pb.HeartReq{
		Msg: 1,
	})

	if err != nil {
		rpc++
		if rpc > 3 {
			log.Println("rpc...", rpc)
		}

		log.Printf("could not greet: %v", err)

		return
	}

	log.Printf("Greeting: %d", r.Result)

	if r.Result != 1 {
		rpc++
		log.Println("...", r.Result)
		if rpc > 3 {
			log.Println("rpc...", rpc)
		}
	}
}

func restart(dir, bra string) error {
	rpc = 0
	cmd := exec.Command("/bin/sh", "-c", bra)
	cmd.Dir = dir
	// cmd.Run()
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd.Output: ", err)
		return err
	}

	return nil
}

// Json json
type Json struct {
	Code      int32 `json:"code"`
	Timestamp int64 `json:"timestamp"`
	Health    int32 `json:"health"`
}

// RestartSer restart server
func RestartSer(w http.ResponseWriter, req *http.Request) {
	var name string

	req.ParseForm()
	if len(req.Form["name"]) > 0 {
		name = req.Form["name"][0]
	}
	kill(name)
}

// Health server health
func Health(w http.ResponseWriter, req *http.Request) {
	timestamp := time.Now().Unix()
	response, _ := json.Marshal(&Json{
		Code:      0,
		Timestamp: timestamp,
		Health:    rpc})

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	io.WriteString(w, string(response))
}

// KillSer kill server
func KillSer(w http.ResponseWriter, req *http.Request) {
	var name string

	req.ParseForm()
	if len(req.Form["name"]) > 0 {
		name = req.Form["name"][0]
	}

	kill(name)

	timestamp := time.Now().Unix()
	response, _ := json.Marshal(&Json{
		Code:      0,
		Timestamp: timestamp})

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	io.WriteString(w, string(response))
}

func kill(name string) {
	err := restart("", "killall -9 "+name)
	if err != nil {
		log.Println("kill...", err)
	}
}
