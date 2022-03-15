package raft

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/hashicorp/go-hclog"
	"io"
	"log"
)

type rlogger struct {
	logger hclog.Logger
}

func (r rlogger) Log(level hclog.Level, msg string, args ...interface{}) {
	logger.L.Info(fmt.Sprintf(msg, args...), map[string]interface{}{})
}

func (r rlogger) Trace(msg string, args ...interface{}) {
	logger.L.Info(fmt.Sprintf(msg, args...), map[string]interface{}{})
}

func (r rlogger) Debug(msg string, args ...interface{}) {
	logger.L.Debug(fmt.Sprintf(msg, args...), map[string]interface{}{})
}

func (r rlogger) Info(msg string, args ...interface{}) {
	logger.L.Info(fmt.Sprintf(msg, args...), map[string]interface{}{})
}

func (r rlogger) Warn(msg string, args ...interface{}) {
	logger.L.Warn(fmt.Sprintf(msg, args...), map[string]interface{}{})
}

func (r rlogger) Error(msg string, args ...interface{}) {
	logger.L.Error(fmt.Sprintf(msg, args...), map[string]interface{}{})
}

func (r rlogger) IsTrace() bool {
	return r.logger.IsTrace()
}

func (r rlogger) IsDebug() bool {
	return r.logger.IsDebug()
}

func (r rlogger) IsInfo() bool {
	return r.logger.IsInfo()
}

func (r rlogger) IsWarn() bool {
	return r.logger.IsWarn()
}

func (r rlogger) IsError() bool {
	return r.logger.IsError()
}

func (r rlogger) ImpliedArgs() []interface{} {
	return r.logger.ImpliedArgs()
}

func (r rlogger) With(args ...interface{}) hclog.Logger {
	return r.logger
}

func (r rlogger) Name() string {
	return r.logger.Name()
}

func (r rlogger) Named(name string) hclog.Logger {
	return r.logger
}

func (r rlogger) ResetNamed(name string) hclog.Logger {
	return r.logger
}

func (r rlogger) SetLevel(level hclog.Level) {
	return
}

func (r rlogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return r.logger.StandardLogger(opts)
}

func (r rlogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return r.logger.StandardWriter(opts)
}
