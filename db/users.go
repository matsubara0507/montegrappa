package db

import (
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
)

var (
	BucketUsers = []byte("SlackUsers")
)

var (
	ErrUserNotFound = errors.New("User not found")
)

type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func GetUser(userId string) (*User, error) {
	if conn == nil {
		return nil, ErrNotOpenDB
	}

	tx, err := conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(BucketUsers)
	if b == nil {
		return nil, ErrNotExistBucket
	}
	s := b.Get([]byte(userId))
	if s == nil {
		return nil, ErrUserNotFound
	}
	u := new(User)
	err = json.Unmarshal(s, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func WriteUser(user *User) error {
	if conn == nil {
		return ErrNotOpenDB
	}

	return conn.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(BucketUsers)
		if err != nil {
			return err
		}
		buf, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return b.Put([]byte(user.Id), buf)
	})
}
