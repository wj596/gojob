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
package icron

import (
	"time"

	"github.com/robfig/cron"
	"go.uber.org/atomic"
)

// 调度任务
type TaskFunc func()

// Cron调度器
type Scheduler struct {
	started   atomic.Bool // 0停止  1运行
	startTime int64
	spec      string
	cronLab   *cron.Cron
	taskFunc  TaskFunc // 任务
	job       cron.Job // 任务
}

// 启动Cron调度器
func (this *Scheduler) Start() {
	if this.started.Load() {
		return
	}
	this.started.Store(true)
	this.startTime = time.Now().Unix()
	this.cronLab.Start()
}

// 停止Cron调度器
func (this *Scheduler) Stop() {
	if this.started.Load() {
		this.cronLab.Stop()
		this.started.Store(false)
	}
}

// 获取下次执行时间
func (this *Scheduler) GetNextTime() int64 {
	if this.started.Load() {
		entries := this.cronLab.Entries()
		if nil != entries && nil != entries[0] {
			return entries[0].Next.Unix()
		}
	}
	now := time.Now().In(this.cronLab.Location())
	actual, _ := cron.Parse(this.spec)
	return actual.Next(now).Unix()
}

// 获取Task
func (this *Scheduler) GetTaskFunc() TaskFunc {
	return this.taskFunc
}

// 获取Task
func (this *Scheduler) GetJob() cron.Job {
	return this.job
}

// 创建Cron调度器
func NewFuncScheduler(spec string, task TaskFunc) (*Scheduler, error) {
	scheduler := &Scheduler{
		taskFunc: task,
		spec:     spec,
	}
	cronLab := cron.New()
	err := cronLab.AddFunc(spec, task)
	if err != nil {
		return nil, err
	}
	scheduler.cronLab = cronLab
	return scheduler, nil
}

// 创建Cron调度器
func NewJobScheduler(spec string, job cron.Job) (*Scheduler, error) {
	scheduler := &Scheduler{
		job:  job,
		spec: spec,
	}
	cronLab := cron.New()
	err := cronLab.AddJob(spec, job)
	if err != nil {
		return nil, err
	}
	scheduler.cronLab = cronLab
	return scheduler, nil
}

// 验证Cron表达式
func ValidateCronSpec(spec string) error {
	_, err := cron.Parse(spec)
	return err
}

// 获取两次执行时间的间隔
func GetTimeStep(spec string) int64 {
	now := time.Now()
	actual, _ := cron.Parse(spec)
	next1 := actual.Next(now)
	next2 := actual.Next(next1)
	return next2.Unix() - next1.Unix()
}
