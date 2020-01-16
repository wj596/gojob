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
	"gojob/models"
	"gojob/util/netutil"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gojob/conf"
	"gojob/util/dateutil"
	"gojob/util/fileutil"
	"gojob/util/logs"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/vmihailenco/msgpack"
	"go.uber.org/atomic"
)

const (
	raftLogFileName   = "cluster.log"
	tcpMaxPool        = 3
	raftOptTimeout    = 10 * time.Second
	snapshotInterval  = 1800 * time.Second
	snapshotThreshold = 3
	trailingLogs      = 65535
	snapshotRetain    = 1
	detectNodeTimeout = 12 * time.Second
)

type raftCluster struct {
	conf        *raftClusterConfig
	raft        *raft.Raft
	fsm         raft.FSM //有限状态机
	transport   *raft.NetworkTransport
	logStore    raft.LogStore
	stableStore raft.StableStore
	leaderCutCh chan bool //leader切换通知
}

type raftClusterConfig struct {
	NodeName          string        // 节点名称
	TcpAddr           string        // TCP地址
	StorePath         string        //持久化目录
	LogStoreFile      string        // 日志数据文件名称
	StableStoreFile   string        // 状态数据文件名称
	TcpMaxPool        int           // TCP连接池最大值
	TcpTimeout        time.Duration // TCP超时时间
	SnapshotInterval  time.Duration //
	SnapshotThreshold uint64        //
	TrailingLogs      uint64        // 快照之后保留的日志条数
	SnapshotRetain    int           // 快照保留数量
}

func defaultRaftClusterConfig(clusterConfig *conf.ClusterConfig) *raftClusterConfig {
	storePath := filepath.Join(conf.GetConfig().DataStorePath, "raft")
	if err := fileutil.MkdirIfNecessary(storePath); err != nil {
		log.Panicf("blot存储目录:%s，创建失败 \n", storePath)
	}
	return &raftClusterConfig{
		NodeName:          clusterConfig.CurrentNodeName,
		TcpAddr:           clusterConfig.CurrentTcpAddr,
		StorePath:         storePath,
		LogStoreFile:      filepath.Join(storePath, "raft-log.bolt"),
		StableStoreFile:   filepath.Join(storePath, "raft-stable.bolt"),
		TcpMaxPool:        tcpMaxPool,
		TcpTimeout:        raftOptTimeout,
		SnapshotInterval:  snapshotInterval,
		SnapshotThreshold: snapshotThreshold,
		TrailingLogs:      trailingLogs,
		SnapshotRetain:    snapshotRetain,
	}
}

func newRaftCluster(clusterConfig *raftClusterConfig) (*raftCluster, error) {
	sysLogConf := conf.GetConfig().LoggerConfig
	logConf := &logs.LoggerConfig{
		Level:    sysLogConf.Level,
		LogPath:  sysLogConf.LogPath,
		LogFile:  filepath.Join(sysLogConf.LogPath, raftLogFileName),
		MaxSize:  sysLogConf.MaxSize,
		MaxAge:   sysLogConf.MaxAge,
		Compress: sysLogConf.Compress,
		Encoding: sysLogConf.Encoding,
	}
	stdLogOutput := logs.NewLumberjackLogger(logConf)
	stdLog := log.New(stdLogOutput, "", log.LstdFlags)

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(clusterConfig.NodeName)
	raftConf.Logger = hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Info,
		Output:     stdLogOutput,
		TimeFormat: dateutil.DayTimeSecondFormatter,
	})
	raftConf.SnapshotInterval = clusterConfig.SnapshotInterval
	raftConf.SnapshotThreshold = clusterConfig.SnapshotThreshold
	raftConf.TrailingLogs = clusterConfig.TrailingLogs
	leaderCutCh := make(chan bool, 1)
	raftConf.NotifyCh = leaderCutCh

	logStore, err := raftboltdb.NewBoltStore(clusterConfig.LogStoreFile)
	if err != nil {
		return nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(clusterConfig.StableStoreFile)
	if err != nil {
		return nil, err
	}

	snapshotStore, err := raft.NewFileSnapshotStoreWithLogger(clusterConfig.StorePath, 1, stdLog)
	if err != nil {
		return nil, err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", clusterConfig.TcpAddr)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransportWithLogger(tcpAddr.String(), tcpAddr, clusterConfig.TcpMaxPool, clusterConfig.TcpTimeout, stdLog)
	if err != nil {
		return nil, err
	}

	fsmImpl := new(FsmImpl)
	raft, err := raft.NewRaft(raftConf, fsmImpl, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	return &raftCluster{
		conf:        clusterConfig,
		raft:        raft,
		fsm:         fsmImpl,
		transport:   transport,
		logStore:    logStore,
		stableStore: stableStore,
		leaderCutCh: leaderCutCh,
	}, nil
}

func (this *raftCluster) bootstrap() {
	logs.Infof("bootstrap node : %s - %s", this.conf.NodeName, this.conf.TcpAddr)
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(this.conf.NodeName),
				Address: raft.ServerAddress(this.conf.TcpAddr),
			},
		},
	}
	this.raft.BootstrapCluster(configuration)
}

