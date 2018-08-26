package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"

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

func (config *Config) PostgresDSN() string {
	dsn := []string{}
	include := func(name, value string) {
		if value != "" {
			dsn = append(dsn, name+"="+value)
		}
	}
	p := &config.Postgres
	include("user", p.User)
	include("password", p.Password)
	include("dbname", p.DBName)
	include("host", p.Host)
	if p.Port == 0 {
		include("port", "5432")
	} else {
		include("port", strconv.Itoa(p.Port))
	}

	return strings.Join(dsn, " ")
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

	if flag.Arg(0) == "migrate" {
		log.Println("Migrating")
		log.Println("LevelDB: ", config.LevelDB.Dir)
		ldb, err := NewLevelDB(config.LevelDB.Dir)
		if err != nil {
			log.Fatal(err)
		}
		defer ldb.Close()

		dsn := config.PostgresDSN()
		log.Println("Postgres: ", dsn)
		pgdb, err := NewPostgresDB(dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer pgdb.Close()

		count, err := pgdb.MigrateFrom(ldb)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Migration complete", count)
		return
	}

	var store Store

	if dsn := config.PostgresDSN(); dsn != "" {
		log.Println("Postgres: ", dsn)
		pgdb, err := NewPostgresDB(dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer pgdb.Close()
		store = pgdb
	} else {
		log.Println("LevelDB: ", config.LevelDB.Dir)
		ldb, err := NewLevelDB(config.LevelDB.Dir)
		if err != nil {
			log.Fatal(err)
		}
		defer ldb.Close()
		store = ldb
	}

	server := NewServer(store)

	log.Printf("Starting server on %s\n", config.HTTP.Listen)
	log.Fatal(http.ListenAndServe(config.HTTP.Listen, cors.Default().Handler(server)))
}
