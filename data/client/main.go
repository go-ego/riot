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
	"context"
	"flag"
	"log"
	"time"

	pb "github.com/go-ego/riot/net/grpc/riot-pb"
	grpclb "github.com/go-vgo/grpclb"
	"google.golang.org/grpc"
)

var (
	msg  = "在路上, in the way"
	serv = flag.String("service", "riot_service", "service name")
	reg  = flag.String("reg", "http://127.0.0.1:2379", "register etcd address")
)

func main() {
	flag.Parse()
	r := grpclb.NewResolver(*serv)
	b := grpc.RoundRobin(r)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	conn, err := grpc.DialContext(
		ctx, *reg, grpc.WithInsecure(), grpc.WithBalancer(b), grpc.WithBlock())

	if err != nil {
		log.Println("dial grpc: ", err)
		// return
	}

	add(conn)
	search(conn)

	defer cancel()
}

func add(conn *grpc.ClientConn) {
	client := pb.NewGreeterClient(conn)
	for i := 0; i < 1000; i++ {
		resp, err := client.DocInx(context.Background(), &pb.DocReq{
			DocId:       "20",
			Content:     msg,
			ForceUpdate: false,
		})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
			return
		}

		log.Printf("Greeting: %d", resp.Result)
	}
}

func search(conn *grpc.ClientConn) {
	client := pb.NewGreeterClient(conn)
	resp, err := client.Search(context.Background(), &pb.SearchReq{
		Id:    "20",
		Query: "路上",
	})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return
	}

	log.Printf("Greeting: %v", resp)

}

func del(conn *grpc.ClientConn) {
	client := pb.NewGreeterClient(conn)
	resp, err := client.Delete(context.Background(), &pb.DeleteReq{
		DocId: "20",
	})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return
	}

	log.Printf("Greeting: %d", resp.Result)
}
