package db

import (
	"github.com/go-sql-driver/mysql"
	"time"
)

type Option func(conf *mysql.Config)

func Param(key, value string) Option {
	return func(conf *mysql.Config) {
		if conf.Params == nil {
			conf.Params = make(map[string]string)
		}
		conf.Params[key] = value
	}
}

func Loc(loc *time.Location) Option {
	return func(conf *mysql.Config) {
		conf.Loc = loc
	}
}

func ParseTime(ok bool) Option {
	return func(conf *mysql.Config) {
		conf.ParseTime = ok
	}
}

func AllowCleartextPasswords(ok bool) Option {
	return func(conf *mysql.Config) {
		conf.AllowCleartextPasswords = ok
	}
}

func InterpolateParams(ok bool) Option {
	return func(conf *mysql.Config) {
		conf.InterpolateParams = ok
	}
}

func Timeout(timeout time.Duration) Option {
	return func(conf *mysql.Config) {
		conf.Timeout = timeout
	}
}

func ReadTimeout(timeout time.Duration) Option {
	return func(conf *mysql.Config) {
		conf.ReadTimeout = timeout
	}
}
func WriteTimeout(timeout time.Duration) Option {
	return func(conf *mysql.Config) {
		conf.WriteTimeout = timeout
	}
}
