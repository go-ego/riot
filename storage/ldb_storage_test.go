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
	"os"
	"testing"

	"github.com/go-ego/gwk/utils"
)

func TestOpenOrCreateLdb(t *testing.T) {
	db, err := OpenLeveldbStorage("ldb_test")
	utils.Expect(t, "<nil>", err)
	db.Close()

	db, err = OpenLeveldbStorage("ldb_test")
	utils.Expect(t, "<nil>", err)
	err = db.Set([]byte("key1"), []byte("value1"))
	utils.Expect(t, "<nil>", err)

	buffer := make([]byte, 100)
	buffer, err = db.Get([]byte("key1"))
	utils.Expect(t, "<nil>", err)
	utils.Expect(t, "value1", string(buffer))

	walFile := db.WALName()
	db.Close()
	os.Remove(walFile)
	os.RemoveAll("ldb_test")
}
