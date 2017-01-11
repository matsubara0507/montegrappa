package db

import (
	"errors"
	"github.com/boltdb/bolt"
	"golang.org/x/oauth2"
	"time"
)

var conn *bolt.DB
var (
	ErrNotExistToken = errors.New("Not exist access token")
	ErrNotOpenDB     = errors.New("Not open database")
)

var (
	BucketAuthenticationInformation = []byte("SlackAuth")
)

func Open(name string) error {
	d, err := bolt.Open(name, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	conn = d
	return nil
}

func GetToken() (string, error) {
	token, err := getStringValue(BucketAuthenticationInformation, []byte("access_token"))
	if token == "" && err == nil {
		return "", ErrNotExistToken
	}

	return token, err
}

func GetBotToken() (string, error) {
	token, err := getStringValue(BucketAuthenticationInformation, []byte("bot_access_token"))
	if token == "" && err == nil {
		return "", ErrNotExistToken
	}

	return token, err
}

func WriteToken(token *oauth2.Token) error {
	if conn == nil {
		return ErrNotOpenDB
	}

	botToken := token.Extra("bot").(map[string]interface{})
	return conn.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(BucketAuthenticationInformation)
		if err != nil {
			return err
		}
		err = b.Put([]byte("access_token"), []byte(token.AccessToken))
		if err != nil {
			return err
		}
		err = b.Put([]byte("bot_user_id"), []byte(botToken["bot_user_id"].(string)))
		if err != nil {
			return err
		}
		err = b.Put([]byte("bot_access_token"), []byte(botToken["bot_access_token"].(string)))
		if err != nil {
			return err
		}

		return nil
	})
}

func getStringValue(bucket, key []byte) (string, error) {
	if conn == nil {
		return "", ErrNotOpenDB
	}

	tx, err := conn.Begin(false)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	b := tx.Bucket(bucket)
	if b == nil {
		return "", ErrNotExistToken
	}
	v := b.Get(key)

	return string(v), nil
}
