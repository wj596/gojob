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
package routes

import (
	"sync"
	"time"

	"gojob/models"
	"gojob/util/logs"
	"gojob/util/stringutil"

	"github.com/pkg/errors"
)

const certificateClearInterval = 1800

type Certificate struct {
	Token      string
	activeTime int64
	User       *models.User
}

var certificates sync.Map

func doLogin(name string, password string) (*Certificate, error) {
	user, err := models.GetUser(name)
	if nil != err {
		return nil, errors.Errorf("用户名或密码不正确")
	}
	if password != user.Password {
		return nil, errors.Errorf("用户名或密码不正确")
	}
	cf := new(Certificate)
	cf.Token = stringutil.UUID()
	cf.activeTime = time.Now().Unix()
	cf.User = user
	certificates.Store(cf.Token, cf)

	return cf, nil
}

func doLogout(token string) {
	certificates.Delete(token)
}

func doAuthorised(token string) (*Certificate, error) {
	v, exist := certificates.Load(token)
	if !exist {
		return nil, errors.Errorf("无效的token")
	}
	cf := v.(*Certificate)
	cf.activeTime = time.Now().Unix()
	return cf, nil
}

// 清理过期Token
func StartCertificateClearTask() {
	ticker := time.NewTicker(certificateClearInterval * time.Second)
	go func(ticker *time.Ticker) {
		for {
			<-ticker.C
			overdueList := make([]interface{}, 0)
			certificates.Range(func(key, value interface{}) bool {
				cf := value.(*Certificate)
				if time.Now().Unix()-cf.activeTime >= 1800 {
					overdueList = append(overdueList, key)
				}
				return true
			})
			logs.Infof("清理过期的用户登录凭证,%v条", len(overdueList))
			for _, v := range overdueList {
				certificates.Delete(v)
			}
		}
	}(ticker)
}
