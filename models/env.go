/*
 * Copyright 2020-2021 the original author(https://github.com/wj596)
 *
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * </p>
 */
package models

import (
	"gojob/util/byteutil"

	"github.com/boltdb/bolt"
)

var fixSnapshotVersionId = byteutil.Uint64ToBytes(uint64(1))
var fixRaftFirstStartId = byteutil.Uint64ToBytes(uint64(2))
var fixLastTcpPortId = byteutil.Uint64ToBytes(uint64(3))
var fixLastNodeNameId = byteutil.Uint64ToBytes(uint64(4))

func UpdateSnapshotVersion(snapshotVersion uint64) {
	GetBoltDB().Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(envBucket)
		return bucket.Put(fixSnapshotVersionId, byteutil.Uint64ToBytes(snapshotVersion))
	})
}

func GetSnapshotVersion() uint64 {
	var snapshotVersion uint64
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(envBucket)
		val := bucket.Get(fixSnapshotVersionId)
		if val != nil {
			snapshotVersion = byteutil.BytesToUint64(val)
		}
		return nil
	})
	return snapshotVersion
}

func IsRaftFirstStart() bool {
	firstStart := true
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(envBucket)
		val := bucket.Get(fixRaftFirstStartId)
		if val != nil {
			temp := byteutil.BytesToUint64(val)
			if temp != 0 {
				firstStart = false
			}
		}
		return nil
	})
	return firstStart
}

func NegationRaftFirstStart() {
	GetBoltDB().Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(envBucket)
		return bucket.Put(fixRaftFirstStartId, byteutil.Uint64ToBytes(1))
	})
}

func GetLastTcpPort() int {
	var tcpPort uint64
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(envBucket)
		val := bucket.Get(fixLastTcpPortId)
		if val != nil {
			tcpPort = byteutil.BytesToUint64(val)
		}
		return nil
	})
	return int(tcpPort)
}

func SetLastTcpPort(tcpPort int) {
	GetBoltDB().Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(envBucket)
		return bucket.Put(fixLastTcpPortId, byteutil.Uint64ToBytes(uint64(tcpPort)))
	})
}

func GetLastNodeName() string {
	var nodeName string
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(envBucket)
		val := bucket.Get(fixLastNodeNameId)
		if val != nil {
			nodeName = string(val)
		}
		return nil
	})
	return nodeName
}

func SetLastNodeName(nodeName string) {
	GetBoltDB().Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(envBucket)
		return bucket.Put(fixLastNodeNameId, []byte(nodeName))
	})
}
