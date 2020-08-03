package db

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"log"
)

const driverName = "mysql"

var (
	dbInstance      *sql.DB
	dbInstanceSlave *sql.DB
)

type Conf struct {
	Name     string
	User     string
	Password string
	Host     string
	Port     int
}

func buildDSN(conf Conf, opts []Option) string {
	mysqlConf := mysql.NewConfig()
	mysqlConf.DBName = conf.Name
	mysqlConf.User = conf.User
	mysqlConf.Passwd = conf.Password
	if conf.Port > 0 {
		mysqlConf.Addr = fmt.Sprintf("%s:%d", conf.Host, conf.Port)
		mysqlConf.Net = "tcp"
	} else {
		mysqlConf.Addr = conf.Host
		mysqlConf.Net = "unix"
	}
	for _, opt := range opts {
		opt(mysqlConf)
	}
	return mysqlConf.FormatDSN()
}

func Setup(conf Conf, opts ...Option) {
	if dbInstance == nil {
		var err error
		fmt.Println(buildDSN(conf, opts))
		dbInstance, err = sql.Open(driverName, buildDSN(conf, opts))
		if err == nil {
			if err = dbInstance.Ping(); err == nil {
				return
			}
		}
		log.Fatalln("dbHelper.DbInstance,", err)
	}
}

func SetupSlave(conf Conf, opts ...Option) {
	if dbInstanceSlave == nil {
		var err error
		dbInstanceSlave, err = sql.Open(driverName, buildDSN(conf, opts))
		if err == nil {
			if err = dbInstance.Ping(); err == nil {
				return
			}
		}
		log.Fatalln("dbHelper.DbInstanceSlave,", err)
	}
}

func GetInstance() *sql.DB {
	if dbInstance == nil {
		log.Fatalln("dbHelper.DbInstance, dbInstance is nil")
	}
	return dbInstance
}

func GetSlaveInstance() *sql.DB {
	if dbInstanceSlave == nil {
		return GetInstance()
	}
	return dbInstanceSlave
}
