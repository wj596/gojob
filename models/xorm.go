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
	"fmt"
	"log"
	"strings"
	"sync"

	"gojob/util/dateutil"
	"gojob/util/logs"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"go.uber.org/atomic"
)

const tablePrefix = "t_"

var currentDBName string
var redundancyMap sync.Map
var redundancyMapSize atomic.Int32
var invalidMap sync.Map
var alarmedMap sync.Map
var hasWaitRecoverDb atomic.Bool
var dbCutLock sync.Mutex

// 持久化属性
type DataSourceConfig struct {
	DriverName     string `yaml:"driver_name"`      //数据库驱动名称
	DataSourceName string `yaml:"data_source_name"` //数据源名称
	ShowSQL        bool   `yaml:"show_sql"`         // 是否打印SQL
}

type redundancy struct {
	name    string
	mixName string
	engine  *xorm.Engine
}

func getMixDataSourceName(dataSourceName string) string {
	temp := strings.Split(dataSourceName, "@")
	if len(temp) > 1 {
		return "***@" + temp[1]
	}
	return dataSourceName
}

func InitXorm(configs []*DataSourceConfig) {
	for _, config := range configs {
		ds := new(redundancy)
		ds.engine = newXormEngine(config)
		ds.name = config.DataSourceName
		ds.mixName = getMixDataSourceName(config.DataSourceName)
		err := createTraceTableNecessary(ds.engine)
		if err != nil {
			logs.Errorf("创建表失败，您可以使用数据库初始化SQL自行建表: %s \n", err.Error())
			log.Panicf("创建表失败，您可以使用数据库初始化SQL自行建表: %s \n", err.Error())
		}
		redundancyMap.Store(ds.name, ds)
		redundancyMapSize.Add(1)
	}
	if redundancyMapSize.Load() > 1 {
		var lastUpdateTime int64
		redundancyMap.Range(func(key, value interface{}) bool {
			r := value.(*redundancy)
			max := selectMaxStartTime(r.engine)
			if max > lastUpdateTime {
				lastUpdateTime = max
				currentDBName = r.name
			}
			return true
		})
	}
	if currentDBName == "" {
		currentDBName = configs[0].DataSourceName
	}
	log.Printf("MySQL数据连接成功,共 %v 个数据库节点", redundancyMapSize.Load())
	logs.Infof("MySQL数据连接成功,共 %v 个数据库节点,当前数据库位：%s", redundancyMapSize.Load(), getMixDataSourceName(currentDBName))
	startTraceSyncQueueListener()
}

func isRedundancy() bool {
	return redundancyMapSize.Load() > 1
}

func getCurrentDB() *redundancy {
	r, _ := redundancyMap.Load(currentDBName)
	return r.(*redundancy)
}

func tryCutDB() bool {
	dbCutLock.Lock()
	defer dbCutLock.Unlock()

	isCut := false
	if ok, err := pingDB(GetOrm()); !ok { // 当前数据库不可用
		logs.Warnf("数据库：%s,无法链接:%s ", getMixDataSourceName(currentDBName), err.Error())
		invalidMap.Store(getCurrentDB().name, true)
		hasWaitRecoverDb.Store(true)
		redundancyMap.Range(func(key, value interface{}) bool {
			r := value.(*redundancy)
			if getCurrentDB().name != r.name {
				if ok, _ := pingDB(r.engine); ok {
					invalidMap.Delete(r.name)
					currentDBName = r.name
					isCut = true
					logs.Infof("切换数据库：%s", r.mixName)
					return false
				} else {
					invalidMap.Store(r.name, true)
				}
			}
			return true
		})
	}
	return isCut
}

func pingDB(engine *xorm.Engine) (bool, error) {
	err := engine.Ping()
	if err != nil {
		return false, err
	}
	return true, nil
}

func GetDBAmount() (int, int) {
	var okAmount int
	var noAmount int

	redundancyMap.Range(func(key, value interface{}) bool {
		r := value.(*redundancy)
		if ok, _ := pingDB(r.engine); ok {
			invalidMap.Delete(r.name)
			okAmount = okAmount + 1
		} else {
			invalidMap.Store(r.name, true)
			noAmount = noAmount + 1
		}
		return true
	})
	return okAmount, noAmount
}

