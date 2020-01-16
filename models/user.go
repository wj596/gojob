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
	"sort"
	"strings"

	"gojob/util/byteutil"
	"gojob/util/dateutil"
	"gojob/util/logs"
	"gojob/util/stringutil"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
)

type User struct {
	Id         uint64 `json:"-"`    // 主键
	IdStr      string `json:"id"`   // 主键
	Name       string `json:"name"` // 用户名称
	Password   string `json:"password"`
	UpdateTime int64  `json:"updateTime"` // 更新时间
	Email      string `json:"email"`
}

type UserSortableList []*User

func (ls UserSortableList) Len() int {
	return len(ls)
}

//Less()
func (ls UserSortableList) Less(i, j int) bool {
	return ls[i].UpdateTime > ls[j].UpdateTime
}

//Swap()
func (ls UserSortableList) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}

func CreateDefaultUserIfNecessary() {
	err := GetBoltDB().Update(func(tx *bolt.Tx) error {
		bt := tx.Bucket(userBucket)
		if nil == bt {
			admin := User{
				Id:         1,
				Name:       "admin",
				Password:   "123456",
				UpdateTime: dateutil.NowMillisecond(),
			}

			bt, err := tx.CreateBucket(userBucket)
			if err != nil {
				return err
			}
			bs, err := msgpack.Marshal(admin)
			if err != nil {
				return err
			}
			logs.Infof("创建默认账户admin")
			return bt.Put(byteutil.Uint64ToBytes(admin.Id), bs)
		}
		return nil
	})
	if nil != err {
		log.Panicf("创建默认账户失败：%s \n", err.Error())
	}
}

func SaveUser(entity *User) error {
	err := GetBoltDB().Update(func(tx *bolt.Tx) error {
		bt := tx.Bucket(userBucket)
		bs, err := msgpack.Marshal(entity)
		if err != nil {
			return err
		}
		return bt.Put(byteutil.Uint64ToBytes(entity.Id), bs)
	})
	return err
}

func BatchSaveUser(entities []*User) error {
	err := GetBoltDB().Batch(func(tx *bolt.Tx) error {
		bt := tx.Bucket(userBucket)
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

func DeleteUser(id uint64) error {
	err := GetBoltDB().Update(func(tx *bolt.Tx) error {
		bt := tx.Bucket(userBucket)
		return bt.Delete(byteutil.Uint64ToBytes(id))
	})
	return err
}

func ForEachUser() ([]*User, error) {
	var list []*User
	err := GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var entity = new(User)
			if err := msgpack.Unmarshal(v, entity); err == nil {
				list = append(list, entity)
			}
		}
		return nil
	})
	return list, err
}

func GetUser(name string) (*User, error) {
	var entity = new(User)
	var find bool
	err := GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			if err := msgpack.Unmarshal(v, entity); err == nil {
				if entity.Name == name {
					find = true
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !find {
		return nil, errors.Errorf("Name Not Found")
	}

	return entity, nil
}

func SelectUserList(ps *Condition) []*User {
	var list []*User
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var entity = new(User)
			if err := msgpack.Unmarshal(v, entity); err == nil {
				name := ps.GetStringParam("name")
				if name != "" && !strings.Contains(entity.Name, name) {
					continue
				}
				hasEmail := ps.GetStringParam("hasEmail")
				if hasEmail != "" && entity.Email == "" {
					continue
				}
				entity.IdStr = stringutil.UintToStr(entity.Id)
				list = append(list, entity)
			}
		}
		return nil
	})
	sortables := UserSortableList(list)
	sort.Sort(sortables)
	return sortables
}
