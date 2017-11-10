package bill

import (
	dbbill "playcards/model/bill/db"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	config "playcards/model/config"
	"playcards/utils/db"
	"strconv"
	enumconf "playcards/model/config/enum"
	"github.com/jinzhu/gorm"
	"fmt"
)

func GetUserBalance(uid int32) (*mdbill.UserBalance, error) {
	return dbbill.GetUserBalance(db.DB(), uid)
}

func GainBalance(uid int32, aid int32, balance *mdbill.Balance) (
	*mdbill.UserBalance, error) {
	f := func(tx *gorm.DB) error {
		err := dbbill.GainBalance(tx, uid, balance, enumbill.JournalTypeDash,
			strconv.Itoa(int(aid)), enumbill.SystemOpUserID, enumbill.DefaultChannel)
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
	rechangeType int32, channel string) (int32, *mdbill.UserBalance, error) {
	exist := CheckBalanceIsDone(uid, orderid)
	if exist == enumbill.OrderExist {
		return enumbill.OrderExist, nil, nil
	}
	f := func(tx *gorm.DB) error {
		balance := &mdbill.Balance{0, 0, diamond}
		err := dbbill.GainBalance(tx, uid, balance,
			rechangeType, orderid, aid, channel)
		if err != nil {
			return err
		}
		return nil
	}
	if err := db.Transaction(f); err != nil {
		return enumbill.OrderFail, nil, err
	}
	b, err := dbbill.GetUserBalance(db.DB(), uid)

	if err != nil {
		return 0, nil, err
	}
	return enumbill.OrderSuccess, b, nil
}

func CheckBalanceIsDone(uid int32, orderid string) int32 {
	return dbbill.GetJournal(db.DB(), uid, orderid)
}

func GainBalanceType(uid int32, aid int64, balance *mdbill.Balance, balanceType int32) (
	*mdbill.UserBalance, error) {
	f := func(tx *gorm.DB) error {
		err := dbbill.GainBalance(tx, uid, balance, balanceType,
			strconv.Itoa(int(aid)), enumbill.SystemOpUserID, enumbill.DefaultChannel)
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

func GainBalanceCondition(uid int32, channel string, version string, mobileOs string, aid string, balance *mdbill.Balance, balanceType int32) (
	*mdbill.UserBalance, error) {
	rate := CheckConfigCondition(channel, version, mobileOs)
	balance.Diamond = int64(rate * float64(balance.Diamond))
	fmt.Printf("GainBalanceCondition:%f",balance.Diamond)
	f := func(tx *gorm.DB) error {
		err := dbbill.GainBalance(tx, uid, balance, balanceType,
			aid, enumbill.SystemOpUserID, enumbill.DefaultChannel)
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

func CheckConfigCondition(channel string, version string, mobileOs string) float64 {
	rate := 1.00
	cm :=config.GetConfigs(channel, version, mobileOs)
	for itemID,co :=range cm{
		if itemID == enumconf.ConsumeOpen{
			value, _ := strconv.Atoi(co.ItemValue)
			rate = float64(value) / 100
		}
	}
	return rate
}
