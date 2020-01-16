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
	"strconv"
	"time"

	"gojob/util/dateutil"
	"gojob/util/logs"
	"gojob/util/sqlutil"
	"gojob/util/stringutil"

	"github.com/go-xorm/xorm"
)

const (
	// 调度类型 -- 手动
	ScheduleTypeManual = 0
	// 调度类型 -- 自动
	ScheduleTypeAuto = 1
	// 调度类型 -- 补偿
	ScheduleTypeCompensation = 2
	// 调度类型 -- 依赖
	ScheduleTypeDepend = 3
	// 执行状态 -- 失败
	ExecuteStatusFailed = 0
	// 执行状态 -- 成功
	ExecuteStatusSucceed = 1
	// 日志数据清理范围 -- 全部
	cleanScopeAll = "1"
	// 日志数据清理范围 -- 一周前
	cleanScopeWeekAgo = "2"
	// 日志数据清理范围 -- 一月前
	cleanScopeMonthAgo = "3"
	// 日志数据清理范围 -- 二月前
	cleanScopeTwoMonthAgo = "4"
	// 日志数据清理范围 -- 三月前
	cleanScopeThreeMonthAgo = "5"
	// 日志数据清理范围 -- 六月前
	cleanScopeSixMonthAgo = "6"
	// 日志数据清理范围 -- 六月前
	cleanScopeYearAgo     = "7"
	selectMaxStartTimeSql = "SELECT MAX(START_TIME) FROM T_TRACE"
	statisticTraceSql     = "SELECT T.JOB_ID," +
		"COUNT(1) AS TOTAL," +
		"SUM(CASE WHEN T.EXECUTE_STATUS = 1 THEN 1  ELSE 0  END) SUCCEED," +
		"SUM(CASE WHEN T.EXECUTE_STATUS = 0 THEN 1  ELSE 0  END) FAILED," +
		"ROUND(ROUND(SUM(CASE WHEN T.EXECUTE_STATUS = 0 THEN 1  ELSE 0  END)/COUNT(1),2)*100) AS RATE " +
		"FROM T_TRACE T " +
		"GROUP BY T.JOB_ID " +
		"ORDER BY RATE DESC " +
		"LIMIT 20"
	statisticTraceByTimeSql = "SELECT T.JOB_ID," +
		"COUNT(1) AS TOTAL," +
		"SUM(CASE WHEN T.EXECUTE_STATUS = 1 THEN 1  ELSE 0  END) SUCCEED," +
		"SUM(CASE WHEN T.EXECUTE_STATUS = 0 THEN 1  ELSE 0  END) FAILED," +
		"ROUND(ROUND(SUM(CASE WHEN T.EXECUTE_STATUS = 0 THEN 1  ELSE 0  END)/COUNT(1),2)*100) AS RATE " +
		"FROM T_TRACE T " +
		"WHERE T.START_TIME BETWEEN ? AND ? " +
		"GROUP BY T.JOB_ID " +
		"ORDER BY RATE DESC " +
		"LIMIT 20"
	createTraceTableSql = "CREATE TABLE `t_trace`  (" +
		"`ID` bigint(18) NOT NULL COMMENT '主键'," +
		"`JOB_ID` bigint(18) NULL DEFAULT NULL COMMENT 'JOB主键'," +
		"`JOB_NAME` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT 'JOB名称'," +
		"`SCHEDULE_TYPE` int(2) NULL DEFAULT NULL COMMENT '调度类型 0手动/1自动/2补偿'," +
		"`START_TIME` bigint(10) NULL DEFAULT NULL COMMENT '开始时间'," +
		"`END_TIME` bigint(10) NULL DEFAULT NULL COMMENT '结束时间'," +
		"`EXECUTE_STATUS` int(2) NULL DEFAULT NULL COMMENT '执行状态 0失败/1成功'," +
		"`EXECUTE_RESULT` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '调度信息'," +
		"`EXECUTE_DETAIL` varchar(1000) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '执行明细'," +
		"PRIMARY KEY (`ID`) USING BTREE," +
		"INDEX `index_job_id`(`JOB_ID`) USING BTREE," +
		"INDEX `index_start_time`(`START_TIME`) USING BTREE" +
		") "
)

