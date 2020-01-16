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
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"gojob/internal/bl"
	"gojob/models"
	"gojob/util/dateutil"
	"gojob/util/httputil"
	"gojob/util/logs"
	"gojob/util/stringutil"
	"gojob/util/syncutil"
)

// 执行节点
type executeNode struct {
	address   string // 访问地址
	parameter string // 参数
	weight    int
}

// HTTP任务
type HttpTask struct {
	jobId      uint64 // JOB主键
	httpClient *httputil.HttpClient
}

// 创建任务
func newTask(job *models.Job) *HttpTask {
	retryConditionResponseNil := func(res *http.Response) bool {
		return res == nil
	}
	retryConditionStatusNot200 := func(res *http.Response) bool {
		return res.StatusCode != 200
	}
	client := httputil.NewHttpClient().
		AddRetryCondition(retryConditionResponseNil).
		AddRetryCondition(retryConditionStatusNot200)
	return &HttpTask{
		jobId:      job.Id,
		httpClient: client,
	}
}

func (this *HttpTask) Run() {
	if IsStandaloneOrLeader() {
		job, err := models.GetJob(this.jobId)
		if err != nil {
			logs.Errorf("任务调度失败,查找Job信息错误:%s", err.Error())
			return
		}

		sch, exist := getScheduler(this.jobId)
		if !exist {
			logs.Errorf("任务调度失败,未找到Job(%v)的调度器:%s", this.jobId)
			return
		}

		startTime := time.Now().Unix()
		nextTime := sch.GetNextTime()

		updateTriggered(this.jobId, startTime, nextTime)

		ctx := &scheduleContext{
			job:          job,
			scheduleType: models.ScheduleTypeAuto,
			startTime:    startTime,
			details:      make([]string, 0),
		}
		this.doRun(ctx)

		scanMisfires()
	}
}

// http任务run
func (this *HttpTask) doRun(ctx *scheduleContext) {
	executeNodes := make([]*executeNode, 0)
	if len(ctx.job.Executors) > 0 {
		for _, v := range ctx.job.Executors {
			if models.ExecutorStatusOk == v.Status {
				executeNodes = append(executeNodes, &executeNode{
					address: v.Address,
					weight:  v.Weight,
				})
			}
		}
	}
	if len(executeNodes) == 0 {
		logs.Errorf("任务调度失败，Job(%s)无执行节点", ctx.job.Name)
		ctx.failed("无执行节点")
		return
	}
	if IsClusterMode() {
		ctx.detail(fmt.Sprintf("调度节点：%s - %s", GetLeaderId(), GetLeaderServerAddress()))
	}
	ctx.detail(fmt.Sprintf("执行节点数量：%d，执行节点选择策略：%s", len(executeNodes), ctx.job.ExecutorSelectStrategy))
	if models.ExecutorSelectStrategySharding == ctx.job.ExecutorSelectStrategy {
		shardingResults := this.shardingExecutors(ctx, executeNodes)
		if len(shardingResults) == 0 {
			ctx.failed("执行节点分片错误")
			return
		}

		var wg sync.WaitGroup
		failedNodes := syncutil.NewMutexSlice()
		for _, shardingResult := range shardingResults {
			wg.Add(1)
			go func(exeNode *executeNode) {
				url := this.buildRequestUrl(ctx, exeNode)
				succeed := this.doExecute(ctx, url)
				if !succeed {
					failedNodes.Add(exeNode)
				}
				wg.Done()
			}(shardingResult)
		}
		wg.Wait()

		takeoverSucceed := true
		if models.FailTakeoverEnabled == ctx.job.FailTakeover && failedNodes.Size() > 0 && len(executeNodes) > failedNodes.Size() {
			takeoverSucceed = this.shardingTakeover(ctx, executeNodes, failedNodes)
		}
		if takeoverSucceed {
			ctx.succeed()
		} else {
			ctx.failed("执行失败")
		}
	} else { // 非分片执行
		selected := this.selectExecutor(ctx, executeNodes)
		ctx.detail(fmt.Sprintf("选中执行节点：%s", selected.address))
		executeUrl := this.buildRequestUrl(ctx, selected)
		succeed := this.doExecute(ctx, executeUrl)
		if models.FailTakeoverEnabled == ctx.job.FailTakeover && !succeed && len(executeNodes) > 1 {
			succeed = this.standaloneTakeover(ctx, selected, executeNodes)
		}
		if succeed {
			ctx.succeed()
		} else {
			ctx.failed("执行失败")
		}
	}
}

