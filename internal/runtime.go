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
	"github.com/pkg/errors"
	"log"
	"strconv"
	"sync"
	"time"

	"gojob/conf"
	"gojob/models"
	"gojob/util/dateutil"
	"gojob/util/logs"
	"gojob/util/secureutil"

	"github.com/sony/sonyflake"
	"go.uber.org/atomic"
)

const (
	runModeCluster      = "cluster"
	runModeStandalone   = "standalone"
	monitorTaskInterval = 600
)

// 运行模式
var runMode atomic.String

// ID生成器
var sf *sonyflake.Sonyflake
var startTime time.Time
var alarmedMap sync.Map

func SetRunMode(mode string) {
	if runMode.Load() == "" {
		runMode.Store(mode)
		startTime = time.Now()
	}
}

func IsClusterMode() bool {
	return runModeCluster == runMode.Load()
}

func IsStandaloneMode() bool {
	return runModeStandalone == runMode.Load()
}

func IsStandaloneOrLeader() bool {
	if IsStandaloneMode() {
		return true
	}
	if IsLeader() {
		return true
	}
	return false
}

func InitSnowflake() {
	if sf != nil {
		return
	}

	var nodeName string
	if IsStandaloneMode() {
		nodeName = strconv.Itoa(conf.GetConfig().HttpServerPort)
		node, err := models.GetNode(nodeName)
		if err != nil {
			node = &models.Node{
				Name: nodeName,
			}
			InsertNode(node)
		}
	}
	if IsClusterMode() {
		nodeName = conf.GetClusterConfig().CurrentNodeName
		node, err := models.GetNode(nodeName)
		if err != nil {
			node = &models.Node{
				Name:     nodeName,
				HttpAddr: conf.GetClusterConfig().CurrentHttpAddr,
				TcpAddr:  conf.GetClusterConfig().CurrentTcpAddr,
			}
			InsertNode(node)
		} else {
			if node.TcpAddr != conf.GetClusterConfig().CurrentTcpAddr ||
				node.HttpAddr != conf.GetClusterConfig().CurrentHttpAddr {
				node.TcpAddr = conf.GetClusterConfig().CurrentTcpAddr
				node.HttpAddr = conf.GetClusterConfig().CurrentHttpAddr
				UpdateNode(node)
			}
		}
	}

	logs.Infof("初始化Snowflake，nodeName：%s.", nodeName)

	current, _ := models.GetNode(nodeName)
	machineNum := current.MachineNum
	var st sonyflake.Settings
	st.MachineID = func() (u uint16, e error) {
		return machineNum, nil
	}
	sf = sonyflake.NewSonyflake(st)
}

func GetSnowId() uint64 {
	id, err := sf.NextID()
	if err != nil {
		logs.Errorf("GetSnowId出错，无法生成ID：%s. \n", err.Error())
		log.Panicf("GetSnowId出错，无法生成ID：%s. \n", err.Error())
	}
	return id
}

func Signature(plaintext string) string {
	return secureutil.HmacMD5(plaintext, conf.GetConfig().SignSecretKey)
}

// 运行时监控任务
func StartMonitorTask() {
	ticker := time.NewTicker(monitorTaskInterval * time.Second)
	go func(ticker *time.Ticker) {
		for {
			<-ticker.C
			conf, err := models.GetAlarmConfig()
			if nil == err && "" != conf.SysAlarmEmail {
				if IsStandaloneOrLeader() {
					models.DBAlarmNecessary(conf.SysAlarmEmail)
				}
				if IsClusterMode() {
					if IsLeader() {
						clusterAlarmNecessary(conf.SysAlarmEmail)
					}
					if "" == GetLeaderId() {
						models.SendAlarmEmail(&models.AlarmEmail{
							Toers:   conf.SysAlarmEmail,
							Subject: fmt.Sprintf("Go-Job告警,集群不可用"),
							Body:    fmt.Sprintf("告警时间：%s  <br>集群故障：当前集群无法获取主节点，请检查集群各节点是否已正确启动 ", dateutil.NowFormatted()),
						})
					}
				}
			}
		}
	}(ticker)
}

// 集群告警
func clusterAlarmNecessary(sysAlarmEmail string) {
	followers := getFollowers()
	if len(followers) > 0 {
		for follower, ok := range followers {
			if ok {
				alarmedMap.Delete(follower)
			} else {
				if _, exist := alarmedMap.Load(follower); !exist {
					alarmedMap.Store(follower, true)
					logs.Warnf("集群节点：%s，告警")
					models.SendAlarmEmail(&models.AlarmEmail{
						Toers:   sysAlarmEmail,
						Subject: fmt.Sprintf("Go-Job告警,集群节点故障"),
						Body:    fmt.Sprintf("告警时间：%s  <br>集群节点：%s ，失去心跳超过30秒 ", dateutil.NowFormatted(), follower),
					})
				} else {
					logs.Warnf("集群节点：%s，未恢复，已告警", follower)
				}
			}
		} // end for
	}
}

type Runtime struct {
	RunMode            string `json:"runMode"`            // 运行模式
	StartTime          string `json:"startTime"`          // 启动时间
	ClusterNodeCount   int    `json:"clusterNodeCount"`   // 集群节点数量
	JobCount           int    `json:"jobCount"`           // Job数量
	ExecuteNodeCount   int    `json:"executeNodeCount"`   // 执行节点数量
	TriggerTimes       int64  `json:"triggerTimes"`       // 调度次数
	UsableDBAmount     int    `json:"usableDBAmount"`     // 可用数据库数量
	DisabledDBAmount   int    `json:"disabledDBAmount"`   // 不可用数据库数量
	UsableNodeAmount   int    `json:"usableNodeAmount"`   // 可用节点数量
	DisabledNodeAmount int    `json:"disabledNodeAmount"` // 不可用节点数量
}

type RuntimeClusterNode struct {
	Name     string
	HttpAddr string
	TcpAddr  string
}

func GetRuntime() *Runtime {
	r := new(Runtime)
	r.RunMode = runMode.Load()
	if IsClusterMode() {
		r.ClusterNodeCount = len(getRaftServers())
		r.UsableNodeAmount, r.DisabledNodeAmount = getNodeAmount()
	}
	r.StartTime = dateutil.Layout(startTime, dateutil.DayTimeMinuteFormatter)
	r.JobCount = models.GetJobAmount()
	r.ExecuteNodeCount = models.GetExecutorAmount()
	r.TriggerTimes = models.GetTriggeredAmount()
	r.UsableDBAmount, r.DisabledDBAmount = models.GetDBAmount()
	return r
}

func GetRuntimeClusterNode(nodeName string) (*RuntimeClusterNode, error) {
	var rcn *RuntimeClusterNode
	node, err := models.GetNode(nodeName)
	if err == nil {
		rcn = &RuntimeClusterNode{
			Name:     node.Name,
			HttpAddr: node.HttpAddr,
			TcpAddr:  node.TcpAddr,
		}
		return rcn, nil
	}

	for _, node := range conf.GetClusterConfig().Nodes {
		if nodeName == node.Name {
			rcn = &RuntimeClusterNode{
				Name:     node.Name,
				HttpAddr: node.Addr,
			}
			break
		}
	}

	if nil == rcn {
		return nil, errors.Errorf("无法获取名称为：%s 的节点信息", nodeName)
	}

	return rcn, nil
}
