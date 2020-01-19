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
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gojob/conf"
	"gojob/internal"
	"gojob/models"
	"gojob/routes"
	"gojob/util/logs"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	ac := flag.String("ac", "application.yml", "application config file path")
	cc := flag.String("cc", "cluster.yml", "cluster config file path")
	mode := flag.String("m", "standalone", "running mode standalone or cluster")

	flag.Usage = usage
	flag.Parse()
	log.Print("Go-Job 启动")
	config := conf.InitConfig(*ac)
	logs.InitLogger(config.LoggerConfig)
	internal.SetRunMode(*mode)
	logs.Infof("配置文件：%s", *ac)
	logs.Infof("运行模式：%s", *mode)
	log.Printf("配置文件：%s", *ac)
	log.Printf("运行模式：%s", *mode)
	models.InitBoltDB(config.DataStorePath)
	models.CreateDefaultUserIfNecessary()
	models.InitXorm(config.DataSourceConfig)
	models.InitAlarm()
	if internal.IsClusterMode() {
		internal.BootstrapCluster(conf.InitClusterConfig(*cc))
	} else { // 单机
		internal.InitSnowflake()
		internal.InitSchedulers()
	}
	internal.StartMonitorTask()
	routes.StartCertificateClearTask()
	routes.StartRouter(config.HttpServerBind, config.HttpServerPort)
}

func usage() {
	fmt.Fprintf(os.Stderr, `gojob v1.0.0
Usage: gojob [-c filename] [-m standalone/cluster]
Options:
  -ac string
        application config file path (default "application.yml")
  -cc string
        cluster config file path (default "cluster.yml")
  -m string
        running mode standalone or cluster (default "standalone")
`)
}
