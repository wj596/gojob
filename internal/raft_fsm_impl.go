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
	"io"
	"log"

	"gojob/models"
	"gojob/util/dateutil"
	"gojob/util/logs"

	"github.com/hashicorp/raft"
	"github.com/vmihailenco/msgpack"
)

const (
	commandTypeInsertJob              uint8 = 10
	commandTypeUpdateJob              uint8 = 11
	commandTypeDeleteJob              uint8 = 12
	commandTypeSaveTriggered          uint8 = 21
	commandTypeSaveNode               uint8 = 31
	commandTypeSaveUser               uint8 = 41
	commandTypeDeleteUser             uint8 = 42
	commandTypeSaveAlarmConfig        uint8 = 51
	commandTypeNegationRaftFirstStart uint8 = 61
)

type RaftSnapshot struct {
	Version     uint64
	Job         []*models.Job
	Triggered   []*models.Triggered
	Node        []*models.Node
	User        []*models.User
	AlarmConfig *models.AlarmConfig
}

type RaftCommand struct {
	Type        uint8
	EntityId    uint64
	Job         *models.Job
	Triggered   *models.Triggered
	Node        *models.Node
	User        *models.User
	AlarmConfig *models.AlarmConfig
	Snapshot    *RaftSnapshot
}

// 有限状态机FSM(finite state machine)接口的实现
type FsmImpl struct {
}

// 有限状态机快照FSMSnapshot接口的实现
type FSMSnapshotImp struct {
}

// 应用entry
func (this *FsmImpl) Apply(entry *raft.Log) interface{} {
	if IsLeader() {
		return nil
	}

	var command RaftCommand
	err := msgpack.Unmarshal(entry.Data, &command)
	if err != nil {
		logs.Infof("Raft同步数据异常:%s", err.Error())
		return err
	}

	switch command.Type {
	case commandTypeInsertJob:
		job := command.Job
		logs.Infof("Raft Command: 新建JOB(%v)", job.Id)
		models.CascadeInsertJob(job)
	case commandTypeUpdateJob:
		job := command.Job
		logs.Infof("Raft Command: 更新JOB(%v)", job.Id)
		models.UpdateJob(job)
	case commandTypeDeleteJob:
		logs.Infof("Raft Command: 删除JOB(%v)", command.EntityId)
		models.DeleteJob(command.EntityId)
	case commandTypeSaveTriggered:
		triggered := command.Triggered
		logs.Infof("Raft Command: 更新Triggered(%v)", triggered.Id)
		models.SaveTriggered(triggered)
	case commandTypeSaveNode:
		node := command.Node
		logs.Infof("Raft Command: 更新Node(%v)", node.Name)
		models.UpdateNode(node)
	case commandTypeSaveUser:
		user := command.User
		logs.Infof("Raft Command: 更新User(%v)", user.Id)
		models.SaveUser(user)
	case commandTypeDeleteUser:
		logs.Infof("Raft Command: 删除User(%v)", command.EntityId)
		models.DeleteUser(command.EntityId)
	case commandTypeSaveAlarmConfig:
		alarmConfig := command.AlarmConfig
		logs.Infof("Raft Command: 更新AlarmConfig(%v)", alarmConfig.SysAlarmEmail)
		models.SaveAlarmConfig(alarmConfig)
	case commandTypeNegationRaftFirstStart:
		logs.Info("Raft Command: NegationRaftFirstStart")
		models.NegationRaftFirstStart()
	}
	return nil
}

// 返回快照实例
func (this *FsmImpl) Snapshot() (raft.FSMSnapshot, error) {
	return &FSMSnapshotImp{}, nil
}

// 从快照重建
func (this *FsmImpl) Restore(readable io.ReadCloser) error {
	var snapshot RaftSnapshot
	err := msgpack.NewDecoder(readable).Decode(&snapshot)
	if err != nil {
		return err
	}

	if shouldRestoreSnapshot(snapshot.Version) {
		logs.Infof("Raft Command: 恢复快照Version(%v)", snapshot.Version)
		models.RestBucket()
		models.BatchSaveJob(snapshot.Job)
		models.BatchSaveTriggered(snapshot.Triggered)
		models.BatchSaveNode(snapshot.Node)
		models.BatchSaveUser(snapshot.User)
		models.SaveAlarmConfig(snapshot.AlarmConfig)
		models.UpdateSnapshotVersion(snapshot.Version)
	} else {
		logs.Infof("不需要恢复版本为%v的快照", snapshot.Version)
	}
	return err
}

// 创建系统快照
func (this *FSMSnapshotImp) Persist(sink raft.SnapshotSink) error {
	snapshot, err := createRaftSnapshot()
	if err != nil {
		return err
	}
	snapshotBytes, err := msgpack.Marshal(snapshot)
	if err != nil {
		return err
	}
	_, err = sink.Write(snapshotBytes)
	if err != nil {
		sink.Cancel()
		return err
	}

	models.UpdateSnapshotVersion(snapshot.Version)
	logs.Infof("生成快照成功,Version: %v ", snapshot.Version)
	return err
}

//释放资源
func (this *FSMSnapshotImp) Release() {

}

func shouldRestoreSnapshot(targetVersion uint64) bool {
	currentVersion := models.GetSnapshotVersion()
	logs.Infof("当前Snapshot Version: %v \n", currentVersion)
	logs.Infof("目标Snapshot Version: %v \n", targetVersion)
	return targetVersion > currentVersion
}

func createRaftSnapshot() (*RaftSnapshot, error) {
	jobs, err := models.ForEachJob()
	if err != nil {
		return nil, err
	}

	triggeredList, err := models.ForEachTriggered()
	if err != nil {
		return nil, err
	}

	nodes, err := models.ForEachNode()
	if err != nil {
		return nil, err
	}

	users, err := models.ForEachUser()
	if err != nil {
		return nil, err
	}

	alarmConfig, err := models.GetAlarmConfig()
	if err != nil {
		return nil, err
	}

	return &RaftSnapshot{
		Version:     uint64(dateutil.NowMillisecond()),
		Job:         jobs,
		Triggered:   triggeredList,
		Node:        nodes,
		User:        users,
		AlarmConfig: alarmConfig,
	}, nil
}

func initRaftCommands() {

	logs.Infof("IsRaftFirstStart: %v", models.IsRaftFirstStart())

	if models.GetSnapshotVersion() > 0 {
		return
	}

	if !models.IsRaftFirstStart() {
		return
	}

	jobs, err := models.ForEachJob()
	if err == nil {
		for _, job := range jobs {
			SubmitCommand(&RaftCommand{
				Type: commandTypeInsertJob,
				Job:  job,
			})
		}
	}

	triggeredList, err := models.ForEachTriggered()
	if err == nil {
		for _, triggered := range triggeredList {
			SubmitCommand(&RaftCommand{
				Type:      commandTypeSaveTriggered,
				Triggered: triggered,
			})
		}
	}

	users, err := models.ForEachUser()
	if err == nil {
		for _, user := range users {
			SubmitCommand(&RaftCommand{
				Type: commandTypeSaveUser,
				User: user,
			})
		}
	}

	alarmConfig, err := models.GetAlarmConfig()
	if err == nil {
		SubmitCommand(&RaftCommand{
			Type:        commandTypeSaveAlarmConfig,
			AlarmConfig: alarmConfig,
		})
	}

	SubmitCommand(&RaftCommand{
		Type: commandTypeNegationRaftFirstStart,
	})

	log.Println("初始化数据准备完毕")
	logs.Infof("初始化数据准备完毕")

	models.NegationRaftFirstStart()
}
