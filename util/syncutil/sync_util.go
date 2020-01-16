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
package syncutil

import "sync"

type MutexSlice struct {
	slice []interface{}
	lock  sync.Mutex //互斥锁
}

func NewMutexSlice() *MutexSlice {
	return &MutexSlice{
		slice: make([]interface{}, 0),
	}
}

func (this *MutexSlice) Add(val interface{}) *MutexSlice {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.slice = append(this.slice, val)
	return this
}

func (this *MutexSlice) Del(index int) *MutexSlice {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.slice = append(this.slice[:index], this.slice[index:]...)
	return this
}

func (this *MutexSlice) Get(index int) interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()

	return this.slice[index]
}

func (this *MutexSlice) Size() int {
	this.lock.Lock()
	defer this.lock.Unlock()

	return len(this.slice)
}
