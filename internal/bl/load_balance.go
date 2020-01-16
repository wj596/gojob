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
package bl

import (
	"math/rand"
	"sync"

	"go.uber.org/atomic"
)

var lbs sync.Map

type LoadItem struct {
	Index  int
	Weight int
}

type LoadBalance interface {
	DoSelect([]LoadItem) int
}

// -------------------- 随机负载均衡
type RandomLoadBalance struct {
}

func NewRandomLoadBalance() LoadBalance {
	return &RandomLoadBalance{}
}

func (this *RandomLoadBalance) DoSelect(items []LoadItem) int {
	return items[rand.Intn(len(items))].Index
}

// -------------------- 轮询负载均衡
type RoundLoadBalance struct {
	round atomic.Int64
}

func NewRoundLoadBalance() LoadBalance {
	return &RoundLoadBalance{}
}

func (this *RoundLoadBalance) DoSelect(items []LoadItem) int {
	this.round.Add(1)
	index := this.round.Load() % int64(len(items))
	return items[index].Index
}

// -------------------- 加权随机负载均衡
type WeightRandomLoadBalance struct {
}

func NewWeightRandomLoadBalance() LoadBalance {
	return &WeightRandomLoadBalance{}
}

func (this *WeightRandomLoadBalance) DoSelect(items []LoadItem) int {
	ns := make([]int, 0)
	for _, item := range items {
		for i := 0; i < item.Weight; i++ {
			ns = append(ns, item.Index)
		}
	}
	return ns[rand.Intn(len(ns))]
}

// -------------------- 加权轮询负载均衡
type WeightRoundLoadBalance struct {
	round atomic.Int64
}

func NewWeightRoundLoadBalance() LoadBalance {
	return &WeightRoundLoadBalance{}
}

func (this *WeightRoundLoadBalance) DoSelect(items []LoadItem) int {
	ns := make([]int, 0)
	for _, item := range items {
		for i := 0; i < item.Weight; i++ {
			ns = append(ns, item.Index)
		}
	}
	this.round.Add(1)
	index := this.round.Load() % int64(len(ns))
	return ns[index]
}
