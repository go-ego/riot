package storage

import (
	"log"
	"os"

	"github.com/dgraph-io/badger"
)

type Badger struct {
	db *badger.KV
}

func openBadgerStorage(dbpath string) (Storage, error) {

	err := os.MkdirAll(dbpath, 0777)
	if err != nil {
		log.Fatal("MkdirAll: ", err)
		os.Exit(1)
	}
	// os.MkdirAll(path.Dir(dbpath), os.ModePerm)

	opt := badger.DefaultOptions
	opt.Dir = dbpath
	opt.ValueDir = dbpath
	opt.SyncWrites = true
	kv, err := badger.NewKV(&opt)
	if err != nil {
		log.Fatal("newKv: ", err)
	}

	return &Badger{kv}, err
}

func (s *Badger) WALName() string {
	return "" //对于此数据库，本函数没用~
}

func (s *Badger) Set(k, v []byte) error {
	return s.db.Set(k, v)
}

func (s *Badger) Get(k []byte) ([]byte, error) {
	var item badger.KVItem
	err := s.db.Get(k, &item)
	return item.Value(), err
}

func (s *Badger) Delete(k []byte) error {
	return s.db.Delete(k)
}

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

func (s *Badger) Close() error {
	return s.db.Close()
}
