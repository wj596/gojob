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
	"net/http"

	"gojob/models"

	"github.com/gin-gonic/gin"
)

func respond400(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"succeed": false,
		"msg":     err,
	})
}

func respond500(c *gin.Context, err string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"succeed": false,
		"msg":     err,
	})
}

func respondOK(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"succeed": true,
		"msg":     "操作成功",
	})
}

func respondMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"succeed": true,
		"msg":     msg,
	})
}

func respondData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"succeed": true,
		"data":    data,
	})
}

func respondPage(c *gin.Context, page *models.Page) {
	c.JSON(http.StatusOK, gin.H{
		"succeed": true,
		"total":   page.Total,
		"data":    page.Data,
	})
}

func initActions() {

	router.GET("ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	bolt := router.Group("/bolt")
	bolt.GET("/job", forEachJob)
	bolt.GET("/triggered", forEachTriggered)
	bolt.GET("/user", forEachUser)
	bolt.GET("/node", forEachNode)
	bolt.GET("/alarm_config", forEachAlarmConfig)
	bolt.GET("/snapshot_version", forEachSnapshotVersion)
	bolt.GET("/raft_flag", forEachRaftFlag)

	cluster := router.Group("/cluster")
	//cluster.Use(signMiddleware())
	cluster.GET("/join/:peer_node_name/:peer_http_addr/:peer_tcp_addr", joinCluster)
	cluster.GET("/leader_id", getClusterLeaderId)

	ui := router.Group("/ui")
	ui.Use(authMiddleware())
	ui.GET("index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	ui.POST("jobs", insertJob)
	ui.POST("jobs/cron_validate", validateCron)
	ui.DELETE("jobs/:id", deleteJob)
	ui.PUT("jobs", updateJob)
	ui.PUT("jobs/update_status/:id/:status", updateJobStatus)
	ui.GET("jobs", searchJob)
	ui.GET("jobs/:id", getJob)
	ui.GET("jobs/:id/launch", launchJob)

	ui.GET("users", searchUser)
	ui.GET("users/name/:name", getUser)
	ui.PUT("users", updateUser)
	ui.POST("users", insertUser)
	ui.DELETE("users/:id", deleteUser)
	ui.POST("users/login", login)
	ui.GET("users/logout", logout)
	ui.GET("users/authorised", authorised)

	ui.GET("alarm_configs", getAlarmConfig)
	ui.PUT("alarm_configs", updateAlarmConfig)
	ui.POST("alarm_configs/test", testAlarmConfig)

	ui.GET("/runtimes", getRuntime)
	ui.GET("/runtimes/runmode", getRunmode)

	ui.GET("/cluster/nodes", getClusterNodes)
	ui.GET("/cluster/leader_id", getClusterLeaderId)
	ui.GET("/cluster/remove/:peer_node_name", removePeer)

	ui.GET("traces", tracePage)
	ui.GET("traces/:id", getTrace)
	ui.POST("traces/clean", cleanTrace)
	ui.GET("statistic/today", statisticTodayTrace)
	ui.GET("statistic/week", statisticWeekTrace)
	ui.GET("statistic/month", statisticMonthTrace)
	ui.GET("statistic/all", statisticAllTrace)
}
