package logs

import (
	"github.com/go-logr/zerologr"
	"os"

	"github.com/go-logr/logr"
)

type LogRLogger struct {
	zlogger zerologr.Logger
}

func (l LogRLogger) toLogr() logr.Logger {
	return l.zlogger
}

func (l LogRLogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.toLogr().V(1).Info(msg, keysAndValues...)
}

func (l LogRLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.toLogr().Info(msg, keysAndValues...)
}

func (l LogRLogger) Warnw(msg string, err error, keysAndValues ...interface{}) {
	if err != nil {
		keysAndValues = append(keysAndValues, "error", err)
	}
	l.toLogr().Info(msg, keysAndValues...)
}

func (l LogRLogger) Errorw(msg string, err error, keysAndValues ...interface{}) {
	l.toLogr().Error(err, msg, keysAndValues...)
}
func (l LogRLogger) Fatalw(msg string, err error, keysAndValues ...interface{}) {
	l.toLogr().Error(err, msg, keysAndValues...)
	os.Exit(1)
}

func (l LogRLogger) WithValues(keysAndValues ...interface{}) Logger {
	return LogRLogger{zlogger: l.toLogr().WithValues(keysAndValues...)}
}

func (l LogRLogger) WithName(name string) Logger {
	return LogRLogger{zlogger: l.toLogr().WithName(name)}
}

func (l LogRLogger) WithCallDepth(depth int) Logger {
	return LogRLogger{zlogger: l.toLogr().WithCallDepth(depth)}
}
