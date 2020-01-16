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
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"gojob/conf"
	"gojob/internal"
	"gojob/models"
	_ "gojob/statik"
	"gojob/util/dateutil"
	"gojob/util/logs"

	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"
	"go.uber.org/atomic"
)

const (
	httpLogFileName = "http.log"
	version         = "v1"
)

var router *gin.Engine
var staticFS http.FileSystem
var httpProxy *httputil.ReverseProxy
var proxyAddress atomic.String
var proxyLock sync.Mutex

func StartRouter(bind string, port int) {
	gin.SetMode(gin.ReleaseMode)
	router = gin.New()
	staticFS, _ = fs.New()
	initTemplate()
	router.Use(corsMiddleware())
	router.Use(proxyMiddleware())
	router.Use(loggerMiddleware())
	router.Use(staticResMiddleware())
	initActions()

	listen := fmt.Sprintf(":%s", strconv.Itoa(port))
	if "" != bind {
		listen = fmt.Sprintf("%s:%s", bind, strconv.Itoa(port))
	}
	server := &http.Server{
		Addr:           listen,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			logs.Errorf("Http Server 启动失败：%s \n", err.Error())
			signalCh <- syscall.SIGINT
		}
	}()
	log.Printf("启动 Http Server，进程 %d  监听端口 %s \n", syscall.Getpid(), listen)
	sig := <-signalCh
	log.Printf("Http Server 服务停止：%s \n", sig.String())
	models.CloseBoltDB()
	err := server.Shutdown(nil)
	if err != nil {
		logs.Error(err.Error())
	}
}

func initTemplate() {
	tpl := template.New("")
	indexTpl := tpl.New("index.html")
	file, err := staticFS.Open("/index.html")
	var content string
	if err == nil {
		if bytes, err := ioutil.ReadAll(file); err == nil {
			content = string(bytes)
		}
		defer file.Close()
	}
	indexTpl.Parse(content)

	router.SetHTMLTemplate(tpl)
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	router.GET("index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("origin")
		if "" == origin {
			origin = c.Request.Header.Get("Referer")
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "18000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT,DELETE,OPTIONS,PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, XMLHttpRequest, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}

func loggerMiddleware() gin.HandlerFunc {
	sysLogConf := conf.GetConfig().LoggerConfig
	logConf := logs.LoggerConfig{
		Level:    sysLogConf.Level,
		LogPath:  sysLogConf.LogPath,
		LogFile:  filepath.Join(sysLogConf.LogPath, httpLogFileName),
		MaxSize:  sysLogConf.MaxSize,
		MaxAge:   sysLogConf.MaxAge,
		Compress: sysLogConf.Compress,
		Encoding: sysLogConf.Encoding,
	}
	httpLogger := logs.NewFileZapLogger(&logConf)

	return func(c *gin.Context) {
		start := dateutil.NowMillisecond()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		c.Next()
		latency := dateutil.NowMillisecond() - start
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()
		if raw != "" {
			path = path + "?" + raw
		}
		httpLogger.Sugar().Infof("GIN: %3d | %5v | %15s |%-5s %s %s",
			statusCode, latency, clientIP, method, path, comment)
	}
}

func checkProxyCondition() bool {
	leaderId := internal.GetLeaderId()
	if "" == leaderId {
		return false
	}

	leader, err := internal.GetRuntimeClusterNode(leaderId)
	if err != nil {
		return false
	}

	if proxyAddress.Load() != leader.HttpAddr {
		proxyLock.Lock()
		defer proxyLock.Unlock()

		proxyAddress.Store(leader.HttpAddr)
		target, _ := url.Parse("http://" + proxyAddress.Load())
		httpProxy = httputil.NewSingleHostReverseProxy(target)
	}

	return true
}

func proxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if internal.IsClusterMode() {
			feasible := checkProxyCondition()
			if !feasible {
				c.JSON(http.StatusInternalServerError, gin.H{
					"succeed": false,
					"msg":     "找不到集群主节点，请确定集群是否正常启动",
				})
				c.Abort()
				return
			}
			if internal.IsLeader() {
				c.Next()
			} else {
				logs.Infof("Proxy ServeHTTP :%s", proxyAddress.Load())
				httpProxy.ServeHTTP(c.Writer, c.Request)
				c.Abort()
			}
		} else {
			c.Next()
		}
	}
}

func staticResMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		var contentType string
		switch {
		case strings.HasPrefix(path, "/js/"):
			contentType = "text/javascript; charset=utf-8"
		case strings.HasPrefix(path, "/css/"):
			contentType = "text/css; charset=utf-8"
		case strings.HasPrefix(path, "/favicon.ico"):
			contentType = "image/vnd.microsoft.icon"
		case strings.HasPrefix(path, "/fonts/"):
			contentType = "font/woff"
		}

		if "" != contentType {
			file, err := staticFS.Open(path)
			if err == nil {
				if bytes, err := ioutil.ReadAll(file); err == nil {
					c.Data(http.StatusOK, contentType, bytes)
					c.Abort()
				}
				file.Close()
			}
		}

		c.Next()
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		token := c.Request.Header.Get("Authorization")

		if "/ui/users/login" == path || "/ui/index" == path {
			c.Next()
			return
		}

		if "" == token {
			c.AbortWithStatus(401)
			return
		}

		_, err := doAuthorised(token)
		if nil != err {
			c.AbortWithStatus(401)
			return
		}

		c.Next()
	}
}

func signMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sign := c.Request.Header.Get("X-Sign")
		if "" == sign {
			c.String(http.StatusBadRequest, "Header参数X-Sign,不能为空")
			c.Abort()
			return
		}

		timestamp := c.Request.Header.Get("X-Timestamp")
		if "" == timestamp {
			c.String(http.StatusBadRequest, "Header参数X-Timestamp,不能为空")
			c.Abort()
			return
		}

		stamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil || len(timestamp) != 10 {
			c.String(http.StatusBadRequest, "Header参数timestamp不正确，必须为10位时间戳")
			c.Abort()
			return
		}

		if stamp < (time.Now().Unix() - 1800) {
			c.String(http.StatusBadRequest, "签名过期")
			c.Abort()
			return
		}

		plaintext := c.Request.RequestURI + timestamp
		backstage := internal.Signature(plaintext)

		if backstage != sign {
			c.String(http.StatusUnauthorized, "签名无效")
			c.Abort()
			return
		}

		c.Next()
	}
}
