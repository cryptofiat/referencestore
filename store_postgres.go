package main

import (
	"github.com/jackc/pgx"
)

var _ Store = &PostgresDB{}

type PostgresDB struct {
	pool *pgx.ConnPool
}

func NewPostgresDB(params string) (*PostgresDB, error) {
	conf, err := pgx.ParseConnectionString(params)
	if err != nil {
		return nil, err
	}

	poolconf := pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: 5,
	}

	pool, err := pgx.NewConnPool(poolconf)
	if err != nil {
		return nil, err
	}

	store := &PostgresDB{pool}
	return store, store.init()
}

func (store *PostgresDB) do(fn func(conn *pgx.Conn) error) error {
	conn, err := store.pool.Acquire()
	if err != nil {
		return err
	}
	defer store.pool.Release(conn)
	return fn(conn)
}

func (store *PostgresDB) init() error {
	return store.do(func(conn *pgx.Conn) error {
		_, err := conn.Exec(`
			CREATE TABLE IF NOT EXISTS
			Info(
				Hash BYTEA  PRIMARY KEY  NOT NULL,
				Data BYTEA               NOT NULL
			)
		`)
		return err
	})
}

func (store *PostgresDB) Put(txhash Hash, data []byte) error {
	return store.do(func(conn *pgx.Conn) error {
		_, err := conn.Exec(`
			INSERT 
			INTO   Info(Hash, Data)
			VALUES ($1, $2)`, txhash[:], data)
		return err
	})
}

func (store *PostgresDB) Get(txhash Hash) ([]byte, error) {
	var data []byte
	err := store.do(func(conn *pgx.Conn) error {
		row := conn.QueryRow(`
			SELECT Data
			FROM Info
			WHERE Hash = $1`, txhash[:])
		return row.Scan(&data)
	})
	return data, err
}
