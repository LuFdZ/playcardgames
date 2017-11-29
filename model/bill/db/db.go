package db

import (
	"playcards/model/bill/enum"
	errbill "playcards/model/bill/errors"
	mdbill "playcards/model/bill/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jinzhu/gorm"
	//"github.com/Masterminds/squirrel"
	"fmt"
)

func CreateBalance(tx *gorm.DB, uid int32, balance *mdbill.Balance) error {
	var err error

	b := &mdbill.UserBalance{
		UserID:  uid,
		Balance: *balance,
	}
	fmt.Printf("CreateBalance:%v\n",b)

	if err = tx.Create(b).Error; err != nil {
		return errors.Internal("create balance", err)
	}

	err = InsertJournalByBalance(tx, uid, balance,&mdbill.UserBalance{UserID:  uid}, enum.JournalTypeInitBalance,int64(b.UserID), b.UserID, enum.DefaultChannel)
	if err != nil {
		return err
	}

	//err = InsertJournal(tx, uid, balance, enum.JournalTypeInitBalance,
	//	int64(b.UserID), b.UserID, enum.DefaultChannel)

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

func GetJournal(tx *gorm.DB, uid int32, orderid int64) int32 {
	out := &mdbill.Journal{}
	if found, _ := db.FoundRecord(tx.Where("user_id = ? and `foreign` = ? ",
		uid, orderid).Find(&out).Error); found {
		return 2
	}
	return 1
}

//func Deposit(tx *gorm.DB, uid int32, amount int64, typ int32) error {
//	if amount == 0 {
//		return nil
//	}
//
//	if amount < 0 {
//		return errbill.ErrInvalidParameter
//	}
//
//	d := &mdbill.Deposit{
//		UserID: uid,
//		Amount: amount,
//		Type:   typ,
//	}
//
//	err := tx.Create(d).Error
//	if err != nil {
//		return errors.Internal("deposit failed", err)
//	}
//
//	b, err := GetLockUserBalance(tx, uid)
//	if err != nil {
//		return errors.Internal("deposit failed", err)
//	}
//
//	b.Amount += amount
//	//b.Deposit += amount
//	err = tx.Save(b).Error
//	if err != nil {
//		return errors.Internal("deposit failed", err)
//	}
//
//	bal := &mdbill.Balance{
//		Amount: amount,
//	}
//	return InsertJournal(tx, b.UserID, bal, typ, int64(d.ID),
//		enum.SystemOpUserID, enum.DefaultChannel)
//}

// Type:
// JournalTypeInitBalance -> Foreign deposit.id
// JournalTypeDeposit -> Foreign deposit.id
// JournalTypeMap -> Foreign map_profits.id
func InsertJournal(tx *gorm.DB, uid int32, amounttype int32,amount int64,amountbefore int64,amountafter int64,
	typ int32, fid int64, opuid int32, channel string) error {
	now := gorm.NowFunc()

	m := make(map[string]interface{})
	m["amount_type"] = amounttype
	m["amount"] = amount
	m["amount_before"] = amountbefore
	m["amount_after"] = amountafter
	m["user_id"] = uid
	m["type"] = typ
	m["`foreign`"] = fid
	m["created_at"] = now
	m["updated_at"] = now
	m["op_user_id"] = opuid
	m["channel"] = channel

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
	fid int64, opuid int32, channel string) error {
	//if b.Amount != 0 {
	//	return errbill.ErrNotAllowAmount
	//}

	ub, err := GetLockUserBalance(tx, uid)
	if err != nil {
		return err
	}

	if ub.Gold+b.Gold < 0 || ub.Diamond+b.Diamond < 0 {
		return errbill.ErrOutOfBalance
	}

	//ub.Gold += b.Gold
	//ub.Diamond += b.Diamond
	//
	//if b.Diamond < 0 {
	//	ub.CumulativeConsumption += b.Diamond
	//}
	//
	//if typ == enum.JournalTypeGive {
	//	ub.CumulativeGift += b.Diamond
	//}
	//if typ == enum.JournalTypeRecharge {
	//	ub.CumulativeRecharge += b.Diamond
	//}
	err = InsertJournalByBalance(tx, uid, b, ub,typ, fid, opuid, channel)
	if err != nil {
		return err
	}
	if err = tx.Save(ub).Error; err != nil {
		return errors.Internal("gain balance failed", err)
	}
	return nil
	//return InsertJournal(tx, uid, b, typ, fid, opuid, channel)
}

func InsertJournalByBalance(tx *gorm.DB, uid int32, b *mdbill.Balance,ub *mdbill.UserBalance, typ int32,
	fid int64, opuid int32, channel string) error{
	if b.Gold != 0 {
		amountType := enum.TypeGold
		amountBefore := ub.Gold
		ub.Gold += b.Gold
		amountAfter := ub.Gold
		err := InsertJournal(tx, uid, int32(amountType),b.Gold,amountBefore,amountAfter, typ, fid, opuid, channel)
		if err != nil{
			return err
		}
	}

	if b.Diamond != 0 {
		amountType := enum.TypeDiamond
		amountBefore := ub.Diamond
		ub.Diamond += b.Diamond
		amountAfter := ub.Diamond
		err := InsertJournal(tx, uid, int32(amountType),b.Diamond,amountBefore,amountAfter, typ, fid, opuid, channel)
		if err != nil{
			return err
		}
	}
	return nil
}

//func GetUserConsumption(tx *gorm.DB,uid int32) int64 {
//	diamond := tx.Table(enum.JournalTableName).Select("sum(diamond) as total").
//	Where("user_id = ? and diamond <0 ,",uid).Value
//	return int64(diamond)
//}
