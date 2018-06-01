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
	// DefaultStorage default storage engine
	DefaultStorage = "ldb"
	// DefaultStorage = "bad"
	// DefaultStorage = "bolt"
)

var supportedStorage = map[string]func(path string) (Storage, error){
	"ldb":  OpenLeveldb,
	"bg":   OpenBadger, // bad to bg
	"bolt": OpenBolt,
	// "kv":   OpenKV,
	// "ledisdb": Open,
}

// RegisterStorage register Storage engine
func RegisterStorage(name string, fn func(path string) (Storage, error)) {
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
func OpenStorage(path string, args ...string) (Storage, error) {
	storeName := DefaultStorage

	if len(args) > 0 && args[0] != "" {
		storeName = args[0]
	} else {
		storeEnv := os.Getenv("RIOT_STORAGE_ENGINE")
		if storeEnv != "" {
			storeName = storeEnv
		}
	}

	if fn, has := supportedStorage[storeName]; has {
		return fn(path)
	}

	return nil, fmt.Errorf("unsupported storage engine: %v", storeName)
}
