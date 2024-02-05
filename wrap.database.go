package orm

import (
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

type Database struct {
	sql *sql.DB
	orm *gorm.DB
}

func (c Database) OrmDB() *gorm.DB { return c.orm }

func (c Database) SqlDB() *sql.DB { return c.sql }

func (c Database) Transaction(handler func(Database) error, opts *sql.TxOptions) error {
	return c.orm.Transaction(func(txReal *gorm.DB) error {
		return handler(Database{orm: txReal})
	}, opts)
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, gorm.ErrRecordNotFound)
}
