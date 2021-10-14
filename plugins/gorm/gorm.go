package gorm

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"gorm.io/gorm/logger"
)

type ConfigPluginGorm struct {
	EnableTrace bool
	LogLevel    logger.LogLevel
}
type PluginGorm struct {
	config *ConfigPluginGorm
}

var oldLogger = logger.Default

func NewPluginGorm(c *ConfigPluginGorm) *PluginGorm {
	return &PluginGorm{config: c}
}

func (p PluginGorm) Initialize() {

	if !p.config.EnableTrace {
		return
	}

	logger.Default = &loggerDB{
		parent: oldLogger.LogMode(p.config.LogLevel),
	}
}

type loggerDB struct {
	parent logger.Interface
}

func (l *loggerDB) LogMode(level logger.LogLevel) logger.Interface {
	l.parent.LogMode(level)
	return l
}

func (l *loggerDB) Info(ctx context.Context, s string, i ...interface{}) {
	l.parent.Info(ctx, s, i)
}

func (l *loggerDB) Warn(ctx context.Context, s string, i ...interface{}) {
	l.parent.Warn(ctx, s, i)

}

func (l *loggerDB) Error(ctx context.Context, s string, i ...interface{}) {
	l.parent.Error(ctx, s, i)

}

func (l *loggerDB) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	l.parent.Trace(ctx, begin, fc, err)

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		// Нет смысла пихать в context если нету
		return
	}
	sql, _ := fc()
	elapsed := time.Since(begin)

	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "sql.query",
		Message:  sql,
		Level:    sentry.LevelInfo,
		Data: map[string]interface{}{
			"executionTimeMs": float64(elapsed.Nanoseconds()) / 1e6,
			// "connectionName":  l.connectionName,
		},
	}, nil)

}