// 调度跟踪信息
type Trace struct {
	Id            uint64 `xorm:"pk" json:"-"`   // 主键
	IdStr         string `xorm:"-" json:"id"`   // 主键
	JobId         uint64 `json:"jobId"`         //JOB主键
	JobName       string `json:"jobName"`       // JOB名称
	ScheduleType  int    `json:"scheduleType"`  // 调度类型
	StartTime     int64  `json:"startTime"`     // 开始时间
	EndTime       int64  `json:"endTime"`       // 结束时间
	ExecuteStatus int    `json:"executeStatus"` // 执行状态
	ExecuteResult string `json:"executeResult"` // 执行结果
	ExecuteDetail string `json:"executeDetail"` // 调度明细信息
}

// 调度跟踪统计
type TraceStatistic struct {
	JobId   uint64 `json:"-"`       // 作业ID
	Name    string `json:"name"`    // 作业名称
	Total   uint64 `json:"total"`   // 调度次数
	Succeed uint64 `json:"succeed"` // 调度成功次数
	Failed  uint64 `json:"failed"`  // 调度失败次数
	Rate    uint64 `json:"rate"`    // 调度失败比率
}

var traceSyncQueue = make(chan *Trace, 65535)

func createTraceTableNecessary(engine *xorm.Engine) error {
	exist, err := engine.IsTableExist("t_trace")
	if err != nil {
		return err
	}
	if !exist {
		_, err := engine.Exec(createTraceTableSql)
		if err != nil {
			return err
		}
	}
	return nil
}

func selectMaxStartTime(engine *xorm.Engine) int64 {
	var max int64
	engine.DB().QueryRow(selectMaxStartTimeSql).Scan(&max)
	return max
}

func InsertTrace(trace *Trace) error {
	_, err := GetOrm().InsertOne(trace)

	if isRedundancy() {
		if err != nil {
			if tryCutDB() {
				GetOrm().InsertOne(trace)
			}
		}
		traceSyncQueue <- trace
	}

	return err
}

func SelectTracePage(page *Page) error {
	jobName := page.GetStringParam("jobName")
	startTime := page.GetStringParam("startTime")
	endTime := page.GetStringParam("endTime")
	executeStatus := page.GetStringParam("executeStatus")
	scheduleType := page.GetStringParam("scheduleType")
	builder := sqlutil.NewSqlBuilder().
		SELECT("COUNT(1)").
		FROM("T_TRACE T").
		WHEREF_NECESSARY("" != jobName, "T.JOB_NAME like '%s'", sqlutil.Like(jobName)).
		WHEREF_NECESSARY("" != executeStatus, "T.EXECUTE_STATUS = %d", stringutil.ToIntSafe(executeStatus)).
		WHEREF_NECESSARY("" != scheduleType, "T.SCHEDULE_TYPE = %d", stringutil.ToIntSafe(scheduleType))
	if "" != startTime && "" != endTime {
		builder.WHEREF("T.START_TIME BETWEEN %d AND %d", stringutil.ToIntSafe(startTime), stringutil.ToIntSafe(endTime))
	}
	var total int64
	err := GetOrm().DB().QueryRow(builder.Sql()).Scan(&total)
	if nil != err {
		return err
	}
	builder.REST_SELECT().
		SELECT("T.ID,T.JOB_NAME,T.SCHEDULE_TYPE,T.START_TIME,T.END_TIME,T.EXECUTE_STATUS,T.EXECUTE_RESULT").
		ORDER_BY("T.START_TIME DESC").
		LIMIT(page.Limit, page.GetStartRow())
	list := make([]*Trace, 0)
	err = GetOrm().SQL(builder.Sql()).Find(&list)
	if nil != err {
		return err
	}
	page.Total = total
	page.Data = list
	return nil
}

func GetTrace(id uint64) (*Trace, error) {
	var entity Trace
	succeed, err := GetOrm().Where("ID=?", id).Get(&entity)
	if succeed {
		return &entity, nil
	}
	return nil, err
}

