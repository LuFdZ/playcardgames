package db

import (
	"playcards/model/bill/enum"
	errbill "playcards/model/bill/errors"
	mdbill "playcards/model/bill/mod"
	"playcards/utils/db"
	"playcards/utils/errors"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/jinzhu/gorm"
)

func CreateBalance(tx *gorm.DB, uid int32, balance *mdbill.Balance) error {
	var err error

	b := &mdbill.UserBalance{
		UserID:  uid,
		Balance: *balance,
	}

	if err = tx.Create(b).Error; err != nil {
		return errors.Internal("create balance", err)
	}

	err = InsertJournal(tx, uid, balance, enum.JournalTypeInitBalance,
		strconv.Itoa(int(b.UserID)), b.UserID)
	if err != nil {
		return err
	}

	return nil
}

func GetUserBalance(tx *gorm.DB, uid int32) (*mdbill.UserBalance, error) {
	out := &mdbill.UserBalance{}
	if err := tx.Where("user_id = ?", uid).Find(out).Error; err != nil {
		return nil, errors.Internal("get balance failed", err)
	}
	return out, nil
}

func GetLockUserBalance(tx *gorm.DB, uid int32) (*mdbill.UserBalance, error) {
	out := &mdbill.UserBalance{}
	if err := db.ForUpdate(tx).Where("user_id = ?", uid).Find(out).
		Error; err != nil {
		return nil, errors.Internal("get balance failed", err)
	}
	return out, nil
}

func GetJournal(tx *gorm.DB, uid int32, orderid string) int32 {
	out := &mdbill.Journal{}
	if found, _ := db.FoundRecord(tx.Where("user_id = ? and `foreign` = ? ",
		uid, orderid).Find(&out).Error); found {
		return 2
	}
	return 1
}

func Deposit(tx *gorm.DB, uid int32, amount int64, typ int32) error {
	if amount == 0 {
		return nil
	}

	if amount < 0 {
		return errbill.ErrInvalidParameter
	}

	d := &mdbill.Deposit{
		UserID: uid,
		Amount: amount,
		Type:   typ,
	}

	err := tx.Create(d).Error
	if err != nil {
		return errors.Internal("deposit failed", err)
	}

	b, err := GetLockUserBalance(tx, uid)
	if err != nil {
		return errors.Internal("deposit failed", err)
	}

	b.Amount += amount
	b.Deposit += amount
	err = tx.Save(b).Error
	if err != nil {
		return errors.Internal("deposit failed", err)
	}

	bal := &mdbill.Balance{
		Amount: amount,
	}
	return InsertJournal(tx, b.UserID, bal, typ, string(d.ID),
		enum.SystemOpUserID)
}

// Type:
// JournalTypeInitBalance -> Foreign deposit.id
// JournalTypeDeposit -> Foreign deposit.id
// JournalTypeMap -> Foreign map_profits.id
func InsertJournal(tx *gorm.DB, uid int32, b *mdbill.Balance,
	typ int32, fid string, opuid int32) error {
	now := gorm.NowFunc()

	m := make(map[string]interface{})
	m["amount"] = b.Amount
	m["gold"] = b.Gold
	m["diamond"] = b.Diamond

	m["user_id"] = uid
	m["type"] = typ
	m["`foreign`"] = fid
	m["created_at"] = now
	m["updated_at"] = now
	m["op_user_id"] = opuid

	// usql := "amount = amount + ? , gold = gold + ? , " +
	// 	"diamond = diamond + ? , updated_at = ?"
	// uargs := []interface{}{
	// 	b.Amount,
	// 	b.Gold,
	// 	b.Diamond,
	// 	now,
	// }

	// sql, args, _ := sq.Insert(enum.JournalTableName).SetMap(m).
	// 	Suffix("ON DUPLICATE KEY UPDATE "+usql, uargs...).ToSql()

	sql, args, _ := sq.Insert(enum.JournalTableName).SetMap(m).
		ToSql()

	if err := tx.Exec(sql, args...).Error; err != nil {
		return errors.Internal("save journal failed", err)
	}

	return nil
}

func GainBalance(tx *gorm.DB, uid int32, b *mdbill.Balance, typ int32,
	fid string, opuid int32) error {
	if b.Amount != 0 {
		return errbill.ErrNotAllowAmount
	}

	ub, err := GetLockUserBalance(tx, uid)
	if err != nil {
		return err
	}

	if ub.Gold+b.Gold < 0 || ub.Diamond+b.Diamond < 0 {
		return errbill.ErrOutOfBalance
	}

	ub.Gold += b.Gold
	ub.Diamond += b.Diamond

	if b.Gold > 0 {
		ub.GoldProfit += b.Gold
	}

	if b.Diamond > 0 {
		ub.DiamondProfit += b.Diamond

	}

	if b.Diamond < 0 {
		ub.CumulativeConsumption += b.Diamond
	}

	if typ == enum.JournalTypeGive {
		ub.CumulativeGift += b.Diamond
	}
	if typ == enum.JournalTypeRecharge {
		ub.CumulativeRecharge += b.Diamond
	}

	if err = tx.Save(ub).Error; err != nil {
		return errors.Internal("gain balance failed", err)
	}

	return InsertJournal(tx, uid, b, typ, fid, opuid)
}
