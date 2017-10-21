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

package com

type Config struct {
	Engine Engine
	Etcd   Etcd
	Rpc    Rpc

	// Has  [][]string
	Url []string
}

type Engine struct {
	Mode  string
	Using int

	StorageShards int    `toml:"storage_shards"`
	StorageEngine string `toml:"storage_engine"`
	StorageFolder string `toml:"storage_folder"`

	NumShards     int    `toml:"num_shards"`
	OutputOffset  int    `toml:"output_offset"`
	MaxOutputs    int    `toml:"max_outputs"`
	SegmenterDict string `toml:"segmenter_dict"`
	StopTokenFile string `toml:"stop_token_file"`
	Relation      string
	Time          string
	Ts            int64
}

type Rpc struct {
	GrpcPort []string `toml:"grpc_port"`
	// GrpcPort []int    `toml:"grpc_port"`
	DistPort []string `toml:"dist_port"`
	Port     string
}

type Etcd struct {
	Addr     string
	SverName string `toml:"sver_name"`
	Port     []int  `toml:"port"`
}
