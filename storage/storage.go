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

package storage

import (
	"fmt"
	"os"
)

const (
	DEFAULT_STORAGE_ENGINE = "ldb"

	// DEFAULT_STORAGE_ENGINE = "bad"

	// DEFAULT_STORAGE_ENGINE = "bolt"
)

var supportedStorage = map[string]func(path string) (Storage, error){
	"ldb":  OpenLeveldbStorage,
	"bad":  OpenBadgerStorage,
	"bolt": OpenBoltStorage,
	// "kv":   OpenKVStorage,
	// "ledisdb"
}

// RegisterStorageEngine register Storage engine
func RegisterStorageEngine(name string, fn func(path string) (Storage, error)) {
	supportedStorage[name] = fn
}

// Storage is storage interface
type Storage interface {
	Set(k, v []byte) error
	Get(k []byte) ([]byte, error)
	Delete(k []byte) error
	Has(k []byte) (bool, error)
	ForEach(fn func(k, v []byte) error) error
	Close() error
	WALName() string
}

// OpenStorage open Storage engine
func OpenStorage(path string) (Storage, error) {
	wse := os.Getenv("GWK_STORAGE_ENGINE")
	if wse == "" {
		wse = DEFAULT_STORAGE_ENGINE
	}
	if fn, has := supportedStorage[wse]; has {
		return fn(path)
	}
	return nil, fmt.Errorf("unsupported storage engine %v", wse)
}
