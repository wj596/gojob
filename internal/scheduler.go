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
package internal

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gojob/internal/bl"
	"gojob/internal/icron"
	"gojob/models"
	"gojob/util/dateutil"
	"gojob/util/logs"
	"gojob/util/stringutil"

	"github.com/pkg/errors"
)

const detailLineSeparator = "<line>"

var schedulerMap map[uint64]*icron.Scheduler = make(map[uint64]*icron.Scheduler)
var roundLoadBalances map[uint64]bl.LoadBalance = make(map[uint64]bl.LoadBalance)
var weightRoundLoadBalances map[uint64]bl.LoadBalance = make(map[uint64]bl.LoadBalance)
var randomLoadBalance bl.LoadBalance
var weightRandomLoadBalance bl.LoadBalance
var schedulerMapLock sync.Mutex

// 初始化任务调度器
func InitSchedulers() {
	log.Print("启动 任务调度器")
	randomLoadBalance = bl.NewRandomLoadBalance()
	weightRandomLoadBalance = bl.NewWeightRandomLoadBalance()
	jobs, err := models.ForEachJob()
	if err != nil {
		logs.Errorf("查询任务列表失败：%s", err.Error())
		return
	}
	if len(jobs) == 0 {
		logs.Info("没有需要初始化的任务")
		return
	}
	count := 0
	for _, job := range jobs {
		if !existScheduler(job.Id) {
			err := addScheduler(job)
			if err == nil && models.JobStatusOk == job.Status {
				scheduleTask(job.Id)
				count++
			}
		}
	}
	log.Printf("初始化任务数量: %d", count)
	scanMisfires()
}

// 删除任务调度器
func deleteSchedulers() {
	schedulerMapLock.Lock()
	defer schedulerMapLock.Unlock()

	for jobId, _ := range schedulerMap {
		cron, exist := schedulerMap[jobId]
		if exist {
			cron.Stop()
		}
		delete(schedulerMap, jobId)
	}
}

func existScheduler(jobId uint64) bool {
	schedulerMapLock.Lock()
	defer schedulerMapLock.Unlock()

	_, exist := schedulerMap[jobId]
	return exist
}

func getScheduler(jobId uint64) (*icron.Scheduler, bool) {
	schedulerMapLock.Lock()
	defer schedulerMapLock.Unlock()

	sch, exist := schedulerMap[jobId]
	if exist {
		return sch, true
	}
	return nil, false
}

// 添加调度器
func addScheduler(job *models.Job) error {
	schedulerMapLock.Lock()
	defer schedulerMapLock.Unlock()

	roundLoadBalances[job.Id] = bl.NewRoundLoadBalance()
	weightRoundLoadBalances[job.Id] = bl.NewWeightRoundLoadBalance()
	task := newTask(job)
	cron, err := icron.NewJobScheduler(job.Cron, task)
	if err != nil {
		logs.Errorf("Job(%s)创建调度器失败：%s", job.Name, err.Error())
		return err
	}
	schedulerMap[job.Id] = cron
	logs.Infof("Job(%s)成功创建调度器", job.Name)
	return nil
}

// 启动
func scheduleTask(jobId uint64) {
	cron, exist := schedulerMap[jobId]
	td, err := models.GetTriggered(jobId)
	if exist && err == nil {
		if 0 == td.NextTime {
			td.NextTime = cron.GetNextTime()
			models.SaveTriggered(td)
		}
		cron.Start()
	}
	logs.Infof("启动任务：%v", jobId)
}

// 挂起
func suspendTask(jobId uint64) {
	cron, exist := schedulerMap[jobId]
	if exist {
		cron.Stop()
	}
	updateTriggered(jobId, 0, 0)
	logs.Infof("挂起任务：%v", jobId)
}

// 取消任务
func cancelTask(jobId uint64) {
	schedulerMapLock.Lock()
	defer schedulerMapLock.Unlock()

	suspendTask(jobId)
	delete(schedulerMap, jobId)
	logs.Infof("取消任务：%v", jobId)
}

// 手动触发任务
func LaunchTask(jobId uint64) error {
	job, err := models.GetJob(jobId)
	if err != nil {
		logs.Errorf("任务调度失败,查找Job信息错误:%s", err.Error())
		return err
	}

	sch, exist := getScheduler(jobId)
	if !exist {
		logs.Errorf("任务调度失败,未找到Job(%v)的调度器:%s", jobId)
		return errors.Errorf("未找到调度器")
	}

	task := sch.GetJob()
	httpTask, succeed := task.(*HttpTask)
	if !succeed {
		return errors.Errorf("任务类型转换错误")
	}

	ctx := &scheduleContext{
		job:          job,
		scheduleType: models.ScheduleTypeManual,
		startTime:    time.Now().Unix(),
		details:      make([]string, 0),
	}
	logs.Infof("手动执行:%s", jobId)
	ctx.detail("手动执行")
	go httpTask.doRun(ctx)

	return nil
}

// 调度上下文
type scheduleContext struct {
	scheduleType int         // 调度类型
	startTime    int64       // 调度开始时间
	lock         sync.Mutex  // 互斥锁
	details      []string    // 详细信息
	job          *models.Job // 作业
}

