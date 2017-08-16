package bill

import (
	dbbill "playcards/model/bill/db"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	"playcards/utils/db"
	"strconv"

	"github.com/jinzhu/gorm"
)

func GetUserBalance(uid int32) (*mdbill.UserBalance, error) {
	return dbbill.GetUserBalance(db.DB(), uid)
}

func GainBalance(uid int32, aid int32, balance *mdbill.Balance) (
	*mdbill.UserBalance, error) {
	f := func(tx *gorm.DB) error {
		err := dbbill.GainBalance(tx, uid, balance, enumbill.JournalTypeDash,
			strconv.Itoa(int(aid)), enumbill.SystemOpUserID)
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

func Recharge(uid int32, aid int32, diamond int64, orderid string,
	rechangeType int32) (int32, error) {
	exist := CheckBalanceIsDone(uid, orderid)
	if exist == enumbill.OrderExist {
		return enumbill.OrderExist, nil
	}
	f := func(tx *gorm.DB) error {
		balance := &mdbill.Balance{0, 0, diamond}
		err := dbbill.GainBalance(tx, uid, balance,
			rechangeType, orderid, aid)
		if err != nil {
			return err
		}
		return nil
	}
	if err := db.Transaction(f); err != nil {
		return enumbill.OrderFail, err
	}
	return enumbill.OrderSuccess, nil
}

func CheckBalanceIsDone(uid int32, orderid string) int32 {
	//var result int32 = 0
	// f := func(tx *gorm.DB) error {
	// 	result = dbbill.GetJournal(tx, uid, systemcode)
	// 	return nil
	// }

	// if err := db.Transaction(f); err != nil {
	// 	log.Err("get Journal error!%v", err)
	// 	return enumbill.OrderExist
	// }
	// return result

	return dbbill.GetJournal(db.DB(), uid, orderid)
}
