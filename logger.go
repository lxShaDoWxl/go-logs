package logs

import (
	"context"
	"github.com/go-errors/errors"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

var (
	DefaultConfigLogger = &ConfigLogger{
		Name:  "app",
		Level: zerolog.InfoLevel.String(),
		JSON:  false,
	}
	zlog             = zerologr.New(initZLog(DefaultConfigLogger))
	pkgLogger Logger = LogRLogger{zlogger: zlog}
)

type Logger interface {
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, err error, keysAndValues ...interface{})
	Errorw(msg string, err error, keysAndValues ...interface{})
	Fatalw(msg string, err error, keysAndValues ...interface{})
	WithValues(keysAndValues ...interface{}) Logger
	WithName(name string) Logger
	WithCallDepth(depth int) Logger
}
type ConfigLogger struct {
	Name  string
	Level string
	JSON  bool
}
type Config struct {
	ConfigSentry ConfigSentry
	Logger       *ConfigLogger
}

func Initialize(conf *Config) {
	initializeSentry(conf.ConfigSentry)
	SetLogger(LogRLogger{zlogger: zerologr.New(initZLog(conf.Logger))}, conf.Logger.Name)
}

// GetLogger returns the logger that was set with SetLogger with an extra depth of 1.
func GetLogger() Logger {
	return LogRLogger{zlogger: zlog}
}

// SetLogger lets you use a custom logger. Pass in a logr.Logger with default depth.
func SetLogger(l Logger, name string) {
	pkgLogger = l.WithCallDepth(2).WithName(name)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	pkgLogger.Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	pkgLogger.Infow(msg, keysAndValues...)
}

func Warnw(ctx context.Context, msg string, err error, keysAndValues ...interface{}) {
	v := changeTypeError(err)
	sendLogSentry(ctx, v)
	pkgLogger.Warnw(msg, err, keysAndValues...)
}

func Errorw(ctx context.Context, msg string, err error, keysAndValues ...interface{}) {
	v := changeTypeError(err)
	sendLogSentry(ctx, v)
	pkgLogger.Errorw(msg, err, keysAndValues...)
}

func Fatalw(ctx context.Context, msg string, err error, keysAndValues ...interface{}) {
	v := changeTypeError(err)
	sendLogSentry(ctx, v)
	pkgLogger.Fatalw(msg, err, keysAndValues...)
}

// Info logs a message at info level.
func Info(ctx context.Context, f string, v ...interface{}) {
	pkgLogger.Infow(f, v...)
}

// Error logs a message at error level.
func Error(ctx context.Context, err error, keysAndValues ...interface{}) {
	v := changeTypeError(err)
	pkgLogger.Errorw(v.ErrorStack(), err, keysAndValues...)
	sendLogSentry(ctx, v)
}

// FatalError logs a message at critical level.
func FatalError(ctx context.Context, err error) {
	v := changeTypeError(err)
	sendLogSentry(ctx, v)
	pkgLogger.Fatalw("", err)
}

func changeTypeError(err error) ExceptionError {
	if v, ok := err.(ExceptionError); ok {
		return v
	}
	return NewException(errors.Wrap(err, 2), nil)
}
