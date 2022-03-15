package logger

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
)

type bLogger struct{}

func (l bLogger) Errorf(s string, i ...interface{}) {
	L.Error(fmt.Sprintf(s, i...), map[string]interface{}{})
}

func (l bLogger) Warningf(s string, i ...interface{}) {
	L.Warn(fmt.Sprintf(s, i...), map[string]interface{}{})
}

func (l bLogger) Infof(s string, i ...interface{}) {
	L.Info(fmt.Sprintf(s, i...), map[string]interface{}{})
}

func (l bLogger) Debugf(s string, i ...interface{}) {
	L.Debug(fmt.Sprintf(s, i...), map[string]interface{}{})
}

func BadgerLogger() badger.Logger {
	return bLogger{}
}
