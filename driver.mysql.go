package orm

import (
	"crypto/rsa"
	"fmt"
	"sync/atomic"
	"time"

	nativeMysql "github.com/go-sql-driver/mysql"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/go-haru/db-orm/keys"
	"github.com/go-haru/errors"
	"github.com/go-haru/field"
	"github.com/go-haru/network"
)

const (
	DriverTypeMysql = "mysql"
)

var dialectorIdCounter uint32

type MysqlConfig struct {
	DSN          string            `json:"dsn" yaml:"dsn"`
	Network      string            `json:"network" yaml:"network"`
	Address      string            `json:"address" yaml:"address"`
	Username     string            `json:"username" yaml:"username"`
	Password     string            `json:"password" yaml:"password"`
	PublicKey    string            `json:"publicKey" yaml:"publicKey"`
	Database     string            `json:"database" yaml:"database"`
	Collation    string            `json:"collation" yaml:"collation"`
	Timezone     string            `json:"timezone" yaml:"timezone"`
	DialTimeout  string            `json:"dailTimeout" yaml:"dailTimeout"`
	ReadTimeout  string            `json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout string            `json:"writeTimeout" yaml:"writeTimeout"`
	NoReadonly   bool              `json:"noReadonly" yaml:"noReadonly"`
	TlsConfig    network.TlsConfig `json:"tls" yaml:"tls"`
	Params       map[string]string `json:"params" yaml:"params"`
}

func (c *MysqlConfig) Dialector() (gorm.Dialector, string, error) {
	var err error
	var tlsConfigName, serverPubKeyName string

	var dialectorId = atomic.AddUint32(&dialectorIdCounter, 1)

	const NameTLSConfSuffix = "_tls"
	if !c.TlsConfig.Disable {
		var tlsCfg = network.DefaultTlsConfig()
		if err = c.TlsConfig.Apply(tlsCfg); err != nil {
			return nil, c.Address, errors.With(err)
		}
		tlsConfigName = fmt.Sprintf("%08x%s", dialectorId, NameTLSConfSuffix)
		if err = nativeMysql.RegisterTLSConfig(tlsConfigName, tlsCfg); err != nil {
			return nil, c.Address, errors.With(err)
		}
	}

	const NameSrvKeySuffix = "_key"
	if c.PublicKey != "" {
		var rsaPubKey *rsa.PublicKey
		if rsaPubKey, err = keys.ParseRSAPublicKeyPEM([]byte(c.PublicKey)); err != nil {
			return nil, c.Address, errors.With(err)
		} else {
			serverPubKeyName = fmt.Sprintf("%08x%s", dialectorId, NameSrvKeySuffix)
			nativeMysql.RegisterServerPubKey(serverPubKeyName, rsaPubKey)
		}
	}

	var cfg = &nativeMysql.Config{}

	if c.DSN != "" {
		if cfg, err = nativeMysql.ParseDSN(c.DSN); err != nil {
			return nil, c.Address, errors.With(err, field.String("dsn", c.DSN))
		}
	}

	cfg.User = fallback(c.Username, cfg.User)
	c.Username = cfg.User

	cfg.Passwd = fallback(c.Password, cfg.Passwd)
	c.Password = cfg.Passwd

	cfg.Net = fallback(c.Network, cfg.Net, "tcp")
	c.Network = cfg.Net

	cfg.Addr = fallback(c.Address, cfg.Addr)
	c.Address = cfg.Addr
	if cfg.Addr == "" {
		return nil, "", errors.With(fmt.Errorf("address cant be empty"), field.String("ref", "address"))
	}

	cfg.DBName = fallback(c.Database, cfg.DBName)
	c.Database = cfg.DBName

	cfg.Collation = fallback(c.Collation, cfg.Collation, "utf8mb4")
	c.Collation = cfg.Collation

	if c.DialTimeout != "" {
		if cfg.Timeout, err = parseDuration(c.DialTimeout, 0); err != nil {
			return nil, c.Address, errors.With(fmt.Errorf("cant parse time duration"), field.String("ref", "dailTimeout"))
		}
	}
	if c.ReadTimeout != "" {
		if cfg.ReadTimeout, err = parseDuration(c.ReadTimeout, 0); err != nil {
			return nil, c.Address, errors.With(fmt.Errorf("cant parse time duration"), field.String("ref", "readTimeout"))
		}
	}
	if c.WriteTimeout != "" {
		if cfg.WriteTimeout, err = parseDuration(c.WriteTimeout, 0); err != nil {
			return nil, c.Address, errors.With(fmt.Errorf("cant parse time duration"), field.String("ref", "writeTimeout"))
		}
	}

	cfg.TLSConfig = tlsConfigName
	cfg.ServerPubKey = serverPubKeyName
	cfg.ParseTime = true
	cfg.RejectReadOnly = c.NoReadonly
	cfg.CheckConnLiveness = true
	cfg.AllowNativePasswords = true

	if cfg.Loc, err = time.LoadLocation(fallback(c.Timezone, "Local")); err != nil {
		return nil, c.Address, errors.With(err)
	}

	if len(c.Params) > 0 {
		var copyOfParams = make(map[string]string, len(c.Params))
		for k, v := range c.Params {
			copyOfParams[k] = v
		}
		cfg.Params = copyOfParams
	}

	var dialector = gormMysql.New(gormMysql.Config{DSN: cfg.FormatDSN(), DefaultStringSize: 256})

	return dialector, c.Address, nil
}
