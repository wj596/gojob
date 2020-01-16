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
	"net/url"
	"strconv"

	"gojob/internal"
	"gojob/internal/icron"
	"gojob/models"
	"gojob/util/logs"
	"gojob/util/stringutil"

	"github.com/gin-gonic/gin"
)

const (
	searchTypeJobList         = "1"
	searchTypeSubJobSelection = "2"
)

func insertJob(c *gin.Context) {
	job := new(models.Job)
	err := c.BindJSON(job)
	if nil != err {
		respond400(c, err.Error())
		return
	}

	err = internal.InsertJob(job)
	if nil != err {
		respond500(c, err.Error())
		return
	}

	respondOK(c)
}

func deleteJob(c *gin.Context) {
	id := stringutil.ToUintSafe(c.Param("id"))
	err := internal.DeleteJob(id)
	if nil != err {
		respond500(c, err.Error())
		return
	}
	respondOK(c)
}

func updateJob(c *gin.Context) {
	job := new(models.Job)
	err := c.BindJSON(job)
	if nil != err {
		logs.Error(err.Error())
		respond400(c, err.Error())
		return
	}

	job.Id = stringutil.ToUintSafe(job.IdStr)
	job.SubJobDisplay = ""
	err = internal.UpdateJob(job)
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
		return
	}

	respondOK(c)
}

func validateCron(c *gin.Context) {
	spec, _ := url.QueryUnescape(c.Query("spec"))
	if err := icron.ValidateCronSpec(spec); err != nil {
		respondData(c, false)
		return
	}
	respondData(c, true)
}

func updateJobStatus(c *gin.Context) {
	id := stringutil.ToUintSafe(c.Param("id"))
	status, err := strconv.Atoi(c.Param("status"))
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
		return
	}
	err = internal.UpdateJobStatus(id, status)
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
		return
	}

	respondOK(c)
}

func getJob(c *gin.Context) {
	id := stringutil.ToUintSafe(c.Param("id"))
	job, err := models.GetJob(id)
	if nil != err {
		respond500(c, err.Error())
	} else {
		job.IdStr = stringutil.UintToStr(job.Id)
		subJobDisplay := ""
		if job.SubJobIds != nil && len(job.SubJobIds) > 0 {
			for _, v := range job.SubJobIds {
				if temp, err := models.GetJob(stringutil.ToUintSafe(v)); err == nil {
					if "" == subJobDisplay {
						subJobDisplay = subJobDisplay + temp.Name
					} else {
						subJobDisplay = subJobDisplay + " | " + temp.Name
					}
				}
			}
		}
		job.SubJobDisplay = subJobDisplay
		respondData(c, job)
	}
}

func searchJob(c *gin.Context) {
	if "" != c.Query("page_num") && "" != c.Query("page_size") {
		pageJob(c)
	}
	searchType := c.Query("search_type")
	if searchTypeJobList == searchType {
		ps := models.NewCondition().AddParam("name", c.Query("name"))
		respondData(c, models.SelectJobList(ps))
	}
	if searchTypeSubJobSelection == searchType {
		listSubJobSelection(c)
	}
}

func listSubJobSelection(c *gin.Context) {
	id := c.Query("id")
	list := models.SelectJobList(models.NewCondition())
	if "" == id {
		respondData(c, list)
	} else {
		respondData(c, models.SelectSubJobSelectionList(stringutil.ToUintSafe(id)))
	}
}

func pageJob(c *gin.Context) {
	pageNum := c.Query("page_num")
	pageSize := c.Query("page_size")
	ps := models.NewCondition().
		AddParam("name", c.Query("name")).
		AddParam("creator", c.Query("creator")).
		AddParam("status", c.Query("status"))
	startIndex := (stringutil.ToIntSafe(pageNum) - 1) * stringutil.ToIntSafe(pageSize)

	list := models.SelectJobList(ps)
	slice := make([]*models.JobVo, 0)
	for i := 0; i < stringutil.ToIntSafe(pageSize); i++ {
		index := startIndex + i
		if index < len(list) {
			vo := list[index]
			slice = append(slice, vo)
		}
	}
	respondPage(c, &models.Page{
		Total: int64(len(list)),
		Data:  slice,
	})
}

func launchJob(c *gin.Context) {
	id := stringutil.ToUintSafe(c.Param("id"))
	err := internal.LaunchTask(id)
	if nil != err {
		respond500(c, err.Error())
	} else {
		respondOK(c)
	}
}
