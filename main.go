package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	storedir = flag.String("dir", ".store", "data store directory")
	listen   = flag.String("listen", ":8000", "listening address")
)

func main() {
	flag.Parse()

	store := NewStore()

	// Where's the right place to initiate DB vars?
	db, err := leveldb.OpenFile(".refstore/refs.db", nil)

	log.Printf("Starting server on %s\n", *listen)
	log.Fatal(http.ListenAndServe(*listen, store))

	defer db.Close()
}

const HashLength = 32

type Hash [HashLength]byte

func ParseHash(s string) (r Hash, err error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return r, err
	}
	copy(r[:], data)
	return r, nil
}
func (ref *Hash) Hex() string {
	return hex.EncodeToString(ref[:])
}

type Blob []byte

func GenerateHash() (ref Hash) {
	_, err := io.ReadFull(rand.Reader, ref[:])
	if err != nil {
		panic(err)
	}
	return
}

type Store struct {
	mu    sync.Mutex
	blobs map[Hash]Blob
}

func NewStore() *Store {
	return &Store{
		mu:    sync.Mutex{},
		blobs: make(map[Hash]Blob),
	}
}

func (store *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	case http.MethodGet:
		hex := r.URL.Path[1:]

		txhash, err := ParseHash(hex)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid transaction hash %s", hex), http.StatusBadRequest)
			return
		}

		store.mu.Lock()
		blob, ok := store.blobs[txhash]
		store.mu.Unlock()

		//alternative db implementation
		//txhash is effectively [Hashlength]byte, why doesn't it convert automagically?
		dblob, err := db.Get([]byte(txhash), nil)

		if !ok {
			http.Error(w, fmt.Sprintf("Transaction %s does not contain data", hex), http.StatusNotFound)
			return
		}

//		if _, err := w.Write(blob); err != nil {
		if _, err := w.Write(dblob); err != nil {
			log.Println(err)
		}
	case http.MethodPost:
		hex := r.URL.Path[1:]

		txhash, err := ParseHash(hex)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid transaction hash %s", hex), http.StatusBadRequest)
			return
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to receive data.", http.StatusBadRequest)
			return
		}

		// Should the implementation support for both sender and receiver being able to receive the transaction?
		// e.g /txhash/senderaccount encrypted for sender and /txhash/receiveraccount encrypted for receiver (maybe not that important)
		// Probably need some size limit, eg 10MB per post to avoid becoming WeTransfer, maybe some other validation to avoid abuse
		// should there be any validation of encryption/signatures - maybe a job that runs once a day and removes all keys,
		// for  which there is no transaction on blockchain.
		// Eventually someone would be paying for the server, so maybe have throttling for IPs and whitelist paid users (gateways)
		store.mu.Lock()
		_, exists := store.blobs[txhash]
		if !exists {
			store.blobs[txhash] = data
			//alternative db installation
			err = db.Put([]byte(txhash), []byte(data), nil)
			if err { http.Error(w, "Some DB error", http.StatusConflict) }
		}
		store.mu.Unlock()

		if exists {
			http.Error(w, "Transaction hash already stores data", http.StatusConflict)
		}

	}


}
