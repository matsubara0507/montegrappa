package db

import (
	"encoding/binary"
	"github.com/boltdb/bolt"
	"time"
)

type InstanceInfo struct {
	StartAt time.Time
	Seq     uint64
}

var (
	BucketGlobal = []byte("Global")
)

func UpdateLastStart() error {
	return conn.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(BucketGlobal)
		if err != nil {
			return err
		}
		err = b.Put([]byte("start_at"), []byte(time.Now().String()))
		if err != nil {
			return err
		}
		seq, _ := b.NextSequence()
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, seq)
		return b.Put([]byte("seq"), buf)
	})
}

func ReadInstanceInfo() (*InstanceInfo, error) {
	if conn == nil {
		return nil, ErrNotOpenDB
	}

	tx, err := conn.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket(BucketGlobal)
	s := b.Get([]byte("start_at"))
	instanceInfo := new(InstanceInfo)
	if s != nil {
		instanceInfo.StartAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", string(s))
	}
	s = b.Get([]byte("seq"))
	instanceInfo.Seq = binary.BigEndian.Uint64(s)

	return instanceInfo, nil
}
