package integration

import "go.uber.org/zap/zapcore"

// AllLogLevelEnabler is a simple log level enabled that allows to grap logs from all kind of levels
type AllLogLevelEnabler struct{}

func (a AllLogLevelEnabler) Enabled(zapcore.Level) bool {
	return true
}
