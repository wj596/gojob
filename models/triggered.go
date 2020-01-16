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
	"time"

	"gojob/util/byteutil"
	"gojob/util/stringutil"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
)

// 实时触发信息
type Triggered struct {
	Id       uint64 // 主键
	IdStr    string // 主键
	Times    int64  // 调度次数
	PrevTime int64  // 上次触发时间
	NextTime int64  // 下次触发时间
}

func SaveTriggered(entity *Triggered) error {
	err := GetBoltDB().Update(func(tx *bolt.Tx) error {
		bt := tx.Bucket(triggeredBucket)
		bs, err := msgpack.Marshal(entity)
		if err != nil {
			return err
		}
		return bt.Put(byteutil.Uint64ToBytes(entity.Id), bs)
	})
	return err
}

func BatchSaveTriggered(entities []*Triggered) error {
	err := GetBoltDB().Batch(func(tx *bolt.Tx) error {
		bt := tx.Bucket(triggeredBucket)
		for _, entity := range entities {
			bs, err := msgpack.Marshal(entity)
			if err != nil {
				continue
			}
			bt.Put(byteutil.Uint64ToBytes(entity.Id), bs)
		}
		return nil
	})
	return err
}

func GetTriggered(id uint64) (*Triggered, error) {
	var val []byte
	err := GetBoltDB().View(func(tx *bolt.Tx) error {
		bt := tx.Bucket(triggeredBucket)
		val = bt.Get(byteutil.Uint64ToBytes(id))
		if val == nil {
			return errors.Errorf("Key Not Found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var entity = new(Triggered)
	err = msgpack.Unmarshal(val, entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func GetTriggeredAmount() int64 {
	var amount int64
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(triggeredBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var entity = new(Triggered)
			if err := msgpack.Unmarshal(v, entity); err == nil {
				amount = amount + entity.Times
			}
		}
		return nil
	})
	return amount
}

func ForEachTriggered() ([]*Triggered, error) {
	list := make([]*Triggered, 0)
	err := GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(triggeredBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var entity = new(Triggered)
			if err := msgpack.Unmarshal(v, entity); err == nil {
				entity.IdStr = stringutil.UintToStr(entity.Id)
				list = append(list, entity)
			}
		}
		return nil
	})
	return list, err
}

func SelectMisfireList() []*Triggered {
	current := time.Now().Unix()
	list := make([]*Triggered, 0)
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(triggeredBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			entity := new(Triggered)
			if err := msgpack.Unmarshal(v, entity); err == nil {
				if entity.PrevTime >= entity.NextTime {
					continue
				}
				job, err := GetJob(entity.Id)
				if nil != err {
					continue
				}
				if job.Status == JobStatusPause {
					continue
				}
				if job.MisfireThreshold == 0 {
					continue
				}
				if current > entity.NextTime {
					diff := current - entity.NextTime
					if diff <= job.MisfireThreshold {
						list = append(list, entity)
					}
				}
			}
		}
		return nil
	})
	return list
}
