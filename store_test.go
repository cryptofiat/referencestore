package main

import (
	"bytes"
	"os"
	"testing"
)

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
}

func TestLevelDB(t *testing.T) {
	defer os.RemoveAll("temp")

	store, err := NewLevelDB("temp")
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Fatalf("close: %v", err)
		}
	}()

	testStore(t, store)
}
