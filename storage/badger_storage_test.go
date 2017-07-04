package storage

import (
	"os"
	"testing"

	"github.com/go-ego/gwk/utils"
)

func TestOpenOrCreateBad(t *testing.T) {
	db, err := openBadgerStorage("bad_test")
	utils.Expect(t, "<nil>", err)
	db.Close()

	db, err = openBadgerStorage("bad_test")
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
	// os.Remove("bad_test")
	os.RemoveAll("bad_test")
}
