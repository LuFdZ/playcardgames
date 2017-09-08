package activity

import (
	dba "playcards/model/activity/db"
	"playcards/model/activity/errors"
	mda "playcards/model/activity/mod"
	mdu "playcards/model/user/mod"
	"playcards/utils/date"
	"playcards/utils/db"
	"time"

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

func Invite(u *mdu.User) (int32, int64, error) {
	if u.InviteUserID > 0 {
		return 0, 0, errors.ErrHadInviter
	}
	if date.TimeSubDays(*u.CreatedAt, time.Now()) > 1 {
		return 0, 0, errors.ErrInviterNotNewUser
	}
	return dba.ActivityConfigList(db.DB())
}

// message InviteReply{
//     int32 Result = 1;
//     int64 Diamond = 3;
// }
