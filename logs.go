package logs

import (
	"cdr.dev/slog/sloggers/sloghuman"
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"os"
	"strings"

	"cdr.dev/slog"
)

var L = slog.Make(sloghuman.Sink(os.Stdout))

type Config struct {
	ConfigSentry ConfigSentry
}

func Initialize(c Config) {
	initializeSentry(c.ConfigSentry)
}

// Info logs a message at info level.
func Info(ctx context.Context, f interface{}, v ...interface{}) {
	message := formatLog(f, v...)
	L.Info(ctx, message)
}

// Error logs a message at error level.
func Error(ctx context.Context, err error) {
	slog.Helper()
	v := changeTypeError(err)
	// L.Error(ctx, v.ErrorStack())
	L.Error(ctx, fmt.Sprintf("%+v", v.Err))
	sendLogSentry(ctx, v)
}

// FatalError logs a message at critical level.
func FatalError(ctx context.Context, err error) {
	slog.Helper()

	v := changeTypeError(err)

	sendLogSentry(ctx, v)
	L.Fatal(ctx, v.ErrorStack())

}

func changeTypeError(err error) Exception {
	if v, ok := err.(Exception); ok {
		return v
	}
	return NewException(errors.WithStackDepth(err, 2), nil)

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