func CleanTrace(jobId uint64, scope string) {
	var timestamp int64
	switch scope {
	case cleanScopeWeekAgo:
		timestamp = dateutil.PastDayDate(7).Unix()
	case cleanScopeMonthAgo:
		timestamp = dateutil.PastDayDate(30).Unix()
	case cleanScopeTwoMonthAgo:
		timestamp = dateutil.PastDayDate(60).Unix()
	case cleanScopeThreeMonthAgo:
		timestamp = dateutil.PastDayDate(90).Unix()
	case cleanScopeSixMonthAgo:
		timestamp = dateutil.PastDayDate(180).Unix()
	case cleanScopeYearAgo:
		timestamp = dateutil.PastDayDate(365).Unix()
	}
	sql := "DELETE FROM T_TRACE WHERE 1=1"
	if 0 != jobId {
		sql = sql + "AND JOB_ID = " + strconv.FormatUint(jobId, 10)
	}
	if 0 != timestamp {
		sql = sql + "AND START_TIME < " + strconv.FormatInt(timestamp, 10)
	}

	redundancyMap.Range(func(key, value interface{}) bool {
		if !isDBInvalid(key.(string)) {
			if _, err := value.(*redundancy).engine.Exec(sql); err != nil {
				logs.Errorf(err.Error())
			}
		}
		return true
	})
}

func StatisticTodayTrace() ([]*TraceStatistic, error) {
	today := dateutil.NowLayout(dateutil.DayFormatter)
	startTime := dateutil.FromDefaultLayout(today + " 00:00:00").Unix()
	endTime := dateutil.FromDefaultLayout(today + " 23:59:59").Unix()
	return statisticTrace(startTime, endTime)
}

func StatisticWeekTrace() ([]*TraceStatistic, error) {
	startTime := dateutil.WeekStartDayDate().Unix()
	endTime := time.Now().Unix()
	return statisticTrace(startTime, endTime)
}

func StatisticMonthTrace() ([]*TraceStatistic, error) {
	startTime := dateutil.MonthStartDayDate().Unix()
	endTime := time.Now().Unix()
	return statisticTrace(startTime, endTime)
}

func StatisticAllTrace() ([]*TraceStatistic, error) {
	startTime := int64(0)
	endTime := int64(0)
	return statisticTrace(startTime, endTime)
}

func statisticTrace(startTime int64, endTime int64) ([]*TraceStatistic, error) {
	statistics := make([]TraceStatistic, 0)
	if startTime == 0 && endTime == 0 {
		err := GetOrm().SQL(statisticTraceSql).Find(&statistics)
		if nil != err && isRedundancy() {
			if tryCutDB() {
				err = GetOrm().SQL(statisticTraceSql).Find(&statistics)
			}
		}
		if nil != err {
			return nil, err
		}
	} else {
		err := GetOrm().SQL(statisticTraceByTimeSql, startTime, endTime).Find(&statistics)
		if nil != err && isRedundancy() {
			if tryCutDB() {
				err = GetOrm().SQL(statisticTraceByTimeSql, startTime, endTime).Find(&statistics)
			}
		}
		if nil != err {
			return nil, err
		}
	}

	list := make([]*TraceStatistic, 0)
	for i, entity := range statistics {
		job, _ := GetJob(entity.JobId)
		if job != nil {
			item := &statistics[i]
			item.Name = job.Name
			list = append(list, item)
		}
	}
	return list, nil
}

func startTraceSyncQueueListener() {
	if !isRedundancy() {
		return
	}
	logs.Info("启动 多数据库同步队列监听器")
	go func() {
		for {
			trace := <-traceSyncQueue
			redundancyMap.Range(func(key, value interface{}) bool {
				r := value.(*redundancy)
				if isDBInvalid(key.(string)) {
					if ok, _ := pingDB(r.engine); ok {
						invalidMap.Delete(r.name)
						logs.Infof("数据库：%s,恢复", r.mixName)
					} else {
						logs.Warnf("数据库：%s,未恢复", r.mixName)
					}
				} else {
					_, err := r.engine.InsertOne(trace)
					if err != nil {
						if ok, _ := pingDB(r.engine); !ok {
							logs.Warnf("数据库：%s,无法链接", r.mixName)
						}
					}
				}
				return true
			})
		}
	}()
}
