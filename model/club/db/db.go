package db

import (
	enumclub "playcards/model/club/enum"
	mdclub "playcards/model/club/mod"
	errorclub "playcards/model/club/errors"
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
	//mclub.UpdatedAt = &now
	mClub := &mdclub.Club{
		Icon:       mclub.Icon,
		ClubRemark: mclub.ClubRemark,
		Status:     mclub.Status,
		ClubParam:  mclub.ClubParam,
		Notice:     mclub.Notice,
		UpdatedAt:  &now,
	}
	if len(mclub.SettingParam) > 0 {
		mClub.SettingParam = mclub.SettingParam
	}
	if err := tx.Model(mclub).Updates(mClub).Error; err != nil { //.Find(&mclub)
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

func GetLockClubMember(tx *gorm.DB, uid int32) (*mdclub.ClubMember, error) {
	out := &mdclub.ClubMember{}
	if err := db.ForUpdate(tx).Where("user_id = ? and status < ?", uid, enumclub.ClubMemberStatusBan).Find(out).
		Error; err != nil {
		return nil, errors.Internal("get lock club member failed", err)
	}
	return out, nil
}

func ClubBalance(tx *gorm.DB, cid int32, amounttype int32, amount int64, typeid int32, fid int64, opid int64) (*mdclub.Club, error) { //mclub *mdclub.Club
	mclub, err := GetLockClub(tx, cid)
	if err != nil {
		return nil, err
	}
	if mclub == nil {
		return nil, errorclub.ErrClubNotExisted
	}
	now := gorm.NowFunc()
	//mclub.UpdatedAt = &now
	//ub, err := GetLockUserBalance(tx, uid, b.CoinType)
	//if err != nil {
	//	return nil,err
	//}

	//tx *gorm.DB, clubid int32, amounttype int32,amount int64, mclub *mdclub.Club, typ int32, fid int64, opuid int32
	err = InsertJournalByBalance(tx, mclub.ClubID, amounttype, amount, mclub, typeid, fid, opid)
	if err != nil {
		return nil, err
	}

	mCl := mdclub.Club{
		Diamond:   mclub.Diamond,
		Gold:      mclub.Gold,
		ClubCoin:  mclub.ClubCoin,
		UpdatedAt: &now,
	}
	if err := tx.Model(mclub).UpdateColumns(mCl).
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

func PageClubMemberRank(tx *gorm.DB, page *mdpage.PageOption, clubid int32) ([]*mdclub.ClubMember, int64, error) {
	var out []*mdclub.ClubMember
	rows, rtx := page.Find(tx.Where("club_id = ? and status = ?", clubid, enumclub.ClubMemberStatusNon).
		Order("club_coin desc").Find(&out), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page club member rank failed", rtx.Error)
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
	} else if amounttype == enumclub.TypeDiamond && amount != 0 {
		amountBefore := mclub.Diamond
		mclub.Diamond += amount
		amountAfter := mclub.Diamond
		err := InsertJournal(tx, clubid, int32(amounttype), amount, amountBefore, amountAfter, typ, fid, opuid)
		if err != nil {
			return err
		}
	} else if amounttype == enumclub.TypeClubCoin && amount != 0 {
		amountBefore := mclub.ClubCoin
		mclub.ClubCoin += amount
		amountAfter := mclub.ClubCoin
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
	m["status"] = enumclub.JournalStatusInit

	sql, args, _ := sq.Insert(enumclub.ClubJournalTableName).SetMap(m).
		ToSql()

	if err := tx.Exec(sql, args...).Error; err != nil {
		return errors.Internal("save journal failed", err)
	}

	return nil
}

func UpdateJournal(tx *gorm.DB, cjid int32, clubid int32) (*mdclub.ClubJournal, error) {
	now := gorm.NowFunc()
	mMj := &mdclub.ClubJournal{
		ID:        cjid,
		ClubID:    clubid,
		Status:    enumclub.MailClubStatusSure,
		UpdatedAt: &now,
	}
	if err := tx.Model(mMj).Updates(mMj).Find(&mMj).Error; err != nil {
		return nil, errors.Internal("update club journal failed", err)
	}
	return mMj, nil
}

func PageClubJournal(tx *gorm.DB, page *mdpage.PageOption, clubid int32, status int32) ([]*mdclub.ClubJournal, int64, error) {
	var (
		out   []*mdclub.ClubJournal
		mclub *mdclub.ClubJournal
	)
	// club_id = ?
	sqltx := tx.Model(mclub).Where("club_id = ? and amount_type = ? and type != ?", clubid, enumclub.TypeClubCoin, enumclub.JournalTypeClubGame)
	if status > 0 {
		sqltx = tx.Model(mclub).Where("club_id = ? and status = ? and amount_type = ? and type != ?", clubid, status, enumclub.TypeClubCoin, enumclub.JournalTypeClubGame)
	}

	rows, rtx := page.Find(sqltx.Order("created_at desc").Find(&out), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page club journal failed", rtx.Error)
	}
	return out, rows, nil
}

func PageClubMemberJournal(tx *gorm.DB, page *mdpage.PageOption, uid int32, clubid int32) ([]*mdclub.ClubJournal,
	int64, error) {
	var (
		out []*mdclub.ClubJournal
		//mclub *mdclub.ClubJournal
	)
	//mclub.OpUserID = uid
	//SELECT count(0) as `size` FROM `club_journals`  WHERE amount_type = '3' and (op_user_id = '100001' or (`foreign` = '100001' and type =1 )) ORDER BY created_at desc;
	rows, rtx := page.Find(tx.Where("club_id = ? and amount_type = ? and (op_user_id = ? or (`foreign` = ? and type = ?))",
		clubid, enumclub.TypeClubCoin, uid, uid, enumclub.JournalTypeClubAddMemberClubCoin).
		Order("created_at desc").Find(&out), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page club journal failed", rtx.Error)
	}
	return out, rows, nil
}

func GainClubMemberAndClubBalance(tx *gorm.DB, clubid int32, uid int32, amount int64,
	typeid int32) (*mdclub.Club, *mdclub.ClubMember, error) {
	mclub, err := GetLockClub(tx, clubid)
	if err != nil {
		return nil, nil, err
	}
	mclubMember, err := GetLockClubMember(tx, uid)
	if err != nil {
		return nil, nil, err
	}
	//now := gorm.NowFunc()
	if (typeid == enumclub.JournalTypeClubAddMemberClubCoin && mclub.ClubCoin < amount) ||
		(typeid == enumclub.JournalTypeClubMemberClubCoinOfferUp && mclubMember.ClubCoin < amount) {
		return nil, nil, errorclub.ErrClubCoinNotEnough
	}

	fid := clubid
	opid := uid
	if typeid == enumclub.JournalTypeClubAddMemberClubCoin {
		amount = -amount
		fid = uid
		opid = clubid
	}

	err = InsertJournalByBalance(tx, mclub.ClubID, enumclub.TypeClubCoin, amount, mclub, typeid, int64(fid), int64(opid))
	if err != nil {
		return nil, nil, err
	}
	//mCl := mdclub.Club{
	//	ClubCoin:  mclub.ClubCoin,
	//	UpdatedAt: &now,
	//}
	if err := tx.Model(mclub).Update("club_coin", mclub.ClubCoin).
		Find(&mclub).Error; err != nil {
		return nil, nil, errors.Internal("update club failed", err)
	}
	//mCm := mdclub.ClubMember{
	//	MemberId:  mclubMember.UserID,
	//	ClubCoin:  mclubMember.ClubCoin - amount,
	//	UpdatedAt: &now,
	//}
	//fmt.Printf("GainClubMemberAndClubBalance:%+v\n", mCm)
	//mCm.ClubCoin = 0
	if err := tx.Model(mclubMember).Update("club_coin", mclubMember.ClubCoin-amount).
		Find(&mclubMember).Error; err != nil {
		return nil, nil, errors.Internal("update club failed", err)
	}
	return mclub, mclubMember, nil
}

func GainClubMemberGameBalance(tx *gorm.DB, uid int32, amount int64, fid int64, opid int64, gameCost bool) (*mdclub.ClubMember, error) {
	mclubMember, err := GetLockClubMember(tx, uid)
	if err != nil {
		return nil, err
	}
	amountBefore := mclubMember.ClubCoin
	mclubMember.ClubCoin += amount
	amountAfter := mclubMember.ClubCoin

	var out mdclub.ClubJournal
	//out.Type = enumclub.JournalTypeClubGame
	//out.Foreign = fid
	//out.OpUserID = opid
	found, err := db.FoundRecord(tx.Where("type = ? and `foreign` = ? and op_user_id = ? ",
		enumclub.JournalTypeClubGame, fid, opid).Find(&out).Error)
	if err != nil {
		return nil, errors.Internal("get four card by id failed", err)
	}
	if !found || gameCost {
		jtype := enumclub.JournalTypeClubGame
		if gameCost {
			jtype = enumclub.JournalTypeClubGameCostBack
		}
		err = InsertJournal(tx, mclubMember.ClubID, enumclub.TypeClubCoin, amount, amountBefore, amountAfter, jtype, fid, opid)
		if err != nil {
			return nil, err
		}
	} else {
		out.AmountBefore = amountBefore
		out.AmountAfter = amountAfter
		out.Amount += amount
		if err := tx.Model(out).Updates(out).Find(&out).Error; err != nil {
			return nil, errors.Internal("update club journal failed", err)
		}
	}
	if gameCost {
		mclub, err := GetLockClub(tx, mclubMember.ClubID)
		if err != nil {
			return nil, err
		}
		clubAmountBefore := mclub.ClubCoin
		mclub.ClubCoin = mclub.ClubCoin - amount
		clubAmountAfter := mclub.ClubCoin
		err = InsertJournal(tx, mclubMember.ClubID, enumclub.TypeClubCoin, amount, clubAmountBefore, clubAmountAfter, enumclub.JournalTypeClubGameCostBack, fid, mclub.ClubCoin)
		if err != nil {
			return nil, err
		}
		//fmt.Printf("GainClubMemberGameBalance:%f|%f|%f", clubAmountBefore, mclub.ClubCoin, clubAmountAfter)
		if err := tx.Model(mclub).Update("club_coin", mclub.ClubCoin).
			Find(&mclub).Error; err != nil {
			return nil, errors.Internal("update club coin cost back failed", err)
		}
	}
	if err := tx.Model(mclubMember).Update("club_coin", mclubMember.ClubCoin).
		Find(&mclubMember).Error; err != nil {
		return nil, errors.Internal("update club member coin failed", err)
	}
	return mclubMember, nil
}
