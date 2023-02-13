package util

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/openzipkin/zipkin-go/idgenerator"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"meipian.cn/meigo/v2/config"
	"meipian.cn/meigo/v2/log"
)

func InitRequestMetrics() {
	prometheus.MustRegister(httpRequestCount)
	prometheus.MustRegister(httpRequestDuration)
}

var (
	SuccessCode      int
	httpRequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "http",
		Subsystem: "request",
		Name:      "requests_count",
		Help:      "The total number of http request",
	}, []string{"method", "path", "status", "caller"})

	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "http",
		Subsystem: "request",
		Name:      "duration_seconds",
		Help:      "The http request latency in seconds",
	}, []string{"method", "path", "status", "caller"})
)

func init() {
	SuccessCode = GetErrOk()
}

func NewGin() *gin.Engine {
	InitRequestMetrics()

	if !config.GetBool("debug", false) {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Any("/listen", func(c *gin.Context) {
		c.String(http.StatusOK, "Success")
	})
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	globalMiddleware := []gin.HandlerFunc{
		RequestID(),
		Logger(),
		Metrics(),
		Recovery(),
	}

	engine.Use(globalMiddleware...)
	return engine
}

func SendOk(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": SuccessCode,
		"err":  false,
		"data": data,

		// "request_id": log.ParseRequestID(c.Request.Context()),
		"trace_id": log.ParseTraceID(c.Request.Context()),
	})
}

func SendError(c *gin.Context, code int, msg interface{}, data interface{}) {
	var msgStr string
	switch v := msg.(type) {
	case error:
		msgStr = v.Error()
	case string:
		msgStr = v
	}

	rsp := gin.H{
		"code": code,
		"err":  true,
		"msg":  msgStr,
		"data": data,

		// "request_id": log.ParseRequestID(c.Request.Context()),
		"trace_id": log.ParseTraceID(c.Request.Context()),
	}

	c.JSON(http.StatusOK, rsp)
}

func SendCommonError(c *gin.Context, msg interface{}) {
	SendError(c, ErrCommon, msg, nil)
}

func OutJson(c *gin.Context, code int, data interface{}) {
	msg := GetErrMsg(code)
	OutJsonMsg(c, code, msg, data)
}

// OutJsonWithMsg 服务响应，带msg输出
func OutJsonMsg(c *gin.Context, code int, msg string, data interface{}) {

	err := true // 默认有罪

	if code == 0 {
		code = GetErrOk()
		err = false
	}
	obj := gin.H{
		"code": code,
		"err":  err,
		"msg":  msg,
		"data": data,

		// "request_id": log.ParseRequestID(c.Request.Context()),
		"trace_id": log.ParseTraceID(c.Request.Context()),
	}

	c.Set("response", obj)

	bs, _ := json.Marshal(obj)
	str := string(bs)

	log.WithCtxFields(c.Request.Context(), log.Fields{
		"resp": str,
	}).Debug("--服务出参--" + c.Request.RequestURI)

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.String(200, str)
}

// OutJsonOk 输出成功数据
func OutJsonOk(c *gin.Context, data interface{}) {
	OutJson(c, 0, data)
}

// OutJsonErr 输出错误码，通过错误码获取msg
func OutJsonErr(c *gin.Context, code int) {
	OutJson(c, code, nil)
}

// OutJsonErrMsg 输出错误码及msg
func OutJsonErrMsg(c *gin.Context, code int, msg string) {
	OutJsonMsg(c, code, msg, nil)
}

// Logger 中间件 记录接口耗时, 请求
// 请求中包括 access, body, url, agent等
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 避免计算, 需要开启debug
		if config.GetBool("debug", false) || log.ForceDebugMode(c.Request.Context()) {
			startTime := time.Now()
			defer func() {
				bs, _ := httputil.DumpRequest(c.Request, true)
				log.WithCtxFields(c.Request.Context(), log.Fields{
					"req":      string(bs),
					"duration": time.Since(startTime).Seconds(),
				}).Debug("--服务入参--" + c.Request.RequestURI)
			}()
		}

		c.Next()
	}
}

// Err4Gin 通过panic和gin recovery中间件实现类似try..catch..功能。
type Err4Gin int

// Panic 类throw，需搭配gin recovery 中间件使用。
func Panic(code int) {
	panic(Err4Gin(code))
}

// Recovery 中间件，类catch，保证服务输出。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				switch err.(type) {
				case Err4Gin:
					code, _ := err.(Err4Gin)
					OutJson(c, int(code), nil)
				default: // 默认系统错误
					log.WithCtxFields(c.Request.Context(), log.Fields{
						"err": err,
					}).Error("recovery panic")
					// If the connection is dead, we can't write a status to it.
					if brokenPipe {
						c.Error(err.(error)) // nolint: errcheck
						c.Abort()
					} else {
						OutJson(c, ErrSystem, nil)
					}
				}

			}
		}()
		c.Next()
	}
}

func ParamInt(c *gin.Context, key string) int {
	r, _ := strconv.Atoi(c.Param(key))
	return r
}

// RequestID 传递request_id
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Request.Header.Get(log.XRequestID)
		if requestID == "" {
			requestID = uuid.Must(uuid.NewV4()).String()
		}
		newCtx := context.WithValue(c.Request.Context(), log.XRequestID, requestID)

		traceID := c.Request.Header.Get(log.XB3TraceID)
		if traceID == "" {
			traceID = idgenerator.NewRandom64().TraceID().String()
		}
		newCtx = context.WithValue(newCtx, log.XB3TraceID, traceID)

		if c.Request.Header.Get(log.MpDebug) != "" {
			// force debug
			newCtx = context.WithValue(newCtx, log.MpDebug, true)
		}

		c.Request = c.Request.WithContext(newCtx)
		c.Next()
	}
}

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		caller := c.Request.Header.Get("Mp-Caller")
		if caller == "" {
			caller = "unknown"
			// caller = c.ClientIP()
		}

		begin := time.Now()
		httpRequestCount.With(prometheus.Labels{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": strconv.Itoa(c.Writer.Status()),
			"caller": caller,
		}).Inc()

		c.Next()

		duration := float64(time.Since(begin)) / float64(time.Second)
		httpRequestDuration.With(prometheus.Labels{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": strconv.Itoa(c.Writer.Status()),
			"caller": caller,
		}).Observe(duration)
	}
}
