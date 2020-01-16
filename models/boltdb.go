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
	"log"
	"path/filepath"

	"gojob/util/fileutil"
	"gojob/util/logs"

	"github.com/boltdb/bolt"
)

const (
	boltPath     = "meta"
	boltFileName = "job.db"
	boltFileMode = 0600
)

var (
	jobBucket         = []byte("job")
	triggeredBucket   = []byte("triggered")
	nodeBucket        = []byte("node")
	userBucket        = []byte("user")
	alarmConfigBucket = []byte("alarmConfig")
	envBucket         = []byte("env")
	boltDB            *bolt.DB
)

func InitBoltDB(dataStorePath string) {
	blotStorePath := filepath.Join(dataStorePath, boltPath)
	if err := fileutil.MkdirIfNecessary(blotStorePath); err != nil {
		logs.Errorf("blot存储目录:%s，创建失败. \n", blotStorePath)
		log.Panicf("blot存储目录:%s，创建失败. \n", blotStorePath)
	}
	boltFilePath := filepath.Join(blotStorePath, boltFileName)
	handle, err := bolt.Open(boltFilePath, boltFileMode, bolt.DefaultOptions)
	if err != nil {
		logs.Errorf("本地存储引擎blotDB创建失败：%s \n", err.Error())
		log.Panicf("本地存储引擎blotDB创建失败：%s \n", err.Error())
	}
	boltDB = handle
	boltDB.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists(jobBucket)
		tx.CreateBucketIfNotExists(triggeredBucket)
		tx.CreateBucketIfNotExists(nodeBucket)
		tx.CreateBucketIfNotExists(alarmConfigBucket)
		tx.CreateBucketIfNotExists(envBucket)
		return nil
	})
	logs.Info("本地存储引擎boltDB创建成功")
	log.Print("本地存储引擎boltDB创建成功")
}

func RestBucket() {
	boltDB.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(jobBucket)
		tx.CreateBucketIfNotExists(jobBucket)

		tx.DeleteBucket(triggeredBucket)
		tx.CreateBucketIfNotExists(triggeredBucket)

		tx.DeleteBucket(nodeBucket)
		tx.CreateBucketIfNotExists(nodeBucket)

		tx.DeleteBucket(userBucket)
		tx.CreateBucketIfNotExists(userBucket)

		tx.DeleteBucket(alarmConfigBucket)
		tx.CreateBucketIfNotExists(alarmConfigBucket)
		return nil
	})
}

func CloseBoltDB() {
	if nil != boltDB {
		boltDB.Close()
	}
}

func GetBoltDB() *bolt.DB {
	return boltDB
}
