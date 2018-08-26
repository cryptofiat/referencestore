package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/rs/cors"
)

var (
	dbfile = flag.String("data", "data", "database directory")
	listen = flag.String("listen", ":8000", "listening address")
)

func main() {
	flag.Parse()

	store, err := NewLevelDB(*dbfile)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	server := NewServer(store)

	log.Printf("Starting server on %s\n", *listen)
	log.Fatal(http.ListenAndServe(*listen, cors.Default().Handler(server)))
}
