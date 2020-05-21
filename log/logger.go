package log

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

const (
	// DEBUG is the log level of debug. Note that the level of INFO is 0.
	DEBUG = 1
)

// ZLog is a global zap logger.
var ZLog = ZapLogger(true)

// Log is a global base logger.
// It uses a dev ZapLogger by default.
var Log = zapr.NewLogger(ZLog)

// ZapLogger is a Logger implementation.
// If development is true, a Zap development config will be used
// (stacktraces on warnings, no sampling), otherwise a Zap production
// config will be used (stacktraces on errors, sampling).
func ZapLogger(development bool) *zap.Logger {
	var zapLog *zap.Logger
	var err error
	if development {
		zapLogCfg := zap.NewDevelopmentConfig()
		zapLog, err = zapLogCfg.Build(zap.AddCallerSkip(1))
	} else {
		zapLogCfg := zap.NewProductionConfig()
		zapLog, err = zapLogCfg.Build(zap.AddCallerSkip(1))
	}

	if err != nil {
		panic(err)
	}
	return zapLog
}

// SetLogger sets a concrete logging implementation for the base logger.
func SetLogger(l logr.Logger) {
	Log = l
}

// WithName wraps WithName of base logger.
func WithName(n string) logr.Logger {
	return Log.WithName(n)
}

// WithValues wraps WithValues of base logger.
func WithValues(kv ...interface{}) logr.Logger {
	return Log.WithValues(kv...)
}
