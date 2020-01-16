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
	"gojob/internal"
	"gojob/models"
	"gojob/util/logs"

	"github.com/gin-gonic/gin"
)

func getAlarmConfig(c *gin.Context) {
	entity, err := models.GetAlarmConfig()
	if nil != err {
		logs.Error(err.Error())
		respond500(c, err.Error())
	} else {
		respondData(c, entity)
	}
}

func updateAlarmConfig(c *gin.Context) {
	alarmConfig := new(models.AlarmConfig)
	err := c.BindJSON(alarmConfig)
	if nil != err {
		respond400(c, err.Error())
		return
	}

	err = internal.UpdateAlarmConfig(alarmConfig)
	if nil != err {
		respond500(c, err.Error())
		return
	}
	respondOK(c)
}

func testAlarmConfig(c *gin.Context) {
	temp := struct {
		Target string
	}{}
	err := c.BindJSON(&temp)
	if nil != err {
		respond400(c, err.Error())
		return
	}
	err = models.TestMailDialer(temp.Target)
	if nil != err {
		logs.Error(err.Error())
		respond400(c, err.Error())
	} else {
		respondOK(c)
	}
}
