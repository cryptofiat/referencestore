package main

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

type LevelDB struct {
	db *leveldb.DB
}

func NewLevelDB(path string) (*LevelDB, error) {
	db, err := leveldb.OpenFile(*dbfile, nil)
	return &LevelDB{db}, err
}

func (store *LevelDB) Close() error {
	return store.db.Close()
}

func (store *LevelDB) Get(txhash Hash) ([]byte, error) {
	return store.db.Get(txhash[:], nil)
}

func (store *LevelDB) Put(txhash Hash, data []byte) error {
	tx, err := store.db.OpenTransaction()
	if err != nil {
		return fmt.Errorf("Problem starting DB transaction: %v", err)
	}
	defer tx.Discard()

	_, err = tx.Get(txhash[:], nil)
	if err == nil {
		return ErrExists
	} else if err != leveldb.ErrNotFound {
		return err
	}

	if err := tx.Put(txhash[:], data, nil); err != nil {
		return err
	}

	return tx.Commit()
}
