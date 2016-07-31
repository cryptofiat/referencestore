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
)

var (
	storedir = flag.String("dir", ".store", "data store directory")
	listen   = flag.String("listen", ":8000", "listening address")
)

func main() {
	flag.Parse()

	store := NewStore()

	log.Printf("Starting server on %s\n", *listen)
	log.Fatal(http.ListenAndServe(*listen, store))
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

		if !ok {
			http.Error(w, fmt.Sprintf("Transaction %s does not contain data", hex), http.StatusNotFound)
			return
		}

		if _, err := w.Write(blob); err != nil {
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

		store.mu.Lock()
		_, exists := store.blobs[txhash]
		if !exists {
			store.blobs[txhash] = data
		}
		store.mu.Unlock()

		if exists {
			http.Error(w, "Transaction hash already stores data", http.StatusConflict)
		}
	}
}
