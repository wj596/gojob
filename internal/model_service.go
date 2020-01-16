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
	"gojob/internal/icron"
	"gojob/models"
	"gojob/util/dateutil"
)

func InsertJob(job *models.Job) error {
	job.Id = GetSnowId()
	job.Status = models.JobStatusOk
	job.CreateTime = dateutil.NowMillisecond()
	job.TimeStep = icron.GetTimeStep(job.Cron)
	err := models.CascadeInsertJob(job)
	if err != nil {
		return err
	}

	err = addScheduler(job)
	if nil != err {
		return err
	}
	scheduleTask(job.Id)

	if IsClusterMode() {
		return SubmitCommand(&RaftCommand{
			Type: commandTypeInsertJob,
			Job:  job,
		})
	}

	return err
}

func DeleteJob(id uint64) error {
	cancelTask(id)

	err := models.DeleteJob(id)
	if err != nil {
		return err
	}

	if IsClusterMode() {
		return SubmitCommand(&RaftCommand{
			Type:     commandTypeDeleteJob,
			EntityId: id,
		})
	}

	return err
}

func UpdateJob(job *models.Job) error {
	cronChanged := false
	refer, _ := models.GetJob(job.Id)
	if refer.Cron != job.Cron {
		cronChanged = true
		job.TimeStep = icron.GetTimeStep(job.Cron)
	}

	err := models.UpdateJob(job)
	if err != nil {
		return err
	}

	if cronChanged {
		cancelTask(job.Id)
		err = addScheduler(job)
		if err == nil && models.JobStatusOk == refer.Status {
			scheduleTask(job.Id)
		}
	}

	if IsClusterMode() {
		return SubmitCommand(&RaftCommand{
			Type: commandTypeUpdateJob,
			Job:  job,
		})
	}

	return err
}

func UpdateJobStatus(id uint64, status int) error {
	job, err := models.GetJob(id)
	if err != nil {
		return err
	}
	job.Status = status
	err = models.UpdateJob(job)
	if err != nil {
		return err
	}

	if models.JobStatusPause == status {
		suspendTask(id)
	}
	if models.JobStatusOk == status {
		scheduleTask(id)
	}

	if IsClusterMode() {
		err = SubmitCommand(&RaftCommand{
			Type: commandTypeUpdateJob,
			Job:  job,
		})
	}

	return err
}

func updateTriggered(jobId uint64, prev int64, next int64) error {
	triggered, err := models.GetTriggered(jobId)
	if err != nil {
		return err
	}

	if prev != 0 || next != 0 {
		triggered.Times = triggered.Times + 1
	}
	triggered.PrevTime = prev
	triggered.NextTime = next
	err = models.SaveTriggered(triggered)
	if err != nil {
		return err
	}

	if IsClusterMode() {
		return SubmitCommand(&RaftCommand{
			Type:      commandTypeSaveTriggered,
			Triggered: triggered,
		})
	}

	return nil
}

func InsertNode(node *models.Node) error {
	err := models.InsertNode(node)
	if err != nil {
		return err
	}

	if IsClusterMode() {
		return SubmitCommand(&RaftCommand{
			Type: commandTypeSaveNode,
			Node: node,
		})
	}

	return err
}

func UpdateNode(node *models.Node) error {
	err := models.UpdateNode(node)
	if err != nil {
		return err
	}

	if IsClusterMode() {
		return SubmitCommand(&RaftCommand{
			Type: commandTypeSaveNode,
			Node: node,
		})
	}

	return err
}

func InsertUser(user *models.User) error {
	user.Id = GetSnowId()
	user.UpdateTime = dateutil.NowMillisecond()
	err := models.SaveUser(user)
	if err != nil {
		return err
	}

	if IsClusterMode() {
		return SubmitCommand(&RaftCommand{
			Type: commandTypeSaveUser,
			User: user,
		})
	}

	return err
}

func DeleteUser(id uint64) error {
	err := models.DeleteUser(id)
	if err != nil {
		return err
	}

	if IsClusterMode() {
		return SubmitCommand(&RaftCommand{
			Type:     commandTypeDeleteUser,
			EntityId: id,
		})
	}

	return err
}

func UpdateUser(ps *models.User) error {
	user, err := models.GetUser(ps.Name)
	if err != nil {
		return err
	}
	if ps.Password != "" {
		user.Password = ps.Password
	}
	if ps.Email != "" {
		user.Email = ps.Email
	}
	err = models.SaveUser(user)
	if err != nil {
		return err
	}

	if IsClusterMode() {
		return SubmitCommand(&RaftCommand{
			Type: commandTypeSaveUser,
			User: user,
		})
	}

	return err
}

func UpdateAlarmConfig(alarmConfig *models.AlarmConfig) error {
	err := models.SaveAlarmConfig(alarmConfig)
	if err != nil {
		return err
	}

	if IsClusterMode() {
		return SubmitCommand(&RaftCommand{
			Type:        commandTypeSaveAlarmConfig,
			AlarmConfig: alarmConfig,
		})
	}

	return err
}
