package db

import (
	"database/sql"
	"github.com/didi/gendry/manager"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

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

func Setup(conf Conf) {
	if dbInstance == nil {
		var err error
		dbInstance, err = manager.New(conf.Name, conf.User, conf.Password, conf.Host).Set(
			manager.SetCharset("utf8"),
			manager.SetAllowCleartextPasswords(true),
			manager.SetInterpolateParams(true),
		).Port(conf.Port).Open(true)
		if err != nil {
			log.Fatalln("dbhelper.DbInstance,", err)
		}
	}
}

func SetupSlave(conf Conf) {
	if dbInstanceSlave == nil {
		var err error
		dbInstanceSlave, err = manager.New(conf.Name, conf.User, conf.Password, conf.Host).Set(
			manager.SetCharset("utf8"),
			manager.SetAllowCleartextPasswords(true),
			manager.SetInterpolateParams(true),
		).Port(conf.Port).Open(true)
		if err != nil {
			log.Fatalln("dbhelper.DbInstanceSlave,", err)
		}
	}
}

func GetInstance() *sql.DB {
	if dbInstance == nil {
		log.Fatalln("dbhelper.DbInstance, dbInstance is nil")
	}
	return dbInstance
}

func GetSlaveInstance() *sql.DB {
	if dbInstanceSlave == nil {
		return GetInstance()
	}
	return dbInstanceSlave
}