func (this *scheduleContext) succeed() {
	trace := models.Trace{
		Id:            GetSnowId(),
		JobId:         this.job.Id,
		JobName:       this.job.Name,
		ScheduleType:  this.scheduleType,
		StartTime:     this.startTime,
		EndTime:       time.Now().Unix(),
		ExecuteStatus: models.ExecuteStatusSucceed,
		ExecuteResult: "执行成功",
	}

	// 需要触发子任务
	if len(this.job.SubJobIds) > 0 &&
		(models.SubJobScheduleStrategyEnd == this.job.SubJobScheduleStrategy ||
			models.SubJobScheduleStrategyOk == this.job.SubJobScheduleStrategy) {
		this.detail(fmt.Sprintf("开始触发子任务，子任务数量:%d", len(this.job.SubJobIds)))
		go this.launchSubTask()
	}

	trace.ExecuteDetail = strings.Join(this.details, detailLineSeparator)
	models.InsertTrace(&trace)
}

func (this *scheduleContext) failed(reason string) {
	trace := models.Trace{
		Id:            GetSnowId(),
		JobId:         this.job.Id,
		JobName:       this.job.Name,
		ScheduleType:  this.scheduleType,
		StartTime:     this.startTime,
		EndTime:       time.Now().Unix(),
		ExecuteStatus: models.ExecuteStatusFailed,
		ExecuteResult: reason,
	}

	// 需要触发子任务
	if len(this.job.SubJobIds) > 0 &&
		(models.SubJobScheduleStrategyEnd == this.job.SubJobScheduleStrategy ||
			models.SubJobScheduleStrategyFail == this.job.SubJobScheduleStrategy) {
		this.detail(fmt.Sprintf("开始触发子任务，子任务数量:%d", len(this.job.SubJobIds)))
		go this.launchSubTask()
	}

	trace.ExecuteDetail = strings.Join(this.details, detailLineSeparator)

	// 需要告警
	if "" != this.job.AlarmEmail {
		models.SendAlarmEmail(&models.AlarmEmail{
			Toers:   this.job.AlarmEmail,
			Subject: fmt.Sprintf("Go-Job告警,任务(%s)执行失败。", this.job.Name),
			Body:    fmt.Sprintf("告警时间：%s  <br>任务执行结果：%s  <br>详细执行信息：%s", dateutil.NowFormatted(), reason, trace.ExecuteDetail),
		})
		this.detail("发送告警邮件")
	}

	models.InsertTrace(&trace)
}

func (this *scheduleContext) detail(msg string) {
	if len(msg) > 100 {
		msg = string([]byte(msg)[:100]) + " ... ..."
	}
	this.details = append(this.details, msg)
}

func (this *scheduleContext) mutexDetail(msg string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.detail(msg)
}

func (this *scheduleContext) launchSubTask() {
	for _, v := range this.job.SubJobIds {
		subJobId := stringutil.ToUintSafe(v)
		subJob, err := models.GetJob(subJobId)
		if err != nil {
			logs.Warnf("子任务调度失败,查找子任务信息错误:%s", err.Error())
			continue
		}

		sch, exist := getScheduler(subJobId)
		if !exist {
			logs.Warnf("子任务调度失败,未找到子任务(%s)的调度器", subJob.Name)
			continue
		}

		task := sch.GetJob()
		httpTask, succeed := task.(*HttpTask)
		if !succeed {
			logs.Warn("子任务调度失败,任务类型转换错误")
			continue
		}

		ctx := &scheduleContext{
			job:          subJob,
			scheduleType: models.ScheduleTypeDepend,
			startTime:    time.Now().Unix(),
			details:      make([]string, 0),
		}

		logs.Infof("调度子任务:%s", subJob.Name)
		ctx.detail(fmt.Sprintf("子任务触发，父任务名称:%s", this.job.Name))
		httpTask.doRun(ctx)
	}
}

var processingMisfires sync.Map

// 扫描错发任务
func scanMisfires() {
	misfires := models.SelectMisfireList()
	if len(misfires) == 0 {
		logs.Infof("无Misfire任务")
		return
	}

	logs.Infof("Misfire任务数量：%d", len(misfires))
	for _, misfire := range misfires {
		if _, exist := processingMisfires.Load(misfire.Id); !exist {
			processingMisfires.Store(misfire.Id, true)
			go handleMisfire(misfire)
		}
	}
}

func handleMisfire(triggered *models.Triggered) {
	job, err := models.GetJob(triggered.Id)
	if err != nil {
		logs.Errorf("任务调度失败,查找Job信息错误:%s", err.Error())
		return
	}

	sch, exist := schedulerMap[triggered.Id]
	if !exist {
		logs.Errorf("任务调度失败,未找到Job(%v)的调度器:%s", triggered.Id)
		return
	}

	task := sch.GetJob()
	httpTask, succeed := task.(*HttpTask)
	if !succeed {
		logs.Errorf("任务类型转换错误")
		return
	}

	past := triggered.NextTime
	startTime := time.Now().Unix()
	nextTime := past + job.TimeStep
	updateTriggered(triggered.Id, startTime, nextTime)

	ctx := &scheduleContext{
		job:          job,
		scheduleType: models.ScheduleTypeCompensation,
		startTime:    startTime,
		details:      make([]string, 0),
	}

	logs.Infof(fmt.Sprintf("补偿执行,被补偿的执行时间点为：%s", dateutil.Layout(time.Unix(past, 0), dateutil.DayTimeSecondFormatter)))
	ctx.detail(fmt.Sprintf("补偿执行,被补偿的执行时间点为：%s", dateutil.Layout(time.Unix(past, 0), dateutil.DayTimeSecondFormatter)))
	httpTask.doRun(ctx)
	processingMisfires.Delete(triggered.Id)
}
