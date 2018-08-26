package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/rs/cors"
)

type Config struct {
	HTTP struct {
		Listen string
	}

	Postgres struct {
		User     string
		Password string
		DBName   string
		Host     string
		Port     int
	}

	LevelDB struct {
		Dir string
	}
}

var (
	configFile = flag.String("config", "config.toml", "configuration file")
)

func main() {
	flag.Parse()

	var config Config
	meta, err := toml.DecodeFile(*configFile, &config)
	if err != nil {
		log.Fatal(err)
	}
	if len(meta.Undecoded()) > 0 {
		log.Printf("unused keys: %v\n", meta.Undecoded())
	}

	if config.HTTP.Listen == "" {
		config.HTTP.Listen = ":8000"
	}
	if config.LevelDB.Dir == "" {
		config.LevelDB.Dir = "data"
	}

	var store Store

	ldb, err := NewLevelDB(config.LevelDB.Dir)
	if err != nil {
		log.Fatal(err)
	}
	defer ldb.Close()

	store = ldb

	server := NewServer(store)

	log.Printf("Starting server on %s\n", config.HTTP.Listen)
	log.Fatal(http.ListenAndServe(config.HTTP.Listen, cors.Default().Handler(server)))
}
