package db

import (
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	_gorm_db *gorm.DB
)

// support pg and mysql
func InitGorm(addr string, maxIdle, maxOpen int) (err error) {
	attrs := strings.Split(addr, ":")
	var connector string
	if attrs[0] == "postgresql" {
		connector = "postgres"
	} else {
		connector = "mysql"
	}

	_gorm_db, err = gorm.Open(connector, addr)
	if err != nil {
		return
	}

	_gorm_db.LogMode(false)
	_gorm_db.SingularTable(true)

	_gorm_db.DB().SetMaxIdleConns(maxIdle)
	_gorm_db.DB().SetMaxOpenConns(maxOpen)
	_gorm_db.DB().Ping()

	return
}

// custom needs with the db conn
func GormDB() *gorm.DB {
	return _gorm_db
}

// common method ...
func Upsert(object interface{}, fields, wheres map[string]interface{}) error {
	return _gorm_db.Where(wheres).Assign(fields).FirstOrCreate(object).Error
}

func SelectOne(object interface{}, wheres map[string]interface{}) error {
	return _gorm_db.Where(wheres).First(object).Error
}
