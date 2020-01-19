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
package conf

import (
	"gojob/util/stringutil"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"gojob/models"
	"gojob/util/fileutil"
	"gojob/util/logs"
	"gojob/util/netutil"
	"gopkg.in/yaml.v2"
)

const (
	defHttpServerPort = 8080 // 默认HTTP服务端口
	defDataStorePath  = "store"
	defLogStorePath   = "log"
	defSignSecretKey  = "Go-Job-Key" // 默认签名秘钥
)

// 系统配置
var config *Config

// 集群配置
var clusterConfig *ClusterConfig

// 系统属性
type Config struct {
	DataStorePath      string                     `yaml:"data_store_dir"`        // 数据存储地址
	HttpServerBind     string                     `yaml:"http_server_bind"`      // 监听端口绑定的IP
	HttpServerPort     int                        `yaml:"http_server_port"`      // HTTP监听端口
	SignSecretKey      string                     `yaml:"sign_secret_key"`       // 签名秘钥
	ClusterNodeName    string                     `yaml:"cluster_node_name"`     // 集群节点名称
	ClusterNodeTcpPort int                        `yaml:"cluster_node_tcp_port"` // 集群节点TCP监听端口
	LoggerConfig       *logs.LoggerConfig         `yaml:"logger"`
	DataSourceConfig   []*models.DataSourceConfig `yaml:"datasource"`
}

type ClusterItemConfig struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
}

// 集群属性
type ClusterConfig struct {
	CurrentNodeName string
	CurrentHttpAddr string
	CurrentTcpAddr  string
	CurrentTcpPort  int
	Nodes           []*ClusterItemConfig `yaml:"nodes"` // 节点列表
}

// 加载应用配置
func InitConfig(filePath string) *Config {
	data, exist := ioutil.ReadFile(filePath)
	if exist != nil {
		log.Panicf("找不到配置文件:%s \n", filePath)
	}
	var temp Config
	err := yaml.Unmarshal(data, &temp)
	if err != nil {
		log.Panicf("配置文件解析失败:%s \n", err.Error())
	}
	if temp.DataStorePath == "" {
		temp.DataStorePath = filepath.Join(fileutil.GetCurrentDirectory(), defDataStorePath)
	}
	if err := fileutil.MkdirIfNecessary(temp.DataStorePath); err != nil {
		log.Panicf("数据存储目录创建失败,请检查data_store_path配置项 \n ")
	}
	if temp.DataSourceConfig == nil {
		log.Panicf("配置文件解析失败,未找到datasource配置 \n ")
	}
	for _, c := range temp.DataSourceConfig {
		if c == nil || c.DriverName == "" || c.DataSourceName == "" {
			log.Panicf("配置文件解析失败,datasource配置不正确 \n ")
		}
	}
	for i := 0; i < len(temp.DataSourceConfig); i++ {
		ds1 := temp.DataSourceConfig[i]
		for j := i + 1; j < len(temp.DataSourceConfig); j++ {
			ds2 := temp.DataSourceConfig[j]
			if ds1.DataSourceName == ds2.DataSourceName {
				log.Panicf("配置文件解析失败,datasource配置项中存在重复的DataSourceName \n ")
			}
		}
	}
	if temp.LoggerConfig == nil {
		log.Panicf("配置文件解析失败,未找到logger配置 \n ")
	}
	if temp.HttpServerPort == 0 {
		temp.HttpServerPort = defHttpServerPort
	}
	if temp.LoggerConfig.LogPath == "" {
		temp.LoggerConfig.LogPath = filepath.Join(temp.DataStorePath, defLogStorePath)
	}
	if temp.SignSecretKey == "" {
		temp.SignSecretKey = defSignSecretKey
	}

	config = &temp
	return config
}

func GetConfig() *Config {
	if config == nil {
		logs.Errorf("未初始化应用配置 \n")
		log.Panicf("未初始化应用配置 \n")
	}
	return config
}

