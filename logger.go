package orm

import (
	"context"
	"errors"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-haru/field"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	"github.com/go-haru/log"
)

type loggerOptions struct {
	Tracing        bool     `json:"tracing" yaml:"tracing"`
	IgnoreNotfound bool     `json:"ignoreNotfound" yaml:"ignoreNotfound"` // true=ignore
	SlowThreshold  Duration `json:"slowThreshold" yaml:"slowThreshold"`   // by ms, 0=ignore
	Level          string   `json:"level" yaml:"level"`                   // none/info/warn/error
}

func (o *loggerOptions) ParseLevel() gormLogger.LogLevel {
	if o.Level == "off" || o.Level == "none" || o.Level == "silent" {
		return gormLogger.Silent
	}
	switch level, _ := log.ParseLevel(o.Level); level {
	default:
		fallthrough
	case log.DebugLevel:
		return gormLogger.Info + 1
	case log.InfoLevel:
		return gormLogger.Info
	case log.WarningLevel:
		return gormLogger.Warn
	case log.ErrorLevel:
		return gormLogger.Error
	}
}

type logger struct {
	tracing        bool
	ignoreNotfound bool
	slowThreshold  time.Duration
	level          gormLogger.LogLevel
}

func (l *logger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return &logger{level: level}
}

func (l *logger) Debug(ctx context.Context, s string, i ...interface{}) {
	if l.level >= gormLogger.Info+1 {
		log.C(ctx).AddDepth(2).Debugf(s, i...)
	}
}

func (l *logger) Info(ctx context.Context, s string, i ...interface{}) {
	if l.level >= gormLogger.Info {
		log.C(ctx).AddDepth(2).Infof(s, i...)
	}
}

func (l *logger) Warn(ctx context.Context, s string, i ...interface{}) {
	if l.level >= gormLogger.Warn {
		log.C(ctx).AddDepth(2).Warnf(s, i...)
	}
}

func (l *logger) Error(ctx context.Context, s string, i ...interface{}) {
	if l.level >= gormLogger.Error {
		log.C(ctx).AddDepth(2).Errorf(s, i...)
	}
}

var mainPkg string

func init() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		mainPkg = buildInfo.Main.Path
	}
}

func (l *logger) logger(ctx context.Context, elapsed time.Duration, row int64, sql string, caller string) log.Logger {
	var fields = []field.Field{
		field.String("__caller__", caller), field.Duration("elapsed", elapsed), field.String("sql", sql),
	}
	if row >= 0 {
		fields = append(fields, field.Int("affected", row))
	}
	return log.C(ctx).With(fields...)
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if !l.tracing || l.level <= gormLogger.Silent {
		return
	}
	var sql, rows = fc()
	var elapsed = time.Since(begin)
	var caller = strings.TrimLeft(strings.TrimPrefix(utils.FileWithLineNum(), mainPkg), "/")
	var tracedLogger = l.logger(ctx, elapsed, rows, sql, caller)
	switch {
	case err != nil && !(l.ignoreNotfound && errors.Is(err, gormLogger.ErrRecordNotFound)):
		tracedLogger.Error("orm query error: ", err)
	case l.slowThreshold != 0 && elapsed > l.slowThreshold:
		tracedLogger.With(field.Duration("threshold", l.slowThreshold)).Warn("orm slow query")
	default:
		tracedLogger.Debug("orm query")
	}
}

func NewLogger(options *loggerOptions) gormLogger.Interface {
	return &logger{
		tracing:        options.Tracing,
		ignoreNotfound: options.IgnoreNotfound,
		slowThreshold:  options.SlowThreshold.Duration(),
		level:          options.ParseLevel(),
	}
}
