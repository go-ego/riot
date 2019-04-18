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

package utils

import (
	"bytes"
	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"log"
)
var Json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	extra.SupportPrivateFields()
}
// AbsInt return to the opposite number
func AbsInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// MinInt return to the small number
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func EncodeToBytes(input interface{}) []byte {
	var buf bytes.Buffer
	enc := Json.NewEncoder(&buf)
	err := enc.Encode(input)
	if err != nil {
		log.Print(err)
	}
	return buf.Bytes()
}

func DecodeFromBytes(input []byte, output interface{}) error {
	dec := Json.NewDecoder(bytes.NewBuffer(input))
	return dec.Decode(output)
}