func (this *HttpTask) buildRequestUrl(ctx *scheduleContext, executeNode *executeNode) string {
	base := ctx.job.Protocol + "://" + executeNode.address
	if strings.HasPrefix(ctx.job.Uri, "/") {
		base = base + ctx.job.Uri
	} else {
		base = base + "/" + ctx.job.Uri
	}

	params := stringutil.KVsToMap(ctx.job.HttpParam, "|")
	if ctx.job.ShardingCount > 0 && "" != executeNode.parameter {
		params["sharding"] = executeNode.parameter
	}
	executeUrl := stringutil.BuildQueryString(base, params)
	return executeUrl
}

// 根据分片策进行执行器分片
func (this *HttpTask) shardingExecutors(ctx *scheduleContext, executeNodes []*executeNode) []*executeNode {
	ctx.detail(fmt.Sprintf("分片数量:%d", ctx.job.ShardingCount))
	var params []string
	if ctx.job.ShardingParam != "" {
		params = strings.Split(ctx.job.ShardingParam, ",")
	} else {
		params = make([]string, ctx.job.ShardingCount)
		for i := 0; i < ctx.job.ShardingCount; i++ {
			params[i] = strconv.Itoa(i)
		}
	}
	sharding := bl.Sharding(len(params), len(executeNodes))

	shardingResults := make([]*executeNode, 0)
	for i, v := range sharding {
		if len(v) > 0 {
			temp := make([]string, len(v))
			for j, vv := range v {
				temp[j] = params[vv]
			}
			result := &executeNode{
				address:   executeNodes[i].address,
				parameter: strings.Join(temp, ","),
			}
			logs.Infof("分片结果: %s - %s", result.address, result.parameter)
			ctx.detail(fmt.Sprintf("分片结果: %s - %s", result.address, result.parameter))
			shardingResults = append(shardingResults, result)
		}
	}

	return shardingResults
}

func (this *HttpTask) doExecute(ctx *scheduleContext, doUrl string) bool {
	ctx.mutexDetail(fmt.Sprintf("开始执行HTTP请求，URL为：%s", doUrl))

	this.httpClient.SetTimeout(ctx.job.Timeout).
		SetRetryCount(ctx.job.RetryCount).
		SetRetryWaitTime(ctx.job.RetryWaitTime)
	request := this.httpClient.NewRequest()
	if "" != ctx.job.HttpHeaderParam {
		params := stringutil.KVsToMap(ctx.job.HttpHeaderParam, "|")
		for k, v := range params {
			request.AddHeader(k, v)
		}
	}
	if models.HttpSignEnabled == ctx.job.HttpSign {
		requestUrl, _ := url.Parse(doUrl)
		timestamp := strconv.FormatInt(dateutil.NowMillisecond(), 10)
		base := requestUrl.RequestURI() + timestamp
		ctx.mutexDetail(fmt.Sprintf("开始数字签名，被签名字符串为：%s", base))
		sign := Signature(base)
		request.AddHeader("X-Timestamp", timestamp)
		request.AddHeader("X-Sign", sign)
	}
	res, err := request.Get(doUrl)
	if nil != err {
		logs.Errorf("Job(%s) HTTP请求错误：%s", ctx.job.Name, err.Error())
		ctx.mutexDetail(fmt.Sprintf("HTTP请求错误：%s", err.Error()))
		return false
	}
	defer res.Body.Close()
	if 200 != res.StatusCode {
		logs.Errorf("Job(%s) HTTP请求错误StatusCode：%v", ctx.job.Name, res.StatusCode)
		ctx.mutexDetail(fmt.Sprintf("HTTP请求错误StatusCode：%v", res.StatusCode))
		return false
	}
	ctx.mutexDetail("HTTP请求成功")
	return true
}

