package zipkin

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openzipkin/zipkin-go"
	zipkinModel "github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"github.com/openzipkin/zipkin-go/reporter"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
	kafkareporter "github.com/openzipkin/zipkin-go/reporter/kafka"
	logreporter "github.com/openzipkin/zipkin-go/reporter/log"
	"meipian.cn/meigo/v2/config"
)

func init() {
	InitZipkin(&ZipkinConf{})
}

type ZipkinConf struct {
	ServiceName         string // 服务名
	ReporterType        string // repoter type
	HttpReporterUrl     string
	LogReporterFilePath string
	KafkaReporterUrl    []string
	OpenTracer          bool
	Mod                 uint64
}

var (
	ZipkinReporter reporter.Reporter
	ZipkinTracer   *zipkin.Tracer
	HttpClient     *Client
)

const (
	REPORTER_HTTP  = "http"
	REPORTER_LOG   = "log"
	REPORTER_KAFKA = "kafka"
)

var zeroDialer = net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
}

const (
	DefaultRequestTimeout      = 120 * time.Second
	DefaultMaxIdleConns        = 1000
	DefaultMaxIdleConnsPerHost = 50
	DefaultMaxConnsPerHost     = 500
	DefaultIdleConnTimeout     = 10 * time.Minute
)

func DefaultServiceHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        DefaultMaxIdleConns,
			MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
			MaxConnsPerHost:     DefaultMaxConnsPerHost,
			IdleConnTimeout:     DefaultIdleConnTimeout,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// 强制使用ipv4解析
				// 仅仅在go1.17之后有效
				return zeroDialer.DialContext(ctx, "tcp4", addr)
			},
		},
		Timeout: DefaultRequestTimeout,
	}
}

func InitZipkinWithApolloConfig() {
	zipkinConf := &ZipkinConf{}
	zipkinConf.ServiceName = config.GetStr("zipkin.ServiceName")
	zipkinConf.OpenTracer = config.GetBool("zipkin.OpenTracer", false)
	zipkinConf.ReporterType = config.GetStr("zipkin.ReporterType")
	zipkinConf.Mod = uint64(config.GetIntDft("zipkin.Mod", 1))
	zipkinConf.HttpReporterUrl = config.GetStr("zipkin.HttpReporterUrl")
	zipkinConf.LogReporterFilePath = config.GetStr("zipkin.LogReporterFilePath")
	KafkaReporterUrl := config.GetStr("zipkin.KafkaReporterUrl")
	if KafkaReporterUrl != "" {
		zipkinConf.KafkaReporterUrl = strings.Split(KafkaReporterUrl, ",")
	} else {
		zipkinConf.KafkaReporterUrl = []string{}
	}

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100
	InitZipkin(zipkinConf)
}

func InitZipkin(conf *ZipkinConf) {
	// reporter
	var err error
	if !conf.OpenTracer {
		ZipkinReporter = reporter.NewNoopReporter()
	} else {
		switch conf.ReporterType {
		case REPORTER_HTTP:
			ZipkinReporter = httpreporter.NewReporter(conf.HttpReporterUrl, httpreporter.BatchSize(10))
		case REPORTER_LOG:
			rotaterWriter := NewRotaterWriter(conf.LogReporterFilePath)
			zipkinLogger := log.New(rotaterWriter, "", 0)
			ZipkinReporter = logreporter.NewReporter(zipkinLogger)
		case REPORTER_KAFKA:
			ZipkinReporter, err = kafkareporter.NewReporter(conf.KafkaReporterUrl)
			if err != nil {
				log.Println("ZipkinReporter init faild", err)
				ZipkinReporter = reporter.NewNoopReporter()
			}
		default:
			ZipkinReporter = reporter.NewNoopReporter()
		}
	}

	// sampler
	sampler := zipkin.NewModuloSampler(conf.Mod)

	// tracer
	endpoint, _ := zipkin.NewEndpoint(conf.ServiceName, "localhost")
	shadowTracer, err := zipkin.NewTracer(ZipkinReporter, zipkin.WithLocalEndpoint(endpoint), zipkin.WithSharedSpans(false), zipkin.WithSampler(sampler))
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}

	if ZipkinTracer == nil {
		ZipkinTracer = shadowTracer
	} else {
		*ZipkinTracer = *shadowTracer
	}

	c := DefaultServiceHttpClient()
	HttpClient, _ = NewClient(ZipkinTracer, WithClient(c), ClientTrace(true))

}

func GinZipkinMiddleware(ctx *gin.Context) {

	// try to extract B3 Headers from upstream
	sc := ZipkinTracer.Extract(b3.ExtractHTTP(ctx.Request))

	// create Span using SpanContext if found
	sp := ZipkinTracer.StartSpan(
		fmt.Sprintf("%s:%s", ctx.Request.Method, ctx.Request.URL.Path),
		zipkin.Kind(zipkinModel.Server),
		zipkin.Parent(sc),
	)
	defer sp.Finish()

	// add our span to context
	// tag typical HTTP zipkin items
	zipkin.TagHTTPMethod.Set(sp, ctx.Request.Method)
	zipkin.TagHTTPPath.Set(sp, ctx.Request.URL.Path)
	if ctx.Request.ContentLength > 0 {
		zipkin.TagHTTPRequestSize.Set(sp, strconv.FormatInt(ctx.Request.ContentLength, 10))
	}

	// tag found response size and status code on exit
	defer func() {
		code := ctx.Writer.Status()
		sCode := strconv.Itoa(code)
		zipkin.TagHTTPStatusCode.Set(sp, sCode)
	}()

	newCtx := zipkin.NewContext(ctx.Request.Context(), sp)
	ctx.Request = ctx.Request.WithContext(newCtx)

	sp.Annotate(time.Now(), "zipkin ready")

	ctx.Next()

	sp.Annotate(time.Now(), "action done")
}
