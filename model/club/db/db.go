package db

import (
	enumclub "playcards/model/club/enum"
	mdclub "playcards/model/club/mod"
	mdpage "playcards/model/page"
	"playcards/utils/errors"
	"playcards/utils/db"
	sq "github.com/Masterminds/squirrel"
	"github.com/jinzhu/gorm"
)

func CreateClub(tx *gorm.DB, mclub *mdclub.Club) error {
	now := gorm.NowFunc()
	mclub.CreatedAt = &now
	mclub.UpdatedAt = &now
	if err := tx.Create(mclub).Error; err != nil {
		return errors.Internal("create club failed", err)
	}
	return nil
}

func UpdateClub(tx *gorm.DB, mclub *mdclub.Club) (*mdclub.Club, error) {
	now := gorm.NowFunc()
	mclub.UpdatedAt = &now
	if err := tx.Model(mclub).Updates(mclub).Find(&mclub).Error; err != nil {
		return nil, errors.Internal("update club failed", err)
	}
	return mclub, nil
}

func GetLockClub(tx *gorm.DB, cid int32) (*mdclub.Club, error) {
	out := &mdclub.Club{}
	if err := db.ForUpdate(tx).Where("club_id = ? ", cid).Find(out).
		Error; err != nil {
		return nil, errors.Internal("get lock club failed", err)
	}
	return out, nil
}

func ClubBalance(tx *gorm.DB, mclub *mdclub.Club, amounttype int32, amount int64, typeid int32, fid int64, opid int64) (*mdclub.Club, error) {
	now := gorm.NowFunc()
	mclub.UpdatedAt = &now
	//tx *gorm.DB, clubid int32, amounttype int32,amount int64, mclub *mdclub.Club, typ int32, fid int64, opuid int32
	err := InsertJournalByBalance(tx, mclub.ClubID, amounttype, amount, mclub, typeid, fid, opid)
	if err != nil {
		return nil, err
	}

	mCl := mdclub.Club{
		Diamond: mclub.Diamond,
		Gold:    mclub.Gold,
	}
	if err := tx.Model(mclub).Update(mCl).
		Find(&mclub).Error; err != nil {
		return nil, errors.Internal("update club failed", err)
	}
	return mclub, nil
}

func CreateClubMember(tx *gorm.DB, mcm *mdclub.ClubMember) error {
	now := gorm.NowFunc()
	mcm.CreatedAt = &now
	mcm.UpdatedAt = &now
	if err := tx.Create(mcm).Error; err != nil {
		return errors.Internal("create club member failed", err)
	}
	return nil
}

func UpdateClubMember(tx *gorm.DB, mcm *mdclub.ClubMember) (*mdclub.ClubMember, error) {
	now := gorm.NowFunc()
	mMember := &mdclub.ClubMember{
		Status:    mcm.Status,
		UpdatedAt: &now,
	}
	if err := tx.Model(mMember).Updates(mcm).Find(&mcm).Error; err != nil {
		return nil, errors.Internal("update club member failed", err)
	}
	return mcm, nil
}

func PageClub(tx *gorm.DB, page *mdpage.PageOption,
	mclub *mdclub.Club) ([]*mdclub.Club, int64, error) {
	var out []*mdclub.Club
	rows, rtx := page.Find(tx.Model(mclub).Order("created_at desc").
		Where(mclub), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page club failed", rtx.Error)
	}
	return out, rows, nil
}

func PageClubMember(tx *gorm.DB, page *mdpage.PageOption,
	mcm *mdclub.ClubMember) ([]*mdclub.ClubMember, int64, error) {
	var out []*mdclub.ClubMember
	rows, rtx := page.Find(tx.Model(mcm).Order("created_at desc").
		Where(mcm), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page club member failed", rtx.Error)
	}
	return out, rows, nil
}

func GetAllAlineClubList(tx *gorm.DB) ([]*mdclub.Club, error) {
	var (
		out []*mdclub.Club
	)
	if err := tx.Where("status = ?", enumclub.ClubStatusNormal).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("select club list failed", err)
	}
	return out, nil
}

func GetAllAlineClubMemberList(tx *gorm.DB) ([]*mdclub.ClubMember, error) {
	var (
		out []*mdclub.ClubMember
	)
	if err := tx.Where("status = ?", enumclub.ClubMemberStatusNon).Order("club_id").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("select club member list failed", err)
	}
	return out, nil
}

func InsertJournalByBalance(tx *gorm.DB, clubid int32, amounttype int32, amount int64, mclub *mdclub.Club, typ int32,
	fid int64, opuid int64) error {
	if amounttype == enumclub.TypeGold && amount != 0 {
		amountBefore := mclub.Gold
		mclub.Gold += amount
		amountAfter := mclub.Gold
		err := InsertJournal(tx, clubid, int32(amounttype), amount, amountBefore, amountAfter, typ, fid, opuid)
		if err != nil {
			return err
		}
	}
	if amounttype == enumclub.TypeDiamond && amount != 0 {
		amountBefore := mclub.Diamond
		mclub.Diamond += amount
		amountAfter := mclub.Diamond
		err := InsertJournal(tx, clubid, int32(amounttype), amount, amountBefore, amountAfter, typ, fid, opuid)
		if err != nil {
			return err
		}
	}
	return nil
}

//func CreateClubJournal(tx *gorm.DB, mcj *mdclub.ClubJournal) error {
//	now := gorm.NowFunc()
//	mcj.CreatedAt = &now
//	mcj.UpdatedAt = &now
//	if err := tx.Create(mcj).Error; err != nil {
//		return errors.Internal("create club journal failed", err)
//	}
//	return nil
//}

func InsertJournal(tx *gorm.DB, clubid int32, amounttype int32, amount int64, amountbefore int64, amountafter int64,
	typ int32, fid int64, opuid int64) error {
	now := gorm.NowFunc()

	m := make(map[string]interface{})
	m["amount_type"] = amounttype
	m["amount"] = amount
	m["amount_before"] = amountbefore
	m["amount_after"] = amountafter
	m["club_id"] = clubid
	m["type"] = typ
	m["`foreign`"] = fid
	m["created_at"] = now
	m["updated_at"] = now
	m["op_user_id"] = opuid

	sql, args, _ := sq.Insert(enumclub.ClubJournalTableName).SetMap(m).
		ToSql()

	if err := tx.Exec(sql, args...).Error; err != nil {
		return errors.Internal("save journal failed", err)
	}

	return nil
}
