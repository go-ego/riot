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

package tikv

import (
	"github.com/pingcap/tidb/config"
	ti "github.com/pingcap/tidb/store/tikv"
	"strings"
)

// Bolt bolt store struct
type Tikv struct {
	cli *ti.RawKVClient
}

type KvData struct {
	Key []byte
	Val []byte
}

// OpenBolt open Bolt store
func OpenTikv(addr string) (*Tikv, error) {
	cli, err := ti.NewRawKVClient(strings.Split(addr, ","), config.Security{})
	if err != nil {
		cli.Close()
		return nil, err
	}
	return &Tikv{cli: cli}, nil
}

// Set executes a function within the context of a read-write managed
// transaction. If no error is returned from the function then the transaction
// is committed. If an error is returned then the entire transaction is rolled back.
// Any error that is returned from the function or returned from the commit is returned
// from the Update() method.
func (s *Tikv) Set(k []byte, v []byte) error {
	return s.cli.Put(k, v)
}

// Get executes a function within the context of a managed read-only transaction.
// Any error that is returned from the function is returned from the View() method.
func (s *Tikv) Get(k []byte) (b []byte, err error) {
	return s.cli.Get(k)
}

// Delete deletes a key. Exposing this so that user does not
// have to specify the Entry directly.
func (s *Tikv) Delete(k []byte) error {
	return s.cli.Delete(k)
}

// Has returns true if the DB does contains the given key.
func (s *Tikv) Has(k []byte) (bool, error) {
	d, err := s.Get(k)

	if err != nil || len(d) == 0 {
		return false, err
	}
	return true, nil
}

func (s *Tikv) BatchPut(data map[string][]byte) {
	var keys, values [][]byte
	for k,v := range data {
		keys = append(keys, []byte(k))
		values = append(values, v)
	}
	s.cli.BatchPut(keys, values)
}

func (s *Tikv) BatchDelete(keys [][]byte) {
	s.cli.BatchDelete(keys)
}

//左匹配，开区间
func (s *Tikv) PreLike(key []byte) (keys [][]byte, values [][]byte, err error) {
	return s.cli.Scan(key, append(key, 255), ti.MaxRawKVScanLimit)
}

// Close releases all database resources. All transactions
// must be closed before closing the database.
func (s *Tikv) Close() error {
	return s.cli.Close()
}
