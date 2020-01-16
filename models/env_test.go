package models

import (
	"github.com/boltdb/bolt"
	"gojob/util/byteutil"
	"testing"
)

func TestSetLastTcpPort(t *testing.T) {
	InitBoltDB("D:\\test")
	GetBoltDB().Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("env"))
		bucket.Put(byteutil.Uint64ToBytes(uint64(1)), byteutil.Uint64ToBytes(uint64(8888)))
		bucket.Put(byteutil.Uint64ToBytes(uint64(2)), []byte("test"))
		return err
	})
}

func TestGetLastTcpPort(t *testing.T) {
	InitBoltDB("D:\\test")
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("env"))
		v1 := bucket.Get(byteutil.Uint64ToBytes(uint64(1)))
		println(byteutil.BytesToUint64(v1))
		v2 := bucket.Get(byteutil.Uint64ToBytes(uint64(2)))
		println(string(v2))
		return nil
	})
}
