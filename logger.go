package orm

import (
	"context"
	"errors"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	"github.com/go-haru/log"
)

type loggerOptions struct {
	IgnoreNotfound bool     `json:"ignoreNotfound" yaml:"ignoreNotfound"` // true=ignore
	SlowThreshold  Duration `json:"slowThreshold" yaml:"slowThreshold"`   // by ms, 0=ignore
	Level          string   `json:"level" yaml:"level"`                   // none/info/warn/error
	Name           string   `json:"name" yaml:"name"`
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
	ignoreNotfound bool
	slowThreshold  time.Duration
	level          gormLogger.LogLevel
	underlying     log.Logger
}

func (l *logger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return &logger{
		level:      level,
		underlying: l.underlying,
	}
}

func (l *logger) Debug(_ context.Context, s string, i ...interface{}) {
	if l.level >= gormLogger.Info+1 {
		l.underlying.Debugf(s, i...)
	}
}

func (l *logger) Info(_ context.Context, s string, i ...interface{}) {
	if l.level >= gormLogger.Info {
		l.underlying.Infof(s, i...)
	}
}

func (l *logger) Warn(_ context.Context, s string, i ...interface{}) {
	if l.level >= gormLogger.Warn {
		l.underlying.Warnf(s, i...)
	}
}

func (l *logger) Error(_ context.Context, s string, i ...interface{}) {
	if l.level >= gormLogger.Error {
		l.underlying.Errorf(s, i...)
	}
}

var mainPkg string

func init() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		mainPkg = buildInfo.Main.Path
	}
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormLogger.Silent {
		return
	}
	var rowsStr string
	sql, rows := fc()
	if rows == -1 {
		rowsStr = "-"
	} else {
		rowsStr = strconv.FormatInt(rows, 10)
	}
	elapsed := time.Since(begin)
	caller := strings.TrimLeft(strings.TrimPrefix(utils.FileWithLineNum(), mainPkg), "/")
	switch {
	case err != nil && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.ignoreNotfound):
		l.Error(ctx, "error: %v (%d ms, %s row). # %s # %s", err, elapsed.Milliseconds(), rowsStr, sql, caller)
	case l.slowThreshold != 0 && elapsed > l.slowThreshold:
		l.Warn(ctx, "slow query (%d ms, %s row). # %s # %s", elapsed.Milliseconds(), l.slowThreshold, rowsStr, sql, caller)
	default:
		l.Debug(ctx, "query (%d ms, %s row). # %s # %s", elapsed.Milliseconds(), rowsStr, sql, caller)
	}
}

func NewLogger(options *loggerOptions) gormLogger.Interface {
	var underlying = log.Current().AddDepth(2)
	if options.Name != "" {
		underlying = underlying.WithName(options.Name)
	}
	return &logger{
		ignoreNotfound: options.IgnoreNotfound,
		slowThreshold:  options.SlowThreshold.Duration(),
		level:          options.ParseLevel(),
		underlying:     underlying,
	}
}
