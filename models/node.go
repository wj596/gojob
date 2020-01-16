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
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
)

type Node struct {
	Name       string // 节点名称
	HttpAddr   string // Http地址
	TcpAddr    string // Tcp地址
	MachineNum uint16 // 节点序号
}

func InsertNode(node *Node) error {
	err := GetBoltDB().Update(func(tx *bolt.Tx) error {
		var maxMachineNum uint16
		bt := tx.Bucket(nodeBucket)
		cursor := bt.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var temp = new(Node)
			err := msgpack.Unmarshal(v, temp)
			if err == nil && temp.MachineNum > maxMachineNum {
				maxMachineNum = temp.MachineNum
			}
		}

		node.MachineNum = maxMachineNum + 1
		bs, err := msgpack.Marshal(node)
		if err != nil {
			return err
		}
		return bt.Put([]byte(node.Name), bs)
	})

	return err
}

func UpdateNode(node *Node) error {
	err := GetBoltDB().Update(func(tx *bolt.Tx) error {
		bt := tx.Bucket(nodeBucket)
		bs, err := msgpack.Marshal(node)
		if err != nil {
			return err
		}
		return bt.Put([]byte(node.Name), bs)
	})
	return err
}

func BatchSaveNode(nodes []*Node) error {
	err := GetBoltDB().Batch(func(tx *bolt.Tx) error {
		bt := tx.Bucket(nodeBucket)
		for _, node := range nodes {
			bs, err := msgpack.Marshal(node)
			if err != nil {
				continue
			}
			bt.Put([]byte(node.Name), bs)
		}
		return nil
	})
	return err
}

func GetNode(nodeName string) (*Node, error) {
	var val []byte
	err := GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(nodeBucket)
		val = bucket.Get([]byte(nodeName))
		if val == nil {
			return errors.Errorf("Key Not Found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var entity = new(Node)
	err = msgpack.Unmarshal(val, entity)
	return entity, err
}

func GetNodeByTcpAddr(tcpAddr string) (*Node, error) {
	var find bool
	var entity = new(Node)
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(nodeBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			if err := msgpack.Unmarshal(v, entity); err == nil {
				if entity.TcpAddr == tcpAddr {
					find = true
					break
				}
			}
		}
		return nil
	})

	if !find {
		return nil, errors.Errorf("Not Found.")
	}

	return entity, nil
}

func ForEachNode() ([]*Node, error) {
	var list []*Node
	err := GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(nodeBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var entity = new(Node)
			if err := msgpack.Unmarshal(v, entity); err == nil {
				list = append(list, entity)
			}
		}
		return nil
	})
	return list, err
}
