package log

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	XRequestID = "X-Request-Id"
	XB3TraceID = "X-B3-Traceid"
	MpDebug    = "Mp-Debug"

	StackLogField     = "stack"
	CallerLogField    = "caller"
	RequestIDLogField = "request_id"
	TraceLogField     = "trace_id"

	StackLogMaxDepth = 20
)

type Fields = logrus.Fields

type Entry struct {
	*logrus.Entry
}

// Error 记录错误日志
// 日志会自动注入stack字段
func (entry *Entry) Error(args ...interface{}) {
	entry.WithFields(logrus.Fields{
		StackLogField:  string(Stack(2, StackLogMaxDepth)),
		CallerLogField: CallerName(),
	}).Error(args...)
}

// Deprecated: msg中不应该包含变量
func (entry *Entry) Errorf(format string, args ...interface{}) {
	entry.WithFields(logrus.Fields{
		StackLogField:  string(Stack(2, StackLogMaxDepth)),
		CallerLogField: CallerName(),
	}).Errorf(format, args...)
}

func (entry *Entry) Errorln(args ...interface{}) {
	entry.WithFields(logrus.Fields{
		StackLogField:  string(Stack(2, StackLogMaxDepth)),
		CallerLogField: CallerName(),
	}).Errorln(args...)
}

// Deprecated: use WithCtxFields instead
func Debug(args ...interface{}) {
	Logger.Debugln(args...)
}

// Deprecated: use WithCtxFields instead
func Warn(args ...interface{}) {
	Logger.Warnln(args...)
}

// Deprecated: use WithCtxFields instead
func Err(args ...interface{}) {
	Logger.WithFields(logrus.Fields{
		StackLogField:  string(Stack(2, StackLogMaxDepth)),
		CallerLogField: CallerName(),
	}).Errorln(args...)
}

// Deprecated: use WithCtxFields instead
func Info(args ...interface{}) {
	Logger.Infoln(args...)
}

// Deprecated: use WithCtxFields instead
func Print(args ...interface{}) {
	Logger.Print(args...)
}

// Deprecated: use WithCtxFields instead
func Println(args ...interface{}) {
	Logger.Println(args...)
}

// Deprecated: use WithCtxFields instead
func Fatal(args ...interface{}) {
	Logger.Fatalln(args...)
}

// Deprecated: use WithCtxFields instead
func DebugCtx(ctx context.Context, args ...interface{}) {
	cloneLogger(ctx).Debugln(args...)
}

// Deprecated: use WithCtxFields instead
func WarnCtx(ctx context.Context, args ...interface{}) {
	cloneLogger(ctx).Warnln(args...)
}

// Deprecated: use WithCtxFields instead
func ErrCtx(ctx context.Context, args ...interface{}) {
	cloneLogger(ctx).Error(args...)
}

// Deprecated: use WithCtxFields instead
func InfoCtx(ctx context.Context, args ...interface{}) {
	cloneLogger(ctx).Infoln(args...)
}

// Deprecated: use WithCtxFields instead
func PrintCtx(ctx context.Context, args ...interface{}) {
	cloneLogger(ctx).Print(args...)
}

// Deprecated: use WithCtxFields instead
func PrintlnCtx(ctx context.Context, args ...interface{}) {
	cloneLogger(ctx).Println(args...)
}

// Deprecated: use WithCtxFields instead
func FatalCtx(ctx context.Context, args ...interface{}) {
	cloneLogger(ctx).Fatalln(args...)
}

// Deprecated: use WithCtxFields instead
func WithFields(fields Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

// WithCtxFields 记录日志自动记录request_id
// 示例：log.WithCtxFields(ctx, log.Fields{"key":"value"}).Error("this is error msg")
func WithCtxFields(ctx context.Context, fields Fields) *Entry {
	if fields == nil {
		fields = Fields{}
	}
	return &Entry{Entry: cloneLogger(ctx).WithFields(fields)}
}

// WithCtx 记录日志自动记录request_id
// 示例：log.WithCtxFields(ctx).Error("this is error msg")
func WithCtx(ctx context.Context) *Entry {
	return cloneLogger(ctx)
}

func cloneLogger(ctx context.Context) *Entry {
	l := Logger
	if ForceDebugMode(ctx) {
		l = &logrus.Logger{
			Out:          Logger.Out,
			Hooks:        Logger.Hooks,
			Formatter:    Logger.Formatter,
			ReportCaller: Logger.ReportCaller,
			Level:        logrus.DebugLevel,
			ExitFunc:     Logger.ExitFunc,
		}
	}

	e := l.WithField(RequestIDLogField, ParseRequestID(ctx)).
		WithField(TraceLogField, ParseTraceID(ctx))
	return &Entry{Entry: e}
}

func ForceDebugMode(ctx context.Context) bool {
	if forceDebug, ok := ctx.Value(MpDebug).(bool); ok && forceDebug {
		return true
	}

	return false
}

func ParseRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(XRequestID).(string); ok && reqID != "" {
		return reqID
	}
	return ""
}

func ParseTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(XB3TraceID).(string); ok && traceID != "" {
		return traceID
	}
	return ""
}

func InjectHeader(ctx context.Context, header http.Header) http.Header {
	header.Set(XRequestID, ParseRequestID(ctx))
	header.Set(XB3TraceID, ParseTraceID(ctx))
	if ForceDebugMode(ctx) {
		header.Set(MpDebug, "1")
	}

	return header
}
