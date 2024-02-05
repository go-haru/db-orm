package orm

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/go-haru/errors"
	"github.com/go-haru/field"
)

var pureNumberRegex = regexp.MustCompile(`^\d+$`)

func parseDuration(raw string, fallback time.Duration) (time.Duration, error) {
	if raw == "" {
		return fallback, nil
	}
	if pureNumberRegex.MatchString(raw) {
		raw = raw + "ms"
	}
	return time.ParseDuration(raw)
}

type Config struct {
	DryRun      bool          `json:"dryRun" yaml:"dryRun"`
	Migration   bool          `json:"migration" yaml:"migration"`
	UpdateAll   bool          `json:"updateAll" yaml:"updateAll"`
	Driver      string        `json:"driver" yaml:"driver"`
	Prefix      string        `json:"prefix" yaml:"prefix"`
	Mysql       MysqlConfig   `json:"mysql" yaml:"mysql"`
	Logger      loggerOptions `json:"logger" yaml:"logger"`
	PingTimeout Duration      `json:"pingTimeout" yaml:"pingTimeout"`

	address string
}

type configFilter func(*Config) error

func (c *Config) finalize(filters ...configFilter) (err error) {
	for i, filter := range filters {
		if err = filter(c); err != nil {
			return errors.With(err, field.Int("index", i))
		}
	}
	return nil
}

func (c *Config) build() (driver gorm.Dialector, cfg *gorm.Config, err error) {
	switch driverName := fallback(strings.ToLower(c.Driver), DriverTypeMysql); driverName {
	case DriverTypeMysql:
		if driver, c.address, err = c.Mysql.Dialector(); err != nil {
			return nil, nil, errors.With(err, field.String("driver", driverName))
		}
	default:
		return nil, nil, errors.With(fmt.Errorf("invalid enum value"),
			field.String("ref", "driver"),
			field.String("got", c.Driver),
			field.Strings("expected", []string{DriverTypeMysql}))
	}
	cfg = &gorm.Config{
		DryRun:            c.DryRun,
		AllowGlobalUpdate: c.UpdateAll,
		Logger:            NewLogger(&c.Logger),
	}
	if c.Prefix != "" {
		cfg.NamingStrategy = schema.NamingStrategy{TablePrefix: c.Prefix}
	}
	if c.Driver == "" {
		c.Driver = driver.Name()
	}
	return driver, cfg, nil
}

func (c *Config) Address() string { return c.address }

func Instantiate(c *Config, filters ...configFilter) (_ *Database, err error) {
	if err = c.finalize(filters...); err != nil {
		return nil, errors.With(err)
	}
	var driver gorm.Dialector
	var ormConfig *gorm.Config
	if driver, ormConfig, err = c.build(); err != nil {
		return nil, errors.With(err)
	}

	var db *gorm.DB
	if db, err = gorm.Open(driver, ormConfig); err != nil {
		return nil, errors.With(err)
	}

	var sqlDb *sql.DB
	if sqlDb, err = db.DB(); err != nil {
		return nil, errors.With(fmt.Errorf("failed to connect database: %w", err))
	}

	if pingTimeout := c.PingTimeout.Duration(); pingTimeout > 0 {
		var pingCtx, pingCancel = context.WithTimeout(context.Background(), pingTimeout)
		defer pingCancel()
		if err = sqlDb.PingContext(pingCtx); err != nil {
			return nil, fmt.Errorf("cant ping to gorm server: %w", err)
		}
	}

	return &Database{orm: db, sql: sqlDb}, nil
}
