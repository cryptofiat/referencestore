package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"testing"
)

var testPostgres = flag.String("postgres-test-db", "", "postgres test database")

func TestLevelDB(t *testing.T) {
	store, cleanup := newTestLevelDB(t)
	defer cleanup()
	testStore(t, store)
}

func TestPostgresDB(t *testing.T) {
	store, cleanup := newTestPostgres(t)
	defer cleanup()
	testStore(t, store)
}

func testStore(t *testing.T, store Store) {
	if err := store.Put(Hash{1}, []byte("hello")); err != nil {
		t.Fatalf("insert 1: %v", err)
	}
	if err := store.Put(Hash{2}, []byte("world")); err != nil {
		t.Fatalf("insert 2: %v", err)
	}

	if err := store.Put(Hash{1}, []byte("should not insert")); err != ErrExists {
		t.Fatalf("duplicate insert: %v", err)
	}

	{
		data, err := store.Get(Hash{1})
		if err != nil {
			t.Fatalf("get 1: %v", err)
		}
		if !bytes.Equal(data, []byte("hello")) {
			t.Fatalf("get 1: %v", data)
		}
	}

	{
		data, err := store.Get(Hash{2})
		if err != nil {
			t.Fatalf("get 2: %v", err)
		}
		if !bytes.Equal(data, []byte("world")) {
			t.Fatalf("get 2: %v", data)
		}
	}

	{
		_, err := store.Get(Hash{0})
		if err != ErrNotFound {
			t.Fatalf("invalid get 0: %v", err)
		}
	}

	{
		found := make(map[Hash]bool)
		err := store.List(func(h Hash, data []byte) error {
			found[h] = true
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		if !found[Hash{1}] || !found[Hash{2}] {
			t.Fatalf("only got: %v", found)
		}
	}
}

func TestMigration(t *testing.T) {
	pgdb, cleanup := newTestPostgres(t)
	defer cleanup()
	ldb, cleanup := newTestLevelDB(t)
	defer cleanup()

	expected := map[Hash][]byte{
		Hash{1}: []byte("hello"),
		Hash{2}: []byte("hello"),
	}

	for key, value := range expected {
		if err := ldb.Put(key, value); err != nil {
			t.Fatalf("ldb insert %v: %v", key, value)
		}
	}

	if err := pgdb.MigrateFrom(ldb); err != nil {
		t.Fatalf("migration: %v", err)
	}

	err := pgdb.List(func(h Hash, data []byte) error {
		expdata, ok := expected[h]
		if !ok {
			t.Fatalf("item %v not found", h)
		}
		delete(expected, h)

		if !bytes.Equal(expdata, data) {
			t.Fatalf("different data %v: %v\n%v", h, data, expdata)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("list %v:", err)
	}
	if len(expected) != 0 {
		t.Fatalf("some items missing: %v", expected)
	}
}

func newTestLevelDB(t *testing.T) (store *LevelDB, cleanup func()) {
	tempdir, err := ioutil.TempDir("", "transfer-info-leveldb")
	if err != nil {
		t.Fatal(err)
	}
	ldb, err := NewLevelDB(tempdir)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	return ldb, func() {
		defer os.RemoveAll(tempdir)

		err := ldb.Close()
		if err != nil {
			t.Fatalf("close: %v", err)
		}
	}
}

func newTestPostgres(t *testing.T) (store *PostgresDB, cleanup func()) {
	if *testPostgres == "" {
		t.Skip(`postgres flag missing, example:` + "\n" + `-postgres-test-db="user=transferinfo password=q1w2e3r4 dbname=transfer-info"`)
	}

	pgdb, err := NewPostgresDB(*testPostgres)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	return pgdb, func() {
		if err := pgdb.DESTROY_INFO(); err != nil {
			t.Fatal(err)
		}

		if err := pgdb.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
