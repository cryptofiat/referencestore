package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Server struct {
	Store Store
}

func NewServer(store Store) *Server {
	return &Server{store}
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		blob, err := server.Store.Get(txhash)
		if err != nil {
			if err == ErrNotFound {
				http.Error(w, fmt.Sprintf("Transaction %s does not contain data", hex), http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Transaction %s does not contain data: %v", hex, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Cache-Control", "max-age=3600")

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

		err = server.Store.Put(txhash, data)
		if err != nil {
			if err == ErrExists {
				http.Error(w, fmt.Sprintf("Transaction hash already stores data %s", hex), http.StatusConflict)
			}
			http.Error(w, fmt.Sprintf("Problem saving: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
