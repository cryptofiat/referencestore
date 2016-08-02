package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	dbfile = flag.String("data", "data", "database directory")
	listen = flag.String("listen", ":8000", "listening address")
)

func main() {
	flag.Parse()

	db, err := leveldb.OpenFile(*dbfile, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	log.Printf("Starting server on %s\n", *listen)
	log.Fatal(http.ListenAndServe(*listen, store))
}

const HashLength = 32
const MaxPostSize = 1000000 // 1MB

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

type Store struct {
	DB *leveldb.DB
}

func NewStore(db *leveldb.DB) *Store {
	return &Store{db}
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

		blob, err := store.DB.Get(txhash[:], nil)
		if err != nil {
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

		data, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, MaxPostSize))
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

		tx, err := store.DB.OpenTransaction()
		if err != nil {
			http.Error(w, "Problem starting DB transaction.", http.StatusInternalServerError)
			return
		}
		defer tx.Discard()

		if _, err := tx.Get(txhash[:], nil); err != leveldb.ErrNotFound {
			http.Error(w, "Transaction hash already stores data", http.StatusConflict)
			return
		}

		if err := tx.Put(txhash[:], data, nil); err != nil {
			log.Println(err)
			http.Error(w, "Error saving to DB.", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Println(err)
			http.Error(w, "Error committing to DB.", http.StatusInternalServerError)
		}
	}

}