// DB告警
func DBAlarmNecessary(sysAlarmEmail string) {
	redundancyMap.Range(func(key, value interface{}) bool {
		r := value.(*redundancy)
		ok, err := pingDB(r.engine)
		if ok {
			alarmedMap.Delete(r.name)
		} else {
			if _, exist := alarmedMap.Load(r.name); !exist {
				alarmedMap.Store(r.name, true)
				logs.Warnf("数据库：%s，告警", r.mixName)
				SendAlarmEmail(&AlarmEmail{
					Toers:   sysAlarmEmail,
					Subject: fmt.Sprintf("Go-Job告警,数据库故障"),
					Body:    fmt.Sprintf("告警时间：%s  <br>数据库：%s ，无法链接 <br>详情：%s ", dateutil.NowFormatted(), r.mixName, err.Error()),
				})
			} else {
				logs.Warnf("数据库：%s，未恢复，已告警", r.mixName)
			}
		}
		return true
	})
}

func GetOrm() *xorm.Engine {
	return getCurrentDB().engine
}

func isDBInvalid(name string) bool {
	_, ok := invalidMap.Load(name)
	return ok
}

// 查询条件
type Condition struct {
	parameters map[string]interface{}
}

func NewCondition() *Condition {
	return &Condition{
		parameters: make(map[string]interface{}),
	}
}

func (this *Condition) AddParam(key string, val interface{}) *Condition {
	this.parameters[key] = val
	return this
}

func (this *Condition) AddParamNecessary(need bool, key string, val interface{}) *Condition {
	if need {
		this.parameters[key] = val
	}
	return this
}

func (this *Condition) ExistParam(key string) bool {
	_, exist := this.parameters[key]
	return exist
}

func (this *Condition) GetParam(key string) interface{} {
	val, _ := this.parameters[key]
	return val
}

func (this *Condition) GetStringParam(key string) string {
	val, exist := this.parameters[key]
	if exist {
		return val.(string)
	}
	return ""
}

func (this *Condition) GetIntParam(key string) int {
	val, exist := this.parameters[key]
	if exist {
		return val.(int)
	}
	return 0
}

func newXormEngine(config *DataSourceConfig) *xorm.Engine {
	engine, err := xorm.NewEngine(config.DriverName, config.DataSourceName)
	if err != nil {
		log.Panicf("创建XormEngine失败: %s \n", err.Error())
	}
	engine.SetLogLevel(core.LOG_WARNING)
	if err := engine.Ping(); err != nil {
		logs.Errorf("连接数据库:%s, 失败: %s \n", getMixDataSourceName(config.DataSourceName), err.Error())
		log.Panicf("连接数据库:%s, 失败: %s \n", getMixDataSourceName(config.DataSourceName), err.Error())
	}
	engine.ShowSQL(false)
	tbMapper := core.NewPrefixMapper(core.SnakeMapper{}, tablePrefix)
	engine.SetTableMapper(tbMapper)
	return engine
}

// 分页查询条件
type Page struct {
	condition *Condition  // 查询条件
	Limit     int         // 每页显示条数，默认 10
	Current   int         // 当前页
	Pages     int         // 总页数
	Total     int64       // 总条数
	Data      interface{} // 查询结果
}

func NewPage(current int, limit int) *Page {
	if current == 0 {
		current = 1
	}
	if limit == 0 {
		limit = 20
	}
	return &Page{
		Current:   current,
		Limit:     limit,
		condition: NewCondition(),
	}
}

func (this *Page) GetStartRow() int {
	return (this.Current - 1) * this.Limit
}

func (this *Page) SetCondition(c *Condition) *Page {
	this.condition = c
	return this
}

func (this *Page) AddParam(key string, val interface{}) *Page {
	this.condition.AddParam(key, val)
	return this
}

func (this *Page) AddParamNecessary(need bool, key string, val interface{}) *Page {
	this.condition.AddParamNecessary(need, key, val)
	return this
}

func (this *Page) ExistParam(key string) bool {
	return this.condition.ExistParam(key)
}

func (this *Page) GetParam(key string) interface{} {
	return this.condition.GetStringParam(key)
}

func (this *Page) GetStringParam(key string) string {
	return this.condition.GetStringParam(key)
}

func (this *Page) GetIntParam(key string) int {
	return this.condition.GetIntParam(key)
}
