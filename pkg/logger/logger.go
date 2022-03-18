package logger

import (
	"github.com/armon/go-metrics"
	"github.com/autom8ter/morpheus/pkg/version"
	"github.com/palantir/stacktrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"time"
)

func init() {

	L = New(true)
}

var L *Logger

type Logger struct {
	logger  *zap.Logger
	metrics *metrics.Metrics
}

func New(json bool) *Logger {
	hst, _ := os.Hostname()
	fields := map[string]interface{}{
		"host":    hst,
		"service": "morpheus",
		"version": version.Version,
	}
	inm := metrics.NewInmemSink(10*time.Second, time.Minute)
	m, err := metrics.NewGlobal(metrics.DefaultConfig("morpheus"), inm)
	if err != nil {
		panic(stacktrace.Propagate(err, ""))
	}

	zap.NewDevelopmentConfig()
	if json {
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
			metrics: m,
			logger:  zap.New(core).With(toFields(fields)...),
		}
	} else {
		txtEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
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
		core := zapcore.NewCore(txtEncoder, os.Stdout, zap.InfoLevel)
		return &Logger{
			metrics: m,
			logger:  zap.New(core).With(toFields(fields)...),
		}
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

func (l *Logger) Error(msg string, err error, fields map[string]interface{}) {
	if err != nil {
		fields["error.stack"] = err
		fields["error.cause"] = stacktrace.RootCause(err)
	}
	l.logger.Error(msg, toFields(fields)...)
}

func (l *Logger) Metrics() *metrics.Metrics {
	return l.metrics
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
		l.Error(message, err, fields)
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
