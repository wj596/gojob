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
	"sort"
	"strings"
	"sync"

	"gojob/util/byteutil"
	"gojob/util/logs"
	"gojob/util/stringutil"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
)

const (
	// 执行节点选择策略 -- 分片
	ExecutorSelectStrategySharding = "sharding"
	// 执行节点选择策略 -- 随机
	ExecutorSelectStrategyRandom = "random"
	// 执行节点选择策略 -- 轮询
	ExecutorSelectStrategyRound = "round"
	// 执行节点选择策略 -- 加权随机
	ExecutorSelectStrategyWeightRandom = "weight_random"
	// 执行节点选择策略 -- 加权轮询
	ExecutorSelectStrategyWeightRound = "weight_round"
	// 子任务触发策略 -- 执行完毕触发
	SubJobScheduleStrategyEnd = 0
	// 子任务触发策略 -- 执行成功触发
	SubJobScheduleStrategyOk = 1
	// 子任务触发策略 -- 执行失败触发
	SubJobScheduleStrategyFail = 2
	// 作业状态 -- 正常
	JobStatusOk = 1
	// 作业状态 -- 挂起
	JobStatusPause = 0
	// 故障转移 -- 启用
	FailTakeoverEnabled = 1
	// Http签名 -- 启用
	HttpSignEnabled = 1
	// 执行节点状态 -- 可用
	ExecutorStatusOk = 1
)

// 执行节点
type Executor struct {
	Address string `json:"address"` // 执行器地址
	Weight  int    `json:"weight"`  // 权重
	Status  int    `json:"status"`  // 状态 1上线 0下线
}

// 作业
type Job struct {
	Id                     uint64      `json:"-"`                      // 主键
	IdStr                  string      `json:"id"`                     // 主键
	Name                   string      `json:"name"`                   // 任务名称
	Cron                   string      `json:"cron"`                   // cron 表达式
	Protocol               string      `json:"protocol"`               // 网络协议 http / https
	Uri                    string      `json:"uri"`                    // 任务的资源标识符
	Remark                 string      `json:"remark"`                 // 备注
	Status                 int         `json:"status"`                 // 状态 0暂停 1正常
	CreateTime             int64       `json:"createTime"`             // 创建时间
	Creator                string      `json:"creator"`                // 创建人
	PreJobId               string      `json:"preJobId"`               // 前置任务ID
	Timeout                int         `json:"timeout"`                // 任务超时时间
	RetryCount             int         `json:"retryCount"`             // 重试次数
	RetryWaitTime          int         `json:"retryWaitTime"`          // 重试间隔（秒）
	FailTakeover           int         `json:"failTakeover"`           // 故障转移 0不转移 1转移
	MisfireThreshold       int64       `json:"misfireThreshold"`       // 触发器超时时间（秒）
	ExecutorSelectStrategy string      `json:"executorSelectStrategy"` // 执行器选择策略 随机 全部 分片
	HttpParam              string      `json:"httpParam"`              // http参数
	HttpHeaderParam        string      `json:"httpHeaderParam"`        // http头参数
	HttpSign               int         `json:"httpSign"`               // http请求是否签名
	ShardingCount          int         `json:"shardingCount"`          // 分片总数
	ShardingParam          string      `json:"shardingParam"`          // 分片参数
	AlarmEmail             string      `json:"alarmEmail"`             // 告警邮箱
	SubJobScheduleStrategy int         `json:"subJobScheduleStrategy"` // 子JOB触发策略 0执行完毕触发 1执行成功触发 2执行失败触发
	SubJobIds              []string    `json:"subJobIds"`              // 子JOB ID
	SubJobDisplay          string      `json:"subJobDisplay"`          // 子JOB名称展示
	TimeStep               int64       `json:"timeStep"`               // 时间步进
	Executors              []*Executor `json:"executors"`              // 执行器
}

// 作业VO
type JobVo struct {
	Id                     string `json:"id"`                     // 主键
	Name                   string `json:"name"`                   // 任务名称
	Cron                   string `json:"cron"`                   // cron 表达式
	Status                 int    `json:"status"`                 // 状态 0暂停 1正常
	CreateTime             int64  `json:"createTime"`             // 创建时间
	Creator                string `json:"creator"`                // 创建人
	ExecutorSelectStrategy string `json:"executorSelectStrategy"` // 执行器选择策略 随机 全部 分片
	ExecutorCount          int    `json:"executorCount"`          // 执行器数量
}

type JobVoSortableList []*JobVo

func (ls JobVoSortableList) Len() int {
	return len(ls)
}

func (ls JobVoSortableList) Less(i, j int) bool {
	return ls[i].CreateTime > ls[j].CreateTime
}

func (ls JobVoSortableList) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}

var jobCache map[uint64]*Job = make(map[uint64]*Job)
var jobCacheLock sync.RWMutex

func clearJobCache() {
	jobCacheLock.Lock()
	defer jobCacheLock.Unlock()

	jobCache = make(map[uint64]*Job)
}

func putJobCache(id uint64, entity *Job) {
	jobCacheLock.Lock()
	defer jobCacheLock.Unlock()

	jobCache[id] = entity
}

func getJobCache(id uint64) *Job {
	jobCacheLock.RLock()
	defer jobCacheLock.RUnlock()

	entity, exist := jobCache[id]
	if exist {
		return entity
	}
	return nil
}

