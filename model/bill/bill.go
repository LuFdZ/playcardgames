package bill

import (
	pbbill "playcards/proto/bill"
	dbbill "playcards/model/bill/db"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	enumcom "playcards/model/common/enum"
	cachebill "playcards/model/bill/cache"
	"playcards/utils/db"

	"github.com/jinzhu/gorm"
)

func GetUserBalance(uid int32, cointype int32) (*mdbill.Balance, error) {
	return dbbill.GetUserBalance(db.DB(), uid, cointype)
}

func GetAllUserBalance(uid int32) (*pbbill.Balance, error) {
	balances, err := dbbill.GetAllUserBalance(db.DB(), uid)
	if err != nil {
		return nil, err
	}
	pbb := &pbbill.Balance{}
	for _, b := range balances {
		if b.CoinType == enumcom.Gold {
			pbb.Gold = b.Balance
		} else if b.CoinType == enumcom.Diamond {
			pbb.Diamond = b.Balance
		}
	}
	return pbb, nil
}

func GainBalance(uid int32, aid int64, balanceType int32, balance *mdbill.Balance) (
	*mdbill.Balance, error) {
	f := func(tx *gorm.DB) error {
		ub, err := dbbill.GainBalance(tx, uid, balance, balanceType,
			aid, enumbill.SystemOpUserID, enumbill.DefaultChannel)
		if err != nil {
			return err
		}
		err = cachebill.SetUserBalance(uid, ub)
		if err != nil {
			return err
		}
		return nil
	}
	if err := db.Transaction(f); err != nil {
		return nil, err
	}

	b, err := cachebill.GetUserBalance(uid, balance.CoinType)
	//b, err := dbbill.GetUserBalance(db.DB(), uid, balance.CoinType)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func GainGameBalance(uid int32, aid int32, balanceType int32, FreezeType int32, balance *mdbill.Balance) error {
	f := func(tx *gorm.DB) error {
		err := dbbill.SetBalanceFreeze(tx, uid, balance, FreezeType, int64(aid), enumbill.SystemOpUserID)
		if err != nil {
			return err
		}
		ub, err := dbbill.GainBalance(tx, uid, balance, balanceType,
			int64(aid), enumbill.SystemOpUserID, enumbill.DefaultChannel)
		if err != nil {
			return err
		}
		err = cachebill.SetUserBalance(uid, ub)
		if err != nil {
			return err
		}
		return nil
	}
	if err := db.Transaction(f); err != nil {
		return err
	}
	return nil
}

func Recharge(uid int32, aid int32, diamond int64, orderid int64,
	rechangeType int32, channel string) (int32, *mdbill.Balance, error) {
	exist := CheckDiamondBalanceIsDone(uid, orderid)
	if exist == enumbill.OrderExist {
		return enumbill.OrderExist, nil, nil
	}
	f := func(tx *gorm.DB) error {
		balance := &mdbill.Balance{Amount: diamond, CoinType: enumcom.Diamond}
		ub, err := dbbill.GainBalance(tx, uid, balance,
			rechangeType, orderid, aid, channel)
		if err != nil {
			return err
		}
		err = cachebill.SetUserBalance(uid, ub)
		if err != nil {
			return err
		}
		return nil
	}
	if err := db.Transaction(f); err != nil {
		return enumbill.OrderFail, nil, err
	}
	//b, err := dbbill.GetUserBalance(db.DB(), uid, enumcom.Diamond)
	b, err := cachebill.GetUserBalance(uid, enumcom.Diamond)
	if err != nil {
		return 0, nil, err
	}
	return enumbill.OrderSuccess, b, nil
}

func CheckDiamondBalanceIsDone(uid int32, orderid int64) int32 {
	return dbbill.GetJournal(db.DB(), uid, orderid, enumcom.Diamond)
}

func SetBalanceFreeze(uid int32, aid int64, balance *mdbill.Balance, balanceType int32,
) (*mdbill.Balance, error) {
	f := func(tx *gorm.DB) error {
		err := dbbill.SetBalanceFreeze(tx, uid, balance, balanceType,
			int64(aid), enumbill.SystemOpUserID)
		if err != nil {
			return err
		}
		return nil
	}
	if err := db.Transaction(f); err != nil {
		return nil, err
	}
	b, err := dbbill.GetUserBalance(db.DB(), uid, balance.CoinType)
	if err != nil {
		return nil, err
	}
	return b, nil
}

//func SetBalanceUnFreeze(uid int32, aid int64, balance *mdbill.Balance, balanceType int32,
//) (*mdbill.Balance, error) {
//	f := func(tx *gorm.DB) error {
//		err := dbbill.SetBalanceUnFreeze(tx, uid, balance, balanceType,
//			int64(aid), enumbill.SystemOpUserID)
//		if err != nil {
//			return err
//		}
//		return nil
//	}
//	if err := db.Transaction(f); err != nil {
//		return nil,err
//	}
//	b, err := dbbill.GetUserBalance(db.DB(), uid, balance.CoinType)
//	if err != nil {
//		return nil,err
//	}
//	return b,nil
//}

//func GainBalanceType(uid int32, aid int64, balance *mdbill.Balance, balanceType int32) (
//	*mdbill.Balance, error) {
//	f := func(tx *gorm.DB) error {
//		err := dbbill.GainBalance(tx, uid, balance, balanceType,
//			aid, enumbill.SystemOpUserID, enumbill.DefaultChannel)
//		if err != nil {
//			return err
//		}
//		return nil
//	}
//	if err := db.Transaction(f); err != nil {
//		return nil, err
//	}
//	b, err := dbbill.GetUserBalance(db.DB(), uid,balance.CoinType)
//	if err != nil {
//		return nil, err
//	}
//	return b, nil
//}

//func GainBalanceCondition(uid int32, channel string, version string, mobileOs string, aid int64, balance *mdbill.Balance, balanceType int32) (
//	*mdbill.Balance, error) {
//	rate := CheckConfigCondition(channel, version, mobileOs)
//	balance.Amount = int64(rate * float64(balance.Amount))
//	//fmt.Printf("GainBalanceCondition:%f",balance.Diamond)
//	f := func(tx *gorm.DB) error {
//		err := dbbill.GainBalance(tx, uid, balance, balanceType,
//			aid, enumbill.SystemOpUserID, enumbill.DefaultChannel)
//		if err != nil {
//			return err
//		}
//		return nil
//	}
//	if err := db.Transaction(f); err != nil {
//		return nil, err
//	}
//	b, err := dbbill.GetUserBalance(db.DB(), uid,balance.CoinType)
//	if err != nil {
//		return nil, err
//	}
//	return b, nil
//}

//func CheckConfigCondition(channel string, version string, mobileOs string) float64 {
//	rate := 1.00
//	cm := config.GetConfigs(channel, version, mobileOs)
//	for itemID, co := range cm {
//		if itemID == enumconf.ConsumeOpen {
//			value, _ := strconv.Atoi(co.ItemValue)
//			rate = float64(value) / 100
//		}
//	}
//	return rate
//}
