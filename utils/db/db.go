package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var defaultDB *gorm.DB

func Open(debug bool, dialect string, args ...interface{}) error {
	db, err := gorm.Open(dialect, args...)
	if err != nil {
		return err
	}
	if debug {
		db.LogMode(true)
	}
	db.DB().SetMaxOpenConns(100)
	defaultDB = db
	return nil
}

func DB() *gorm.DB {
	return defaultDB
}

func Begin() *gorm.DB {
	return defaultDB.Begin()
}

func Commit() *gorm.DB {
	return defaultDB.Commit()
}

func Rollback() *gorm.DB {
	return defaultDB.Rollback()
}