func ForEachJob() ([]*Job, error) {
	list := make([]*Job, 0)
	err := GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			entity := new(Job)
			if err := msgpack.Unmarshal(v, entity); err == nil {
				entity.IdStr = stringutil.UintToStr(entity.Id)
				list = append(list, entity)
			} else {
				logs.Error(err.Error())
			}
		}
		return nil
	})
	return list, err
}

func CascadeInsertJob(job *Job) error {
	err := GetBoltDB().Update(func(tx *bolt.Tx) error {
		jbt := tx.Bucket(jobBucket)
		jbs, err := msgpack.Marshal(job)
		if err != nil {
			return err
		}
		err = jbt.Put(byteutil.Uint64ToBytes(job.Id), jbs)
		if err != nil {
			return err
		}

		tbt := tx.Bucket(triggeredBucket)
		triggered := &Triggered{
			Id: job.Id,
		}
		tbs, err := msgpack.Marshal(triggered)
		if err != nil {
			return err
		}
		return tbt.Put(byteutil.Uint64ToBytes(triggered.Id), tbs)
	})
	if err == nil {
		clearJobCache()
	}
	return err
}

func BatchSaveJob(jobs []*Job) error {
	err := GetBoltDB().Batch(func(tx *bolt.Tx) error {
		jbt := tx.Bucket(jobBucket)
		for _, job := range jobs {
			jbs, err := msgpack.Marshal(job)
			if err != nil {
				continue
			}
			jbt.Put(byteutil.Uint64ToBytes(job.Id), jbs)
		}
		return nil
	})
	if err == nil {
		clearJobCache()
	}
	return err
}

func UpdateJob(job *Job) error {
	err := GetBoltDB().Update(func(tx *bolt.Tx) error {
		bt := tx.Bucket(jobBucket)
		bs, err := msgpack.Marshal(job)
		if err != nil {
			return err
		}
		return bt.Put(byteutil.Uint64ToBytes(job.Id), bs)
	})
	if err == nil {
		clearJobCache()
	}
	return err
}

func DeleteJob(id uint64) error {
	err := GetBoltDB().Update(func(tx *bolt.Tx) error {
		jbt := tx.Bucket(jobBucket)
		err := jbt.Delete(byteutil.Uint64ToBytes(id))
		if err != nil {
			return err
		}

		tbt := tx.Bucket(triggeredBucket)
		return tbt.Delete(byteutil.Uint64ToBytes(id))
	})
	if err == nil {
		clearJobCache()
	}
	return err
}

func GetJob(id uint64) (*Job, error) {
	if getJobCache(id) != nil {
		logs.Infof("JobId %v 命中缓存", id)
		return getJobCache(id), nil
	}

	var val []byte
	err := GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobBucket)
		val = bucket.Get(byteutil.Uint64ToBytes(id))
		if val == nil {
			return errors.Errorf("Key Not Found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var entity = new(Job)
	err = msgpack.Unmarshal(val, entity)
	if err == nil {
		putJobCache(entity.Id, entity)
	}
	return entity, err
}

func SelectSubJobSelectionList(id uint64) []*JobVo {
	list := make([]*JobVo, 0)
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var entity Job
			if err := msgpack.Unmarshal(v, &entity); err == nil {
				if entity.Id == id {
					continue
				}
				circle := false
				for _, vv := range entity.SubJobIds {
					if vv == stringutil.UintToStr(id) {
						circle = true
						break
					}
				}
				if circle {
					continue
				}
				list = append(list, &JobVo{
					Id:                     stringutil.UintToStr(entity.Id),
					Name:                   entity.Name,
					Cron:                   entity.Cron,
					Status:                 entity.Status,
					CreateTime:             entity.CreateTime,
					Creator:                entity.Creator,
					ExecutorSelectStrategy: entity.ExecutorSelectStrategy,
					ExecutorCount:          len(entity.Executors),
				})
			}
		}
		return nil
	})
	vos := JobVoSortableList(list)
	sort.Sort(vos)
	return vos
}

func SelectJobList(ps *Condition) []*JobVo {
	list := make([]*JobVo, 0)
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var entity Job
			if err := msgpack.Unmarshal(v, &entity); err == nil {
				name := ps.GetStringParam("name")
				if name != "" && !strings.Contains(entity.Name, name) {
					continue
				}
				creator := ps.GetStringParam("creator")
				if creator != "" && !strings.Contains(entity.Creator, creator) {
					continue
				}
				status := ps.GetStringParam("status")
				if status != "" && entity.Status != stringutil.ToIntSafe(status) {
					continue
				}
				list = append(list, &JobVo{
					Id:                     stringutil.UintToStr(entity.Id),
					Name:                   entity.Name,
					Cron:                   entity.Cron,
					Status:                 entity.Status,
					CreateTime:             entity.CreateTime,
					Creator:                entity.Creator,
					ExecutorSelectStrategy: entity.ExecutorSelectStrategy,
					ExecutorCount:          len(entity.Executors),
				})
			}
		}
		return nil
	})
	vos := JobVoSortableList(list)
	sort.Sort(vos)
	return vos
}

func GetJobAmount() int {
	var amount int
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobBucket)
		cursor := bucket.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			amount = amount + 1
		}
		return nil
	})
	return amount
}

func GetExecutorAmount() int {
	amount := 0
	GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobBucket)
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var entity Job
			if err := msgpack.Unmarshal(v, &entity); err == nil {
				if entity.Executors != nil {
					amount = amount + len(entity.Executors)
				}
			}
		}
		return nil
	})
	return amount
}
