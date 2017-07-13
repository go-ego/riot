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

package storage

import (
	"log"
	"os"

	"github.com/dgraph-io/badger"
)

// Badger is badger.KV
type Badger struct {
	db *badger.KV
}

// OpenBadgerStorage open Badger storage
func OpenBadgerStorage(dbPath string) (Storage, error) {

	err := os.MkdirAll(dbPath, 0777)
	if err != nil {
		log.Fatal("MkdirAll: ", err)
		os.Exit(1)
	}
	// os.MkdirAll(path.Dir(dbPath), os.ModePerm)

	opt := badger.DefaultOptions
	opt.Dir = dbPath
	opt.ValueDir = dbPath
	opt.SyncWrites = true
	kv, err := badger.NewKV(&opt)
	if err != nil {
		log.Fatal("NewKV: ", err)
	}

	return &Badger{kv}, err
}

// WALName is useless for this kv database
func (s *Badger) WALName() string {
	return "" //对于此数据库，本函数没用~
}

// Set sets the provided value for a given key.
// If key is not present, it is created. If it is present,
// the existing value is overwritten with the one provided.
func (s *Badger) Set(k, v []byte) error {
	return s.db.Set(k, v)
}

// Get looks for key and returns a value.
// If key is not found, value is nil.
func (s *Badger) Get(k []byte) ([]byte, error) {
	var item badger.KVItem
	err := s.db.Get(k, &item)
	return item.Value(), err
}

// Delete deletes a key. Exposing this so that user does not
// have to specify the Entry directly. For example, BitDelete
// seems internal to badger.
func (s *Badger) Delete(k []byte) error {
	return s.db.Delete(k)
}

// ForEach get all key and value
func (s *Badger) ForEach(fn func(k, v []byte) error) error {
	itrOpt := badger.IteratorOptions{
		PrefetchSize: 1000,
		FetchValues:  true,
		Reverse:      false,
	}
	itr := s.db.NewIterator(itrOpt)

	for itr.Rewind(); itr.Valid(); itr.Next() {
		item := itr.Item()

		key := item.Key()
		val := item.Value()

		if err := fn(key, val); err != nil {
			return err
		}
	}
	return nil
}

// Close closes a KV. It's crucial to call it to ensure
// all the pending updates make their way to disk.
func (s *Badger) Close() error {
	return s.db.Close()
}
