package db

import (
	"playcards/utils/log"
	"runtime/debug"
	"github.com/jinzhu/gorm"
)

func Transaction(f func(*gorm.DB) error) error {
	gdb := Begin()

	defer func() {
		if err := recover(); err != nil {
			log.Crit("critical error in db transaction: %v\nerr:\n", err,string(debug.Stack()))
		}
	}()

	err := f(gdb)
	if err != nil {
		gdb.Rollback()
		log.Err("db transaction failed: %v\nerr:\n", err,string(debug.Stack()))
		return err
	}

	err = gdb.Commit().Error
	if err != nil {
		log.Crit("db transaction commit failed: %v\nerr:\n", err,string(debug.Stack()))
		return err
	}

	return nil
}

func FoundRecord(err error) (bool, error) {
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err == nil {
		return true, nil
	}

	return false, err
}

func RecordCount(tx *gorm.DB) (int64, *gorm.DB) {
	c := &struct {
		Size int64
	}{
		Size: 0,
	}

	tx = tx.Select("count(0) as `size`").Scan(c)
	return c.Size, tx
}

func SQLExpr(sql string, args []interface{}) (string, interface{}) {
	if len(args) > 0 {
		sql += " ?"
	}

	s := " "
	for _, _ = range args {
		s += " ? "
	}

	return sql, gorm.Expr(s, args...)
}

func ForUpdate(db *gorm.DB) *gorm.DB {
	return db.Set("gorm:query_option", "FOR UPDATE")
}
