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
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-ego/riot/net/com"
	pb "github.com/go-ego/riot/net/grpc/riot-pb"
	"github.com/go-ego/riot/types"
	grpclb "github.com/go-vgo/grpclb"
)

// InitErpc init grpc
func InitErpc(num int) {
	log.Println("config.Etcd", config.Etcd)
	serv := flag.String("service", config.Etcd.SverName, "service name")
	port := flag.Int("port", config.Etcd.Port[num], "listening port")
	reg := flag.String("reg", config.Etcd.Addr, "register etcd address")

	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		panic(err)
	}

	err = grpclb.Register(*serv, "127.0.0.1", *port, *reg, time.Second*10, 15)
	if err != nil {
		panic(err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		log.Printf("receive signal '%v'", s)
		grpclb.UnRegister()
		os.Exit(1)
	}()

	log.Printf("starting riot service at %d", *port)
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &eserver{})
	s.Serve(lis)
}

// eserver is used to implement riot.GreeterServer.
type eserver struct{}

func (s *eserver) HeartBeat(ctx context.Context, in *pb.HeartReq) (*pb.Reply, error) {

	return &pb.Reply{Result: in.Msg}, nil
}

func (s *eserver) DocInx(ctx context.Context, in *pb.DocReq) (*pb.Reply, error) {
	addDoc(in)
	return &pb.Reply{Result: 0}, nil
}

func (s *eserver) Delete(ctx context.Context, in *pb.DeleteReq) (*pb.Reply, error) {

	DelDoc(in)
	return &pb.Reply{Result: 0}, nil
}

func (s *eserver) Search(ctx context.Context, in *pb.SearchReq) (*pb.SearchReply, error) {

	var (
		outputOffset = int(in.OutputOffset)
		maxOutputs   = int(in.MaxOutputs)
	)

	if in.MaxOutputs == 0 {
		outputOffset = config.Engine.OutputOffset
		maxOutputs = config.Engine.MaxOutputs
	}

	var logic types.Logic
	if in.Logic != nil {
		logicExp := types.LogicExpression{
			MustLabels:   in.Logic.LogicExpression.MustLabels,
			ShouldLabels: in.Logic.LogicExpression.ShouldLabels,
			NotInLabels:  in.Logic.LogicExpression.NotInLabels,
		}

		logic = types.Logic{
			MustLabels:      in.Logic.MustLabels,
			ShouldLabels:    in.Logic.ShouldLabels,
			NotInLabels:     in.Logic.NotInLabels,
			LogicExpression: logicExp,
		}
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

	rep := wgGrpc(sea)
	log.Println("rep...", rep)
	return rep, nil
}
