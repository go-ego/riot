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
	// "os"
	// "time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-ego/riot/net/com"
	pb "github.com/go-ego/riot/net/grpc/riot-pb"
	zlog "github.com/go-vgo/gt/zlog"
)

type Doc struct {
	DocId   uint64
	Content string
	// Attri   interface{}
	Attri  []byte
	Tokens []*pb.TokenData
	// Tokens      []types.TokenData
	Labels      []string
	Fields      interface{}
	ForceUpdate bool
}

func InitGrpcc(address string, doc Doc) int32 {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		zlog.Error("did not connect: ", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	r, err := c.DocInx(context.Background(), &pb.DocReq{
		Content: doc.Content,
		Attri:   doc.Attri,
		Tokens:  doc.Tokens,
	})
	if err != nil {
		log.Fatalf("init docinx grpcc could not greet: %v", err)
		zlog.Error("init docinx grpcc could not greet:", err)
		return r.Result
	}
	// log.Printf("Greeting: %s", r.Message)
	return r.Result
}

// InitSearchRpc init search grpc
func InitSearchRpc(address string, sea com.SearchArgs) (*pb.SearchReply, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		zlog.Error("did not connect: ", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	logicExp := &pb.LogicExpression{
		MustLabels:   sea.Logic.LogicExpression.MustLabels,
		ShouldLabels: sea.Logic.LogicExpression.ShouldLabels,
		NotInLabels:  sea.Logic.LogicExpression.NotInLabels,
	}

	logic := &pb.Logic{
		MustLabels:      sea.Logic.MustLabels,
		ShouldLabels:    sea.Logic.ShouldLabels,
		NotInLabels:     sea.Logic.NotInLabels,
		LogicExpression: logicExp,
	}

	// Contact the server and print out its response.
	r, err := c.Search(context.Background(), &pb.SearchReq{
		Id:           sea.Id,
		Query:        sea.Query,
		OutputOffset: int32(sea.OutputOffset),
		MaxOutputs:   int32(sea.MaxOutputs),
		Time:         sea.Time,
		DocIds:       sea.DocIds,
		Logic:        logic,
	})

	if err != nil {
		log.Fatalf("init searchRpc could not greet: %v", err)
		zlog.Error("init searchRpc could not greet:", err)
		return nil, err
	}
	log.Printf("Greeting: %s", r)

	return r, nil
}

// InitDelGrpcc init delete grpcc
func InitDelGrpcc(address string, docid uint64) int32 {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		zlog.Error("did not connect: ", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	r, err := c.Delete(context.Background(), &pb.DeleteReq{
		DocId: docid,
	})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
		zlog.Error("could not greet:", err)
		return r.Result
	}

	// log.Printf("Greeting: %d", r.Result)
	return r.Result
}
