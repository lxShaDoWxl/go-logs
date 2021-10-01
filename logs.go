package logs

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"golang.org/x/xerrors"
)

type Config struct {
	ConfigSentry ConfigSentry
}

func Initialize(c Config) {
	initializeSentry(c.ConfigSentry)
}

// Info logs a message at info level.
func Info(ctx context.Context, f interface{}, v ...interface{}) {
	message := formatLog(f, v...)
	logs.Info(message)
}

// Error logs a message at error level.
func Error(ctx context.Context, err error) {
	logs.Error(err.Error())
	sendLogSentry(ctx, err)
}

// Fatalf logs a message at critical level.
func Fatalf(ctx context.Context, f interface{}, v ...interface{}) {
	message := formatLog(f, v...)
	logs.Critical(message)
	sendLogSentry(ctx, xerrors.New(message))
	os.Exit(1)
}
func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	//nolint:gocritic,gosimple
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if !strings.Contains(msg, "%") {
			// do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}