var leaderFlag atomic.Bool
var currentCluster *raftCluster
var leaderDetectCh chan string

// 引导集群
func BootstrapCluster(clusterConfig *conf.ClusterConfig) {
	if clusterConfig.CurrentTcpPort == 0 {
		lastTcpPort := models.GetLastTcpPort()
		if lastTcpPort == 0 {
			lastTcpPort = netutil.GetFreePort()
		}
		models.SetLastTcpPort(lastTcpPort)
		clusterConfig.CurrentTcpPort = lastTcpPort
	}
	clusterConfig.CurrentTcpAddr = strings.Split(clusterConfig.CurrentHttpAddr, ":")[0] + ":" + strconv.Itoa(clusterConfig.CurrentTcpPort)
	raftClusterConfig := defaultRaftClusterConfig(clusterConfig)

	lastNodeName := models.GetLastNodeName()
	if lastNodeName == "" {
		models.SetLastNodeName(clusterConfig.CurrentNodeName)
	} else {
		if lastNodeName != clusterConfig.CurrentNodeName {
			logs.Warn("检测到集群文件配置项cluster_node_name的值有变动，将重置节点的集群状态")
			log.Print("检测到集群文件配置项cluster_node_name的值有变动，将重置节点的集群状态")
			models.SetLastNodeName(clusterConfig.CurrentNodeName)
			os.Remove(raftClusterConfig.StableStoreFile)
			os.Remove(raftClusterConfig.LogStoreFile)
		}
	}
	logs.Infof("当前节点监听的TCP端口为：%v，请留意防火墙设置", clusterConfig.CurrentTcpPort)
	log.Printf("当前节点监听的TCP端口为：%v，请留意防火墙设置", clusterConfig.CurrentTcpPort)

	raftCluster, err := newRaftCluster(raftClusterConfig)
	if err != nil {
		logs.Errorf("集群节点初始化失败:%s. \n", err.Error())
		log.Panicf("集群节点初始化失败:%s. \n", err.Error())
	}
	currentCluster = raftCluster
	leaderDetectCh = make(chan string, 1)
	httpClient := &http.Client{}
	httpClient.Timeout = raftOptTimeout
	detectRaftNode(httpClient, clusterConfig)

	var leaderNode string
	select {
	case leaderNode = <-leaderDetectCh:
		log.Printf("探测到集群主节点为：%s，开始加入集群", leaderNode)
		joinCluster(httpClient, leaderNode, raftClusterConfig.NodeName, clusterConfig.CurrentHttpAddr, raftClusterConfig.TcpAddr)
	case <-time.After(detectNodeTimeout):
		log.Printf("未探测到集群主节点，引导集群创建")
		currentCluster.bootstrap()
	}
	// 启动集群状态监听器
	go startClusterStateListener()
}

