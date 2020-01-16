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
package routes

import (
	"gojob/models"
	"gojob/util/logs"
	"gojob/util/stringutil"

	"github.com/gin-gonic/gin"
)

func tracePage(c *gin.Context) {
	current := stringutil.ToIntSafe(c.Query("page_num"))
	limit := stringutil.ToIntSafe(c.Query("page_size"))
	page := models.NewPage(current, limit).
		AddParam("jobName", c.Query("job_name")).
		AddParam("startTime", c.Query("start_time")).
		AddParam("endTime", c.Query("end_time")).
		AddParam("executeStatus", c.Query("execute_status")).
		AddParam("scheduleType", c.Query("schedule_type"))
	err := models.SelectTracePage(page)
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
	} else {
		ls := page.Data.([]*models.Trace)
		for _, t := range ls {
			t.IdStr = stringutil.UintToStr(t.Id)
		}
		respondPage(c, page)
	}
}

func getTrace(c *gin.Context) {
	traceId := stringutil.ToUintSafe(c.Param("id"))
	ls, err := models.GetTrace(traceId)
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
	} else {
		respondData(c, ls)
	}
}

func cleanTrace(c *gin.Context) {
	temp := struct {
		JobId string
		Scope string
	}{}
	err := c.BindJSON(&temp)
	if nil != err {
		respond400(c, err.Error())
		return
	}
	models.CleanTrace(stringutil.ToUintSafe(temp.JobId), temp.Scope)
	respondOK(c)
}

func statisticTodayTrace(c *gin.Context) {
	ls, err := models.StatisticTodayTrace()
	if err != nil {
		respond500(c, err.Error())
		return
	}
	respondData(c, ls)
}

func statisticWeekTrace(c *gin.Context) {
	ls, err := models.StatisticWeekTrace()
	if err != nil {
		respond500(c, err.Error())
		return
	}
	respondData(c, ls)
}

func statisticMonthTrace(c *gin.Context) {
	ls, err := models.StatisticMonthTrace()
	if err != nil {
		respond500(c, err.Error())
		return
	}
	respondData(c, ls)
}

func statisticAllTrace(c *gin.Context) {
	ls, err := models.StatisticAllTrace()
	if err != nil {
		respond500(c, err.Error())
		return
	}
	respondData(c, ls)
}
