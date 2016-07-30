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

type Reference [32]byte

func ParseReference(s string) (r Reference, err error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return r, err
	}
	copy(r[:], data)
	return r, nil
}
func (ref *Reference) Hex() string {
	return hex.EncodeToString(ref[:])
}

type Blob []byte

func GenerateReference() (ref Reference) {
	_, err := io.ReadFull(rand.Reader, ref[:])
	if err != nil {
		panic(err)
	}
	return
}

type Store struct {
	mu    sync.Mutex
	blobs map[Reference]Blob
}

func NewStore() *Store {
	return &Store{
		mu:    sync.Mutex{},
		blobs: make(map[Reference]Blob),
	}
}

func (store *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	case http.MethodGet:
		hx := r.URL.Path[1:]

		ref, err := ParseReference(hx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid reference %s", hx), http.StatusBadRequest)
			return
		}

		store.mu.Lock()
		blob, ok := store.blobs[ref]
		store.mu.Unlock()

		if !ok {
			http.Error(w, fmt.Sprintf("Reference %s missing", hx), http.StatusNotFound)
			return
		}

		if _, err := w.Write(blob); err != nil {
			log.Println(err)
		}
	case http.MethodPost:
		ref := GenerateReference()

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to receive data.", http.StatusBadRequest)
			return
		}

		store.mu.Lock()
		defer store.mu.Unlock()
		if _, err := w.Write([]byte(ref.Hex())); err == nil {
			store.blobs[ref] = data
		}
	}
}