// 初始化集群配置
func InitClusterConfig(path string) *ClusterConfig {
	if "" == config.ClusterNodeName {
		logs.Errorf("请正确填写cluster_node_name的属性值")
		log.Panic("请正确填写cluster_node_name的属性值")
	}
	if stringutil.IsChineseChar(config.ClusterNodeName) {
		logs.Errorf("请正确填写cluster_node_name的属性值,不支持中文")
		log.Panic("请正确填写cluster_node_name的属性值,不支持中文")
	}
	var ccf ClusterConfig
	data, exist := ioutil.ReadFile(path)
	if exist != nil {
		logs.Errorf("找不到集群配置文件:%s \n", path)
		log.Panicf("找不到集群配置文件:%s \n", path)
	}
	err := yaml.Unmarshal(data, &ccf)
	if err != nil {
		logs.Errorf("集群配置文件解析失败:%s \n", err.Error())
		log.Panicf("集群配置文件解析失败:%s \n", err.Error())
	}
	if ccf.Nodes == nil {
		logs.Errorf("集群配置文件解析失败,未找到nodes配置 \n ")
		log.Panicf("集群配置文件解析失败,未找到nodes配置 \n ")
	}
	if len(ccf.Nodes) < 3 {
		logs.Errorf("高可用raft集群至少需要三个节点组成 \n ")
		log.Panicf("高可用raft集群至少需要三个节点组成 \n ")
	}

	consign := make(map[string]bool)
	var currentHttpAddr string
	for _, node := range ccf.Nodes {
		_, nameExist := consign[node.Name]
		if nameExist {
			logs.Errorf("集群文件配置错误，存在重复的name属性值 \n ")
			log.Panicf("集群文件配置错误，存在重复的name属性值 \n ")
		}
		consign[node.Name] = true

		if !netutil.HostAddrCheck(node.Addr) {
			logs.Errorf("集群文件配置错误，addr属性：%s 不正确,请填写（IP:端口）格式的值 \n ", node.Addr)
			log.Panicf("集群文件配置错误，addr属性：%s 不正确,请填写（IP:端口）格式的值 \n ", node.Addr)
		}

		_, addrExist := consign[node.Addr]
		if addrExist {
			logs.Errorf("集群文件配置错误，存在重复的addr属性值 \n ")
			log.Panicf("集群文件配置错误，存在重复的addr属性值 \n ")
		}
		consign[node.Addr] = true

		if config.ClusterNodeName == node.Name {
			currentHttpAddr = node.Addr
		}
	}

	if "" == currentHttpAddr {
		logs.Errorf("集群文件配置错误，当前节点：%s 未存在nodes属性中 \n ", config.ClusterNodeName)
		log.Panicf("集群文件配置错误，当前节点：%s 未存在nodes属性中 \n ", config.ClusterNodeName)
	}
	currentHttpPort := strings.Split(currentHttpAddr, ":")[1]
	if currentHttpPort != strconv.Itoa(config.HttpServerPort) {
		logs.Errorf("集群文件配置错误，当前节点：%s 的addr属性值错误，请检查端口与application.yml中的http_server_port属性值是否一致 \n ", config.ClusterNodeName)
		log.Panicf("集群文件配置错误，当前节点：%s 的addr属性值错误，请检查端口与application.yml中的http_server_port属性值是否一致 \n ", config.ClusterNodeName)
	}

	ccf.CurrentNodeName = config.ClusterNodeName
	ccf.CurrentHttpAddr = currentHttpAddr
	ccf.CurrentTcpPort = config.ClusterNodeTcpPort
	clusterConfig = &ccf
	return clusterConfig
}

func GetClusterConfig() *ClusterConfig {
	if clusterConfig == nil {
		logs.Error("未初始化集群配置 \n")
		log.Panicf("未初始化集群配置 \n")
	}
	return clusterConfig
}
