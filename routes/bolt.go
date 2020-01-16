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

	"github.com/gin-gonic/gin"
)

func forEachJob(c *gin.Context) {
	datas, err := models.ForEachJob()
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
	} else {
		respondData(c, datas)
	}
}

func forEachTriggered(c *gin.Context) {
	datas, err := models.ForEachTriggered()
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
	} else {
		respondData(c, datas)
	}
}

func forEachUser(c *gin.Context) {
	datas, err := models.ForEachUser()
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
	} else {
		respondData(c, datas)
	}
}

func forEachNode(c *gin.Context) {
	datas, err := models.ForEachNode()
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
	} else {
		respondData(c, datas)
	}
}

func forEachAlarmConfig(c *gin.Context) {
	datas, err := models.GetAlarmConfig()
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
	} else {
		respondData(c, datas)
	}
}

func forEachSnapshotVersion(c *gin.Context) {
	v := models.GetSnapshotVersion()
	respondData(c, v)
}

func forEachRaftFlag(c *gin.Context) {
	v := models.IsRaftFirstStart()
	respondData(c, v)
}
