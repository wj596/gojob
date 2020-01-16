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
	"log"
	"time"

	"gojob/util/byteutil"
	"gojob/util/logs"

	"github.com/boltdb/bolt"
	"github.com/go-gomail/gomail"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
)

// 告警设置
type AlarmConfig struct {
	SysAlarmEmail string `json:"sysAlarmEmail"`
	SmtpHost      string `json:"smtpHost"`
	SmtpPort      int    `json:"smtpPort"`
	SmtpUser      string `json:"smtpUser"`
	SmtpPassword  string `json:"smtpPassword"`
}

// 告警邮件
type AlarmEmail struct {
	Toers   string
	Subject string
	Body    string
}

var fixAlarmId = byteutil.Uint64ToBytes(uint64(1))
var alarmEmailQueue = make(chan *AlarmEmail, 65535)
var mailDialer *gomail.Dialer

func InitAlarm() {
	if _, err := GetAlarmConfig(); err != nil {
		SaveAlarmConfig(new(AlarmConfig))
	}
	initMailDialer()
	startAlarmQueueListener()
}

func initMailDialer() {
	conf, _ := GetAlarmConfig()
	if conf.SmtpHost == "" || conf.SmtpPort == 0 || conf.SmtpUser == "" || conf.SmtpPassword == "" {
		logs.Warn("请注意：当前无法发送告警邮件,请启动系统后在'告警设置'模块中进行相关属性的配置")
		log.Printf("请注意：当前无法发送告警邮件,请启动系统后在'告警设置'模块中进行相关属性的配置")
		mailDialer = nil
		return
	}
	mailDialer = gomail.NewDialer(conf.SmtpHost, conf.SmtpPort, conf.SmtpUser, conf.SmtpPassword)
}

func TestMailDialer(target string) error {
	conf, _ := GetAlarmConfig()
	if conf.SmtpHost == "" || conf.SmtpPort == 0 || conf.SmtpUser == "" || conf.SmtpPassword == "" {
		logs.Errorf("当前无法发送告警邮件,请在'告警设置'模块中进行相关属性的配置")
		mailDialer = nil
		return errors.Errorf("告警设置不正确")
	}
	mailDialer = gomail.NewDialer(conf.SmtpHost, conf.SmtpPort, conf.SmtpUser, conf.SmtpPassword)
	mail := gomail.NewMessage()
	mail.SetHeader("From", mailDialer.Username)
	mail.SetHeader("To", target)
	mail.SetHeader("Subject", "Go-Job系统邮箱配置测试邮件")
	mail.SetBody("text/html", "<b>Go-Job系统邮箱配置测试邮件，收到此邮件说明'告警设置'配置正确</b>")
	return mailDialer.DialAndSend(mail)
}

func SaveAlarmConfig(entity *AlarmConfig) error {
	err := GetBoltDB().Update(func(tx *bolt.Tx) error {
		bt := tx.Bucket(alarmConfigBucket)
		bs, err := msgpack.Marshal(entity)
		if err != nil {
			return err
		}
		return bt.Put(fixAlarmId, bs)
	})
	return err
}

func GetAlarmConfig() (*AlarmConfig, error) {
	var val []byte
	err := GetBoltDB().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(alarmConfigBucket)
		val = bucket.Get(fixAlarmId)
		if val == nil {
			return errors.Errorf("Key Not Found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var entity = new(AlarmConfig)
	err = msgpack.Unmarshal(val, entity)
	return entity, err
}

func SendAlarmEmail(alarm *AlarmEmail) {
	logs.Infof("发送告警邮件：%s", alarm.Subject)
	if mailDialer == nil {
		logs.Warnf("当前无法发送告警邮件,请在'告警设置'模块中进行相关属性的配置")
		return
	}
	alarmEmailQueue <- alarm
}

func startAlarmQueueListener() {
	logs.Info("启动 告警队列监听器")
	go func() {
		for {
			alarm := <-alarmEmailQueue
			mail := gomail.NewMessage()
			mail.SetHeader("From", mailDialer.Username)
			mail.SetHeader("To", alarm.Toers)
			mail.SetHeader("Subject", alarm.Subject)
			mail.SetBody("text/html", alarm.Body)
			if nil != mailDialer {
				for i := 0; i < 3; i++ {
					err := mailDialer.DialAndSend(mail)
					if err == nil {
						logs.Infof("告警邮件：'%s' 发送成功", alarm.Subject)
						break
					}
					time.Sleep(time.Second)
					logs.Errorf("告警邮件：'%s' 发送失败：%s", alarm.Subject, err.Error())
				}
			} else {
				logs.Error("告警邮件发送失败，请正确设置系统邮箱属性")
			}
		}
	}()
}
