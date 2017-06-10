package db

import (
	mdlog "playcards/model/log/mod"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func AddClientErrorLog(tx *gorm.DB, log *mdlog.ClientErrorLog) error {
	err := tx.Create(log).Error
	if err != nil {
		return errors.Internal("add client error log failed", err)
	}
	return nil
}
