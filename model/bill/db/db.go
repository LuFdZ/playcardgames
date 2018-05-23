package db

import (
	"playcards/model/bill/enum"
	errbill "playcards/model/bill/errors"
	mdbill "playcards/model/bill/mod"
	"playcards/utils/db"
	"playcards/utils/errors"
	enumcom "playcards/model/common/enum"
	sq "github.com/Masterminds/squirrel"
	mdpage "playcards/model/page"
	"github.com/jinzhu/gorm"
	"strconv"
)

func CreateAllBalance(tx *gorm.DB, uid int32) error {
	b := &mdbill.Balance{
		UserID:   uid,
		CoinType: enumcom.Gold,
		Amount:   enumcom.NewUserGold,
		Balance:  enumcom.NewUserGold,
	}
	err := CreateBalance(tx, uid, b)
	if err != nil {
		return err
	}
	b = &mdbill.Balance{
		UserID:   uid,
		CoinType: enumcom.Diamond,
		Amount:   enumcom.NewUserDiamond,
		Balance:  enumcom.NewUserDiamond,
	}
	err = CreateBalance(tx, uid, b)
	if err != nil {
		return err
	}
	return nil
}
func CreateBalance(tx *gorm.DB, uid int32, balance *mdbill.Balance) error {
	var err error

	if err = tx.Create(balance).Error; err != nil {
		return errors.Internal("create balanceerr", err)
	}

	err = InsertJournal(tx, uid, balance, &mdbill.Balance{UserID: uid}, enum.JournalTypeInitBalance,
		strconv.Itoa(int(balance.UserID)), balance.UserID, enum.DefaultChannel)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	return nil
}

func GetUserBalance(tx *gorm.DB, uid int32, cointype int32) (*mdbill.Balance, error) {
	out := &mdbill.Balance{}
	if err := tx.Where(" user_id = ? and coin_type = ?", uid, cointype).Find(out).Error; err != nil {
		return nil, errors.Internal("get balance failed", err)
	}
	return out, nil
}

func GetAllUserBalance(tx *gorm.DB, uid int32) ([]*mdbill.Balance, error) {
	var out []*mdbill.Balance
	if err := tx.Where(" user_id = ? ", uid).Find(&out).Error; err != nil {
		return nil, errors.Internal("get balance failed", err)
	}
	return out, nil
}

func GetLockUserBalance(tx *gorm.DB, uid int32, cointype int32) (*mdbill.Balance, error) {
	out := &mdbill.Balance{}
	if err := db.ForUpdate(tx).Where("user_id = ? and coin_type = ?", uid, cointype).Find(out).
		Error; err != nil {
		return nil, errors.Internal("get balance failed", err)
	}
	return out, nil
}

func GetJournal(tx *gorm.DB, uid int32, orderid string, cointype int32) int32 {
	out := &mdbill.Journal{}
	if found, _ := db.FoundRecord(tx.Where("user_id = ? and `foreign` = ? and coin_type = ?",
		uid, orderid, cointype).Find(&out).Error); found {
		return 2
	}
	return 1
}

// Type:
// JournalTypeInitBalance -> Foreign deposit.id
// JournalTypeDeposit -> Foreign deposit.id
// JournalTypeMap -> Foreign map_profits.id
func InsertJournal(tx *gorm.DB, uid int32, b *mdbill.Balance, bbfore *mdbill.Balance,
	typ int32, fid string, opuid int32, channel string) error {
	now := gorm.NowFunc()

	//amount := bnow.Amount + bbfore.Amount
	m := make(map[string]interface{})
	m["coin_type"] = b.CoinType
	m["amount"] = b.Amount
	m["amount_before"] = bbfore.Amount
	m["amount_after"] = bbfore.Amount + b.Amount
	m["user_id"] = uid
	m["type"] = typ
	m["`foreign`"] = fid
	m["created_at"] = now
	m["updated_at"] = now
	m["op_user_id"] = opuid
	m["channel"] = channel

	//bbfore.Amount = amount
	sql, args, _ := sq.Insert(enum.JournalTableName).SetMap(m).ToSql()
	if err := tx.Exec(sql, args...).Error; err != nil {
		return errors.Internal("save journal failed", err)
	}

	return nil
}

func GainBalance(tx *gorm.DB, uid int32, b *mdbill.Balance, typ int32,
	fid string, opuid int32, channel string) (*mdbill.Balance, error) {
	if b.Amount == 0 {
		return nil, errbill.ErrNotAllowAmount
	}

	ub, err := GetLockUserBalance(tx, uid, b.CoinType)
	if err != nil {
		return nil, err
	}

	balance := ub.Amount + b.Amount - ub.Freeze - ub.Deposit
	if balance < 0 {
		return nil, errbill.ErrOutOfBalance
	}
	err = InsertJournal(tx, uid, b, ub, typ, fid, opuid, channel)
	if err != nil {
		return nil, err
	}
	ub.Amount += b.Amount
	ub.Balance = ub.Amount - ub.Freeze - ub.Deposit
	if err = tx.Save(ub).Error; err != nil {
		return nil, errors.Internal("gain balance failed", err)
	}
	return ub, nil
}

func SetBalanceFreeze(tx *gorm.DB, uid int32, b *mdbill.Balance, typ int32,
	fid string, opuid int32) error {
	ub, err := GetLockUserBalance(tx, uid, b.CoinType)
	if err != nil {
		return err
	}
	//fmt.Printf("SetBalanceFreezeTest:%+v|%+v\n",b,ub)
	freeze := ub.Freeze + b.Amount
	if freeze < 0 {
		return errbill.ErrFreezeAmount
	}
	balance := ub.Amount - freeze - ub.Deposit
	if balance < 0 {
		return errbill.ErrOutOfBalance
	}
	//ub.Freeze -= b.Amount

	err = InsertJournal(tx, uid, b, &mdbill.Balance{Amount: ub.Freeze}, typ, fid, opuid, enum.DefaultChannel)
	if err != nil {
		return err
	}
	ub.Freeze += b.Amount
	ub.Balance = balance
	if err = tx.Save(ub).Error; err != nil {
		return errors.Internal("gain balance failed", err)
	}
	return nil
}

func PageJournal(tx *gorm.DB, mdj *mdbill.Journal, page *mdpage.PageOption,
) ([]*mdbill.Journal, int64, error) {
	var out []*mdbill.Journal
	rows, rtx := page.Find(tx.Where(mdj).
		Order("created_at desc").Find(&out), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page journal failed", rtx.Error)
	}
	return out, rows, nil
}

//func SetBalanceUnFreeze(tx *gorm.DB, uid int32, b *mdbill.Balance, typ int32,
//	fid int64, opuid int32) error {
//	ub, err := GetLockUserBalance(tx, uid, b.CoinType)
//	if err != nil {
//		return err
//	}
//
//	err = InsertJournal(tx, uid, b, &mdbill.Balance{Amount: ub.Freeze}, typ, fid, opuid, enum.DefaultChannel)
//	if err != nil {
//		return err
//	}
//	ub.Freeze += b.Amount
//	ub.Balance = ub.Amount - ub.Freeze - ub.Deposit
//	//ub.Amount -= b.Amount
//	if err = tx.Save(ub).Error; err != nil {
//		return errors.Internal("gain balance failed", err)
//	}
//	return nil
//}
