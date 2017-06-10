package bill

import (
	dbbill "playcards/model/bill/db"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	"playcards/utils/db"

	"github.com/jinzhu/gorm"
)

func GetUserBalance(uid int32) (*mdbill.UserBalance, error) {
	return dbbill.GetUserBalance(db.DB(), uid)
}

func GainBalance(uid int32, aid int32,
	balance *mdbill.Balance) (*mdbill.UserBalance, error) {
	f := func(tx *gorm.DB) error {
		err := dbbill.GainBalance(tx, uid, balance, enumbill.JournalTypeDash,
			int64(aid))
		if err != nil {
			return err
		}
		return nil
	}
	if err := db.Transaction(f); err != nil {
		return nil, err
	}
	b, err := dbbill.GetUserBalance(db.DB(), uid)
	if err != nil {
		return nil, err
	}
	return b, nil
}
