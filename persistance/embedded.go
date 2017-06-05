package persistance

import (
	"bytes"
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

var (
	ErrTableNotFound = errors.New("table not found")
	ErrKeyNotFound   = errors.New("key not found")
)

type EmbeddedDB struct {
	conn *bolt.DB
}

func NewEmbeddedDB(filePath string) *EmbeddedDB {
	d, err := bolt.Open(filePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil
	}

	return &EmbeddedDB{conn: d}
}

func (d *EmbeddedDB) Get(tableName string, key string) ([]byte, error) {
	tx, err := d.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte(tableName))
	if b == nil {
		return nil, ErrTableNotFound
	}

	v := b.Get([]byte(key))
	if v == nil {
		return nil, ErrKeyNotFound
	}

	return v, nil
}

func (d *EmbeddedDB) Set(tableName string, key string, value []byte) error {
	return d.conn.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(key))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
	})
}

func (d *EmbeddedDB) List(tableName string) ([]string, error) {
	tx, err := d.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte(tableName))
	if b == nil {
		return nil, ErrTableNotFound
	}

	keys := make([]string, 0)
	c := b.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		keys = append(keys, string(k))
	}

	return keys, nil
}

func (d *EmbeddedDB) ListPrefix(tableName string, prefix string) ([]string, error) {
	tx, err := d.conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte(tableName))
	if b == nil {
		return nil, ErrTableNotFound
	}

	keys := make([]string, 0)
	prefixByte := []byte(prefix)
	c := b.Cursor()
	for k, _ := c.First(); k != nil && bytes.HasPrefix(k, prefixByte); k, _ = c.Next() {
		keys = append(keys, string(k))
	}

	return keys, nil
}
