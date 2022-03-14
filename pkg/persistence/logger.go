package persistence

import (
	"fmt"
	"github.com/autom8ter/morpheus/pkg/logger"
)

type bLogger struct{}

func (l bLogger) Errorf(s string, i ...interface{}) {
	logger.L.Error(fmt.Sprintf(s, i...), map[string]interface{}{})
}

func (l bLogger) Warningf(s string, i ...interface{}) {
	logger.L.Warn(fmt.Sprintf(s, i...), map[string]interface{}{})
}

func (l bLogger) Infof(s string, i ...interface{}) {
	logger.L.Info(fmt.Sprintf(s, i...), map[string]interface{}{})
}

func (l bLogger) Debugf(s string, i ...interface{}) {
	logger.L.Debug(fmt.Sprintf(s, i...), map[string]interface{}{})
}
