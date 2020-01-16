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
package stringutil

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/satori/go.uuid"
)

// 产生UUID
func UUID() string {
	return strings.ReplaceAll(uuid.NewV4().String(), "-", "")
}

// MD2编码
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func ToIntSafe(str string) int {
	v, e := strconv.Atoi(str)
	if nil != e {
		return 0
	}
	return v
}

func ToUintSafe(str string) uint64 {
	v, e := strconv.ParseUint(str, 10, 64)
	if nil != e {
		return 0
	}
	return v
}

func UintToStr(u uint64) string {
	return strconv.FormatUint(u, 10)
}

// 键值对转MAP,类似"name=wangjie,age=20"或者"name=wangjie|age=20"
func KVsToMap(base string, sep string) map[string]string {
	ret := make(map[string]string)
	if "" != base && "" != sep {
		kvs := strings.Split(base, sep)
		for _, kv := range kvs {
			temp := strings.Split(kv, "=")
			if len(temp) < 2 {
				continue
			}
			if temp[0] == "" {
				continue
			}
			ret[temp[0]] = temp[1]
		}
	}
	return ret
}

// 构造查询字符串
func BuildQueryString(base string, parameters map[string]string) string {
	exist := false
	if strings.Contains(base, "?") {
		exist = true
	}
	var buffer bytes.Buffer
	buffer.WriteString(base)
	for k, v := range parameters {
		var temp string
		if !exist {
			temp = "?" + k + "=" + v
			exist = true
		} else {
			temp = "&" + k + "=" + v
		}
		buffer.WriteString(temp)
	}
	return buffer.String()
}

func IsEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func IsChineseChar(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || (regexp.MustCompile("[\u3002\uff1b\uff0c\uff1a\u201c\u201d\uff08\uff09\u3001\uff1f\u300a\u300b]").MatchString(string(r))) {
			return true
		}
	}
	return false
}
