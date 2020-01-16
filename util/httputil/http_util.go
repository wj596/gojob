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
package httputil

import (
	"context"
	"net/http"
	"time"

	"gojob/util/logs"
	"gojob/util/stringutil"

	"github.com/pkg/errors"
)

// 重试条件
type RetryConditionFunc func(*http.Response) bool

type HttpClient struct {
	client          *http.Client
	RetryCount      int
	RetryWaitTime   time.Duration
	RetryConditions []RetryConditionFunc
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{},
	}
}

// 设置超时时间，单位为秒
func (this *HttpClient) SetTimeout(timeout int) *HttpClient {
	if timeout > 0 {
		this.client.Timeout = time.Duration(timeout) * time.Second
	}
	return this
}

// 设置重试次数
func (this *HttpClient) SetRetryCount(retryCount int) *HttpClient {
	if retryCount > 0 {
		this.RetryCount = retryCount
	}
	return this
}

// 设置重试间隔时间，单位为秒
func (this *HttpClient) SetRetryWaitTime(retryWaitTime int) *HttpClient {
	if retryWaitTime > 0 {
		this.RetryWaitTime = time.Duration(retryWaitTime) * time.Second
	}
	return this
}

func (this *HttpClient) AddRetryCondition(retryCondition RetryConditionFunc) *HttpClient {
	if retryCondition != nil {
		this.RetryConditions = append(this.RetryConditions, retryCondition)
	}
	return this
}

func (this *HttpClient) SetTransport(transport http.RoundTripper) *HttpClient {
	if transport != nil {
		this.client.Transport = transport
	}
	return this
}

func (this *HttpClient) retryNecessary(res *http.Response) bool {
	for _, condition := range this.RetryConditions {
		if condition(res) {
			return true
		}
	}
	return false
}

func (this *HttpClient) Execute(request *http.Request) (*http.Response, error) {
	ctx := request.Context()
	startTime := time.Now().UnixNano()
	res, err := this.client.Do(request)
	diff := (time.Now().UnixNano() - startTime) / int64(time.Millisecond)
	logs.Infof("%s %s 耗时%d毫秒", request.Method, request.URL.String(), diff)
	if nil != err {
		logs.Infof("%s %s 请求错误:%v", request.Method, request.URL.String(), err.Error())
	}
	if this.RetryCount < 1 {
		return res, err
	} else {
		if this.retryNecessary(res) {
			for i := 0; i < this.RetryCount; i++ {
				logs.Infof("%s %s 第%d次重试", request.Method, request.URL.String(), i+1)
				res, err = this.client.Do(request)
				if !this.retryNecessary(res) || (i+1) == this.RetryCount {
					return res, err
				}
				select {
				case <-time.After(this.RetryWaitTime):
				case <-ctx.Done():
					return res, err
				}
			}
		}
	}
	return res, err
}

func (this *HttpClient) NewRequest() *HttpRequest {
	return &HttpRequest{
		httpClient: this,
		headers:    make(map[string]string),
		parameters: make(map[string]string),
	}
}

type HttpRequest struct {
	ctx        context.Context
	httpClient *HttpClient
	method     string
	headers    map[string]string
	parameters map[string]string
}

func (this *HttpRequest) SetContext(context context.Context) *HttpRequest {
	this.ctx = context
	return this
}

func (this *HttpRequest) AddHeader(key string, val string) *HttpRequest {
	this.headers[key] = val
	return this
}

func (this *HttpRequest) AddParameter(key string, val string) *HttpRequest {
	this.parameters[key] = val
	return this
}

func (this *HttpRequest) Get(url string) (*http.Response, error) {
	requestUrl := stringutil.BuildQueryString(url, this.parameters)
	request, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if nil != err {
		return nil, errors.Errorf("NewRequest 错误 - %s", err.Error())
	}
	for k, v := range this.headers {
		request.Header.Add(k, v)
	}
	return this.httpClient.Execute(request)
}