func detectRaftNode(httpClient *http.Client, clusterConfig *conf.ClusterConfig) {
	for _, node := range clusterConfig.Nodes {
		if clusterConfig.CurrentNodeName == node.Name {
			continue
		}
		log.Printf("探测集群节点：%s - %s", node.Name, node.Addr)
		go func(inquiry string) {
			detectUrl, _ := url.Parse(fmt.Sprintf("http://%s/cluster/leader_id", inquiry))
			logs.Infof("detectNode URL : %s", detectUrl)
			timestamp := strconv.FormatInt(time.Now().Unix(), 10)
			sign := Signature(detectUrl.RequestURI() + timestamp)
			request, err := http.NewRequest(http.MethodGet, detectUrl.String(), nil)
			request.Header.Add("X-Timestamp", timestamp)
			request.Header.Add("X-Sign", sign)
			if nil != err {
				logs.Error(err.Error())
				return
			}

			res, err := httpClient.Do(request)
			if nil != err {
				logs.Error(err.Error())
				return
			}
			defer res.Body.Close()
			if 200 != res.StatusCode {
				logs.Warnf("detectNode %s StatusCode: %v", inquiry, res.StatusCode)
				if http.StatusUnauthorized == res.StatusCode {
					logs.Errorf("加入集群错误：请检查配置文件中'sign_secret_key'属性，是否与其他集群节点一致")
					log.Panicf("加入集群错误：请检查配置文件中'sign_secret_key'属性，是否与其他集群节点一致")
				}
				return
			}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				logs.Error(err.Error())
				return
			}
			leaderId := string(body)
			if leaderId == "" {
				return
			}
			nodeInfo, err := GetRuntimeClusterNode(leaderId)
			if err != nil {
				log.Panicf("集群文件配置错误，节点：%s 未存在nodes属性中 \n ", leaderId)
			}

			select {
			case leaderDetectCh <- string(nodeInfo.HttpAddr):
			default:
			}
		}(node.Addr)
	}
}

func joinCluster(httpClient *http.Client, masterNode string, peerNodeName string, peerHttpAddr string, peerTcpAddr string) {
	joinUrl, _ := url.Parse(fmt.Sprintf("http://%s/cluster/join/%s/%s/%s", masterNode, peerNodeName, peerHttpAddr, peerTcpAddr))
	logs.Infof("joinCluster URL : %s", joinUrl)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sign := Signature(joinUrl.RequestURI() + timestamp)
	request, err := http.NewRequest(http.MethodGet, joinUrl.String(), nil)
	request.Header.Add("X-Timestamp", timestamp)
	request.Header.Add("X-Sign", sign)
	if nil != err {
		logs.Errorf("加入集群错误：%s", err.Error())
		log.Panicf("加入集群错误：%s", err.Error())
		return
	}

	res, err := httpClient.Do(request)
	if nil != err {
		logs.Errorf("加入集群错误：%s", err.Error())
		log.Panicf("加入集群错误：%s", err.Error())
		return
	}
	defer res.Body.Close()
	if http.StatusOK != res.StatusCode {
		logs.Errorf("加入集群错误,StatusCode：%d", res.StatusCode)
		log.Panicf("加入集群错误,StatusCode：%d", res.StatusCode)
		return
	}
	logs.Infof("成功加入主节点为：%s 的集群", masterNode)
	log.Printf("成功加入主节点为：%s 的集群", masterNode)
}

func IsLeader() bool {
	return leaderFlag.Load()
}

func getRaftServers() []raft.Server {
	future := currentCluster.raft.GetConfiguration()
	if err := future.Error(); err == nil {
		configuration := future.Configuration()
		return configuration.Servers
	}
	return nil
}

func GetLeaderId() string {
	if IsLeader() {
		return conf.GetClusterConfig().CurrentNodeName
	}

	leaderAddress := currentCluster.raft.Leader()
	servers := getRaftServers()
	if nil != servers {
		for _, s := range servers {
			if s.Address == leaderAddress {
				return string(s.ID)
			}
		}
	}

	node, err := models.GetNodeByTcpAddr(string(leaderAddress))
	if err == nil {
		return node.Name
	}

	return ""
}

func GetLeaderServerAddress() string {
	return string(currentCluster.raft.Leader())
}

