# db-orm

[![Go Reference](https://pkg.go.dev/badge/github.com/go-haru/db-orm.svg)](https://pkg.go.dev/github.com/go-haru/db-orm)
[![License](https://img.shields.io/github/license/go-haru/db-orm)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/go-haru/db-orm.svg?style=flat-square)](https://github.com/go-haru/db-orm/releases)
[![Go Test](https://github.com/go-haru/db-orm/actions/workflows/go.yml/badge.svg)](https://github.com/go-haru/db-orm/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-haru/db-orm)](https://goreportcard.com/report/github.com/go-haru/db-orm)

This package provides configurable factory wrapper of gorm, and with mysql driver bundled.

## Usage

```go
package main

import orm "github.com/go-haru/db-orm"

func init() {
    var cfg = orm.Config{
        Driver: "mysql",
        Mysql: orm.MysqlConfig{
            DSN: "mysql:// ........",
        },
    }
    if globalORM, err := orm.Instantiate(&cfg); err != nil {
        // handle error
    }
}
```

## Contributing

For convenience of PM, please commit all issue to [Document Repo](https://github.com/go-haru/go-haru/issues).

## License

This project is licensed under the `Apache License Version 2.0`.

Use and contributions signify your agreement to honor the terms of this [LICENSE](./LICENSE).

Commercial support or licensing is conditionally available through organization email.
