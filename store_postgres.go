package main

import (
	"strings"

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

func (store *PostgresDB) Close() error {
	store.pool.Close()
	return nil
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

func (store *PostgresDB) DESTROY_INFO() error {
	return store.do(func(conn *pgx.Conn) error {
		_, err := conn.Exec(`DROP TABLE Info`)
		return err
	})
}

func (store *PostgresDB) List(fn func(Hash, []byte) error) error {
	return store.do(func(conn *pgx.Conn) error {
		rows, err := conn.Query(`SELECT Hash, Data FROM Info`)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var h []byte
			var data []byte

			err := rows.Scan(&h, &data)
			if err != nil {
				return err
			}

			var txhash Hash
			if copy(txhash[:], h) != HashLength {
				return ErrInvalidHash
			}

			if err := fn(txhash, data); err != nil {
				return err
			}
		}

		return err
	})
}

func (store *PostgresDB) Put(txhash Hash, data []byte) error {
	return store.do(func(conn *pgx.Conn) error {
		_, err := conn.Exec(`
			INSERT 
			INTO   Info(Hash, Data)
			VALUES ($1, $2)`, txhash[:], data)
		if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return ErrExists
		}
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
		err := row.Scan(&data)
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	})
	return data, err
}

func (store *PostgresDB) MigrateFrom(source Store) (int, error) {
	var count int
	err := store.do(func(conn *pgx.Conn) error {
		tx, err := conn.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		err = source.List(func(txhash Hash, data []byte) error {
			count++
			_, err := tx.Exec(`INSERT INTO Info(Hash, Data) VALUES ($1, $2)`, txhash[:], data)
			return err
		})

		if err != nil {
			return err
		}

		return tx.Commit()
	})
	return count, err
}