func GetRaftNodeDetails() []map[string]string {
	nodes := make([]map[string]string, 0)
	servers := getRaftServers()
	if nil != servers {
		for _, s := range servers {
			node := make(map[string]string)
			node["nodeName"] = string(s.ID)
			node["tcpAddr"] = string(s.Address)
			node["suffrage"] = fmt.Sprintf("%+v", s.Suffrage)
			node["lastContact"] = ""
			node["allowOffline"] = "0"
			node["status"] = "-1"
			if currentCluster.raft.Leader() == s.Address {
				node["status"] = "2"
			} else {
				if currentCluster.conf.TcpAddr == string(s.Address) {
					node["status"] = strconv.FormatInt(int64(currentCluster.raft.State()), 10)
				}
			}
			rs, ok := currentCluster.raft.GetReplState()[s.ID]
			if ok {
				node["status"] = "0"
				node["lastContact"] = dateutil.DefaultLayout(rs.LastContact())
				lastContact := rs.LastContact()
				diff := time.Now().Unix() - lastContact.Unix()
				if diff >= 30 {
					node["allowOffline"] = "1"
				}
			}
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func getFollowers() map[string]bool {
	followers := make(map[string]bool)
	servers := getRaftServers()
	if nil != servers {
		for _, s := range servers {
			rs, ok := currentCluster.raft.GetReplState()[s.ID]
			if ok {
				lastContact := rs.LastContact()
				diff := time.Now().Unix() - lastContact.Unix()
				node := string(s.ID) + "/" + string(s.Address)
				effective := true
				if diff >= 30 { // 超过30秒无心跳
					effective = false
				}
				followers[node] = effective
			}
		}
	}
	return followers
}

func getNodeAmount() (int, int) {
	var okAmount int
	var noAmount int
	servers := getRaftServers()
	if nil != servers {
		for _, s := range servers {
			if currentCluster.raft.Leader() == s.Address {
				okAmount = okAmount + 1
			}
			rs, ok := currentCluster.raft.GetReplState()[s.ID]
			if ok {
				lastContact := rs.LastContact()
				diff := time.Now().Unix() - lastContact.Unix()
				if diff >= 30 {
					noAmount = noAmount + 1
				} else {
					okAmount = okAmount + 1
				}
			}
		}
	}
	return okAmount, noAmount
}

func AddVoter(nodeName string, tcpAddr string) error {
	logs.Infof("节点:%s - %s，加入集群", nodeName, tcpAddr)
	future := currentCluster.raft.AddVoter(raft.ServerID(nodeName), raft.ServerAddress(tcpAddr), 0, raftOptTimeout)
	if err := future.Error(); err != nil {
		return err
	}
	return nil
}

func RemovePeer(nodeName string) error {
	logs.Infof("节点:%s，从集群中迁出", nodeName)
	future := currentCluster.raft.RemoveServer(raft.ServerID(nodeName), 0, raftOptTimeout)
	if err := future.Error(); err != nil {
		return err
	}
	return nil
}

func SubmitCommand(command *RaftCommand) error {
	logs.Infof("SubmitCommand Type:%v", command.Type)
	cmd, err := msgpack.Marshal(command)
	if err != nil {
		return err
	}
	applyFuture := currentCluster.raft.Apply(cmd, raftOptTimeout)
	if err := applyFuture.Error(); err != nil {
		return err
	}
	return nil
}

func startClusterStateListener() {
	logs.Info("启动 集群状态监听器")
	for {
		select {
		case leader := <-currentCluster.leaderCutCh:
			if leader {
				log.Printf("当前节点的集群状态为：Leader")
				logs.Info("当前节点的集群状态为：Leader")
				leaderFlag.Store(true)
				initRaftCommands()
				InitSnowflake()
				InitSchedulers()
			} else {
				status := "Follower"
				if currentCluster.raft.State() == 1 {
					status = "Candidate"
				}
				log.Printf("当前节点的集群状态为：%s", status)
				logs.Infof("当前节点的集群状态为：%s", status)
				leaderFlag.Store(false)
				deleteSchedulers()
			}
		}
	}
}