// 根据策略选择执行器
func (this *HttpTask) selectExecutor(ctx *scheduleContext, executeNodes []*executeNode) *executeNode {
	weightItems := make([]bl.LoadItem, len(executeNodes))
	selected := -1
	for i, v := range executeNodes {
		weightItems[i] = bl.LoadItem{
			Index:  i,
			Weight: v.weight,
		}
	}
	switch ctx.job.ExecutorSelectStrategy {
	case models.ExecutorSelectStrategyRandom:
		selected = randomLoadBalance.DoSelect(weightItems)
	case models.ExecutorSelectStrategyRound:
		lb, exist := roundLoadBalances[ctx.job.Id]
		if exist {
			selected = lb.DoSelect(weightItems)
		}
	case models.ExecutorSelectStrategyWeightRandom:
		selected = weightRandomLoadBalance.DoSelect(weightItems)
	case models.ExecutorSelectStrategyWeightRound:
		lb, exist := weightRoundLoadBalances[ctx.job.Id]
		if exist {
			selected = lb.DoSelect(weightItems)
		}
	}

	if -1 == selected {
		return nil
	}
	return executeNodes[selected]
}

// 分片故障转移
func (this *HttpTask) shardingTakeover(ctx *scheduleContext, executeNodes []*executeNode, failedNodes *syncutil.MutexSlice) bool {
	remains := make([]*executeNode, 0)
	for _, v := range executeNodes {
		include := false
		for i := 0; i < failedNodes.Size(); i++ {
			failedNode := failedNodes.Get(i).(*executeNode)
			if v.address == failedNode.address {
				include = true
				break
			}
		}
		if !include {
			remains = append(remains, v)
		}
	}
	if len(remains) == 0 {
		return false
	}

	var succeeds int
	for i := 0; i < failedNodes.Size(); i++ {
		index := i % len(remains)
		selected := remains[index]
		failedNode := failedNodes.Get(i).(*executeNode)
		executeUrl := this.buildRequestUrl(ctx, &executeNode{
			address:   selected.address,
			parameter: failedNode.parameter, //错误节点的分片数据
		})
		succeed := this.doExecute(ctx, executeUrl)
		if succeed {
			succeeds = succeeds + 1
			logs.Infof("Job(%s)失败转移,失败节点:%s,转移节点:%s", ctx.job.Name, failedNode.address, selected.address)
			ctx.detail(fmt.Sprintf("失败转移,失败节点:%s,转移节点:%s", failedNode.address, selected.address))
			break
		} else {
			for _, vvv := range remains {
				if selected.address != vvv.address {
					retryUrl := this.buildRequestUrl(ctx, &executeNode{
						address:   vvv.address,
						parameter: failedNode.parameter, //错误节点的分片数据
					})
					if this.doExecute(ctx, retryUrl) {
						logs.Infof("Job(%s)失败转移,失败节点:%s,转移节点:%s", ctx.job.Name, failedNode.address, vvv.address)
						ctx.detail(fmt.Sprintf("失败转移,失败节点:%s,转移节点:%s", failedNode.address, vvv.address))
						succeeds = succeeds + 1
						break
					}
				}
			}
		}
	}

	return failedNodes.Size() == succeeds
}

// 单节点故障转移
func (this *HttpTask) standaloneTakeover(ctx *scheduleContext, failedNode *executeNode, executeNodes []*executeNode) bool {
	failedList := make([]*executeNode, 0)
	failedList = append(failedList, failedNode)
	for i := 0; i < len(executeNodes)-1; i++ {
		remain := remainExecutor(executeNodes, failedList)
		if remain == nil {
			return false
		}
		logs.Infof("Job(%s) 开始失败转移,失败节点:%s,转移节点:%s", ctx.job.Name, failedNode.address, remain)
		ctx.detail(fmt.Sprintf("开始故障转移,失败节点:%s,转移节点:%s", failedNode.address, remain.address))
		executeUrl := this.buildRequestUrl(ctx, remain)
		takeoverSucceed := this.doExecute(ctx, executeUrl)
		if takeoverSucceed {
			ctx.detail("转移执行成功")
			return true
		} else {
			failedList = append(failedList, remain)
			ctx.detail("转移执行失败")
		}
	}
	return false
}

func remainExecutor(executeNodes []*executeNode, excludes []*executeNode) *executeNode {
	weightItems := make([]bl.LoadItem, 0)
	for index, v := range executeNodes {
		include := false
		for _, vv := range excludes {
			if v.address == vv.address {
				include = true
				break
			}
		}
		if !include {
			weightItems = append(weightItems, bl.LoadItem{
				Index: index,
			})
		}
	}
	if len(weightItems) == 0 {
		return nil
	}

	return executeNodes[randomLoadBalance.DoSelect(weightItems)]
}
