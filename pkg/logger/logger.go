package logger

import (
	"bytes"
	"github.com/autom8ter/morpheus/pkg/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"time"
)

func init() {
	L = New()
}

var L *Logger

type Logger struct {
	logger *zap.Logger
}

func New() *Logger {
	hst, _ := os.Hostname()
	fields := map[string]interface{}{
		"host":    hst,
		"service": "morpheus",
		"version": version.Version,
	}

	zap.NewDevelopmentConfig()
	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "ts",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "function",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	})
	core := zapcore.NewCore(jsonEncoder, os.Stdout, zap.InfoLevel)
	return &Logger{
		logger: zap.New(core).With(toFields(fields)...),
	}
}

func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.logger.Info(msg, toFields(fields)...)
}

func (l *Logger) Fatal(msg string, fields map[string]interface{}) {
	l.logger.Fatal(msg, toFields(fields)...)
}

func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.logger.Warn(msg, toFields(fields)...)
}

func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.logger.Debug(msg, toFields(fields)...)
}

func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.logger.Error(msg, toFields(fields)...)
}

func (l *Logger) HTTPError(w http.ResponseWriter, message string, err error, status int) {
	fields := map[string]interface{}{
		"http_response_status": status,
	}
	if err != nil {
		fields["err"] = err
	}
	switch status {
	case 401, 403, 404:
		l.Info(message, fields)
	default:
		l.Error(message, fields)
	}
	http.Error(w, message, status)
}

func (l *Logger) Zap() *zap.Logger {
	return l.logger
}

func toFields(fields map[string]interface{}) []zap.Field {
	var zfields []zap.Field
	for k, v := range fields {
		zfields = append(zfields, zap.Any(k, v))
	}
	return zfields
}

func Middleware(lgger *Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		ww := &responseWriterWrapper{w: w}
		defer func() {
			lgger.Info("http request/response", map[string]interface{}{
				"url":             r.URL.String(),
				"method":          r.Method,
				"elapsed_ns":      time.Since(now).Nanoseconds(),
				"response_status": ww.StatusCode(),
				"response_body":   ww.Body(),
			})
		}()
		handler.ServeHTTP(ww, r)
	})
}

type responseWriterWrapper struct {
	w          http.ResponseWriter
	body       bytes.Buffer
	statusCode int
}

func (i *responseWriterWrapper) Header() http.Header {
	return i.w.Header()
}

func (i *responseWriterWrapper) Write(buf []byte) (int, error) {
	i.body.Write(buf)
	return i.w.Write(buf)
}

func (i *responseWriterWrapper) WriteHeader(statusCode int) {
	i.statusCode = statusCode
	i.w.WriteHeader(statusCode)
}

func (i *responseWriterWrapper) Body() string {
	return i.body.String()
}

func (i *responseWriterWrapper) StatusCode() int {
	return i.statusCode
}
