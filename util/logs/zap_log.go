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
	"log"
	"path/filepath"
	"time"

	"gojob/util/fileutil"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logFileName        = "system.log"
	logLevelDebug      = "debug"
	logLevelInfo       = "info"
	logLevelWarn       = "warn"
	logLevelError      = "error"
	logMaxSize         = 500
	logMaxAge          = 30
	logEncodingConsole = "console"
	logEncodingJson    = "json"
)

var zapLogger *zap.Logger

// logger 配置
type LoggerConfig struct {
	Level    string `yaml:"level"` //日志级别 debug|info|warn|error
	LogPath  string //日志目录
	MaxSize  int    `yaml:"max_size"` //日志文件最大M字节
	MaxAge   int    `yaml:"max_age"`  //日志文件最大存活的天数
	Compress bool   `yaml:"compress"` //是否启用压缩
	Encoding string `yaml:"encoding"` //日志编码 console|json
	LogFile  string //日志文件
}

func InitLogger(config *LoggerConfig, options ...zap.Option) {
	if config.LogPath == "" { // 标准输出
		zapLogger = NewZapLogger(config, options...)
	} else { // 文件输出
		if err := fileutil.MkdirIfNecessary(config.LogPath); err != nil {
			log.Panicf("日志目录:%s，创建失败 \n", config.LogPath)
		}
		config.LogFile = filepath.Join(config.LogPath, logFileName)
		zapLogger = NewFileZapLogger(config, options...)
	}
}

func GetLogger() *zap.Logger {
	return zapLogger
}

func Debug(msg string, fields ...zapcore.Field) {
	zapLogger.Debug(msg, fields...)
}

func Debugf(template string, args ...interface{}) {
	zapLogger.Sugar().Debugf(template, args...)
}

func Info(msg string, fields ...zapcore.Field) {
	zapLogger.Info(msg, fields...)
}

func Infof(template string, args ...interface{}) {
	zapLogger.Sugar().Infof(template, args...)
}

func Warn(msg string, fields ...zapcore.Field) {
	zapLogger.Warn(msg, fields...)
}

func Warnf(template string, args ...interface{}) {
	zapLogger.Sugar().Warnf(template, args...)
}

func Error(msg string, fields ...zapcore.Field) {
	zapLogger.Error(msg, fields...)
}

func Errorf(template string, args ...interface{}) {
	zapLogger.Sugar().Errorf(template, args...)
}

func NewZapLogger(config *LoggerConfig, options ...zap.Option) *zap.Logger {
	level := toZapLevel(config.Level)
	encoderConfig := newEncoderConfig()
	encoding := logEncodingConsole
	if config.Encoding == logEncodingJson {
		encoding = logEncodingJson
	}
	conf := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		EncoderConfig:    encoderConfig,
		Encoding:         encoding,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	zapLogger, err := conf.Build(options...)
	if err != nil {
		log.Printf("Zap日志创建失败，使用NewExample创建\n")
		zapLogger = zap.NewExample(options...)
	}
	return zapLogger
}

func NewFileZapLogger(config *LoggerConfig, options ...zap.Option) *zap.Logger {
	level := toZapLevel(config.Level)
	encoderConfig := newEncoderConfig()
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
	var encoder zapcore.Encoder
	if config.Encoding == logEncodingJson {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(&loghook),
		level,
	)
	return zap.New(core)
}

func toZapLevel(level string) zapcore.Level {
	var zapLevel zapcore.Level
	switch level {
	case logLevelInfo:
		zapLevel = zap.InfoLevel
	case logLevelWarn:
		zapLevel = zap.WarnLevel
	case logLevelError:
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.DebugLevel
	}
	return zapLevel
}

func newEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	encoderConfig.CallerKey = ""
	return encoderConfig
}
