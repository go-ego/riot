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

package http

import (
	"github.com/go-ego/riot/types"
)

// Text search for documents
type Text struct {
	// Id      string      `json:"id"`
	Id      uint64      `json:"id"`
	Content string      `json:"content"`
	Score   []float32   `json:"score"`
	Attri   types.Attri `json:"attri"`
}

// JsonResponse search Json response
type JsonResponse struct {
	Code      int64  `json:"code"`
	Len       int    `json:"len"`
	Timestamp int64  `json:"timestamp"`
	Docs      []Text `json:"docs"`
}

type docsSlice []Text

func (s docsSlice) Len() int      { return len(s) }
func (s docsSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s docsSlice) Less(i, j int) bool {

	// // When id is string, maybe do it
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
	// 	log.Println("strconv.Atoi(Jid)...", err)
	// }

	if s[i].Attri.Ts == s[j].Attri.Ts {
		// return ival > jval
		return s[i].Id > s[j].Id
	}
	return s[i].Attri.Ts > s[j].Attri.Ts
}
