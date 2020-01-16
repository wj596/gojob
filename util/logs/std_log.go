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
package logs

import (
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLumberjackLogger(config *LoggerConfig) *lumberjack.Logger {
	if config.MaxSize <= 0 {
		config.MaxSize = logMaxSize
	}
	if config.MaxSize <= 0 {
		config.MaxAge = logMaxAge
	}
	loghook := lumberjack.Logger{ //定义日志分割器
		Filename:  config.LogFile,  // 日志文件路径
		MaxSize:   config.MaxSize,  // 文件最大M字节
		MaxAge:    config.MaxAge,   // 最多保留几天
		Compress:  config.Compress, // 是否压缩
		LocalTime: true,
	}
	return &loghook
}
