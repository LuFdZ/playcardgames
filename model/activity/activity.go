package activity

import (
	dba "playcards/model/activity/db"
	mda "playcards/model/activity/mod"
	"playcards/utils/db"

	"github.com/jinzhu/gorm"
)

func AddActivityConfig(ac *mda.ActivityConfig) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return dba.AddActivityConfig(tx, ac)
	})
}

func UpdateActivityConfig(ac *mda.ActivityConfig) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return dba.UpdateActivityConfig(tx, ac)
	})
}

func DeleteActivityConfig(ac *mda.ActivityConfig) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return dba.DeleteActivityConfig(tx, ac)
	})
}

func ActivityConfigList() ([]*mda.ActivityConfig, error) {
	return dba.ActivityConfigList(db.DB())
}
