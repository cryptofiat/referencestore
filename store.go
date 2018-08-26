package main

import (
	"encoding/hex"
	"errors"
)

const (
	HashLength  = 32
	MaxPostSize = 1000000 // 1MB
)

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

type Store interface {
	List(fn func(Hash, []byte) error) error
	Put(h Hash, data []byte) error
	Get(h Hash) ([]byte, error)
}

var (
	ErrInvalidHash = errors.New("invalid hash")
	ErrNotFound    = errors.New("reference-info not found")
	ErrExists      = errors.New("reference-info already exists")
)
