package storage

import (
	"log"
	"os"
	"testing"

	"github.com/vcaesar/tt"
)

var TestDBName = "db_test"

func TestBadger(t *testing.T) {
	db, err := OpenBadger(TestDBName)
	log.Println("TestBadger...")
	DBTest(t, db)

	tt.Expect(t, "<nil>", err)
	db.Close()
}

func TestLdb(t *testing.T) {
	db, err := OpenLeveldb(TestDBName)
	log.Println("TestLdb...")
	DBTest(t, db)

	tt.Expect(t, "<nil>", err)
	db.Close()
}

func TestBolt(t *testing.T) {
	db, err := OpenBolt("db_test")
	log.Println("TestBolt...")
	DBTest(t, db)

	tt.Expect(t, "<nil>", err)
	db.Close()
}

func DBTest(t *testing.T, db Storage) {
	err := db.Set([]byte("key1"), []byte("value1"))
	tt.Expect(t, "<nil>", err)

	has, err := db.Has([]byte("key1"))
	tt.Equal(t, nil, err)
	if err == nil {
		tt.Equal(t, true, has)
	}

	buf := make([]byte, 100)
	buf, err = db.Get([]byte("key1"))
	tt.Expect(t, "<nil>", err)
	tt.Expect(t, "value1", string(buf))

	walFile := db.WALName()
	// db.Close()
	os.Remove(walFile)
	os.RemoveAll(TestDBName)
}
