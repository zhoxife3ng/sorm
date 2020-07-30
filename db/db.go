package db

import (
	"database/sql"
	"fmt"
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
		dbInstance, err = sql.Open("mysql",
			fmt.Sprintf(
				"%s:%s@tcp(%s:%d)/%s?charset=utf8&allowCleartextPasswords=true&interpolateParams=true&loc=Local&parseTime=true",
				conf.User, conf.Password, conf.Host, conf.Port, conf.Name,
			),
		)
		if err == nil {
			if err = dbInstance.Ping(); err == nil {
				return
			}
		}
		log.Fatalln("dbHelper.DbInstance,", err)
	}
}

func SetupSlave(conf Conf) {
	if dbInstanceSlave == nil {
		var err error
		dbInstanceSlave, err = sql.Open("mysql",
			fmt.Sprintf(
				"%s:%s@tcp(%s:%d)/%s?charset=utf8&allowCleartextPasswords=true&interpolateParams=true",
				conf.User, conf.Password, conf.Host, conf.Port, conf.Name,
			),
		)
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
