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
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/go-ego/riot/net/com"
	pb "github.com/go-ego/riot/net/grpc/riot-pb"
	"github.com/go-ego/riot/types"
	zlog "github.com/go-vgo/gt/zlog"
	"github.com/gogo/protobuf/proto"
)

// server is used to implement msg.GreeterServer.
type server struct{}

func addDoc(in *pb.DocReq) {
	tokens := []types.TokenData{}
	for i := 0; i < len(in.Tokens); i++ {
		var loc []int
		inloc := in.Tokens[i].Locations
		for l := 0; l < len(inloc); l++ {
			loc = append(loc, int(inloc[l]))
		}

		tokens = append(tokens, types.TokenData{
			Text:      in.Tokens[i].Text,
			Locations: loc,
		})
		// fmt.Println(in.Tokens[i])
	}
	// fmt.Println(in)

	req := new(pb.Attri)
	err := proto.Unmarshal(in.Attri, req)
	if err != nil {
		log.Println("proto.Unmarshal...", err)
	}

	attri := types.Attri{
		Time: req.Time,
		Ts:   req.Ts,
	}

	data := types.DocIndexData{
		Content: in.Content,
		Attri:   attri,
		Tokens:  tokens,
		// Labels: in.Labels,
		// Fields: in.Fields,
	}

	com.AddDocInx(in.DocId, data, in.ForceUpdate)
}

// DelDoc delete doc
func DelDoc(in *pb.DeleteReq) {
	searcher.RemoveDocument(in.DocId, false)
}

func (s *server) HeartBeat(ctx context.Context, in *pb.HeartReq) (*pb.Reply, error) {

	return &pb.Reply{Result: in.Msg}, nil
}

func (s *server) DocInx(ctx context.Context, in *pb.DocReq) (*pb.Reply, error) {

	addDoc(in)
	return &pb.Reply{Result: 0}, nil
}

func (s *server) Delete(ctx context.Context, in *pb.DeleteReq) (*pb.Reply, error) {

	DelDoc(in)
	return &pb.Reply{Result: 0}, nil
}

func (s *server) Search(ctx context.Context, in *pb.SearchReq) (*pb.SearchReply, error) {

	// time.Sleep(1 * time.Second)
	logicExp := types.LogicExpression{
		MustLabels:   in.Logic.LogicExpression.MustLabels,
		ShouldLabels: in.Logic.LogicExpression.ShouldLabels,
		NotInLabels:  in.Logic.LogicExpression.NotInLabels,
	}

	logic := types.Logic{
		MustLabels:      in.Logic.MustLabels,
		ShouldLabels:    in.Logic.ShouldLabels,
		NotInLabels:     in.Logic.NotInLabels,
		LogicExpression: logicExp,
	}

	var (
		outputOffset = int(in.OutputOffset)
		maxOutputs   = int(in.MaxOutputs)
	)

	if in.MaxOutputs == 0 {
		outputOffset = config.Engine.OutputOffset
		maxOutputs = config.Engine.MaxOutputs
	}

	sea := com.SearchArgs{
		Id:           in.Id,
		Query:        in.Query,
		Time:         in.Time,
		DocIds:       in.DocIds,
		OutputOffset: outputOffset,
		MaxOutputs:   maxOutputs,
		Logic:        logic,
	}

	rep := rpcSearch(sea)
	log.Println("rep...", rep)

	return rep, nil
}

// InitGrpc init grpc
func InitGrpc(port string, args ...bool) {
	log.Println("grpc.port...", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		zlog.Error("failed to listen:", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		zlog.Error("failed to serve:", err)
	}

	log.Printf("listen to: %s \n", port)
}
