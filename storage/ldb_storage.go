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
	"github.com/syndtr/goleveldb/leveldb"
)

type leveldbStorage struct {
	db *leveldb.DB
}

// OpenLeveldbStorage opens or creates a DB for the given storage. The DB
// will be created if not exist, unless ErrorIfMissing is true.
// Also, if ErrorIfExist is true and the DB exist Open will
// returns os.ErrExist error.
func OpenLeveldbStorage(dbPath string) (Storage, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}

	return &leveldbStorage{db}, nil
}

// WALName is useless for this kv database
func (s *leveldbStorage) WALName() string {
	return "" //对于此数据库，本函数没用~
}

// Set sets the provided value for a given key.
// If key is not present, it is created. If it is present,
// the existing value is overwritten with the one provided.
func (s *leveldbStorage) Set(k, v []byte) error {
	return s.db.Put(k, v, nil)
}

// Get gets the value for the given key. It returns
// ErrNotFound if the DB does not contains the key.
//
// The returned slice is its own copy, it is safe to modify
// the contents of the returned slice. It is safe to modify the contents
// of the argument after Get returns.
func (s *leveldbStorage) Get(k []byte) ([]byte, error) {
	return s.db.Get(k, nil)
}

// Delete deletes the value for the given key. Delete will not
// returns error if key doesn't exist. Write merge also applies
// for Delete, see Write.
//
// It is safe to modify the contents of the arguments after Delete
// returns but not before.
func (s *leveldbStorage) Delete(k []byte) error {
	return s.db.Delete(k, nil)
}

// ForEach get all key and value
func (s *leveldbStorage) ForEach(fn func(k, v []byte) error) error {
	iter := s.db.NewIterator(nil, nil)
	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := iter.Key()
		val := iter.Value()
		if err := fn(key, val); err != nil {
			return err
		}
	}
	iter.Release()
	return iter.Error()
}

// Close closes the DB. This will also releases any outstanding snapshot,
// abort any in-flight compaction and discard open transaction.
func (s *leveldbStorage) Close() error {
	return s.db.Close()
}
