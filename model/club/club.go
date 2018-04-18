package club

import (
	cacheclub "playcards/model/club/cache"
	dbclub "playcards/model/club/db"
	enumclub "playcards/model/club/enum"
	errclub "playcards/model/club/errors"
	mdclub "playcards/model/club/mod"
	"playcards/model/common"
	"playcards/model/bill"
	cachecon "playcards/model/common/cache"
	dbcon "playcards/model/common/db"
	enumcon "playcards/model/common/enum"
	errcon "playcards/model/common/errors"
	mdcon "playcards/model/common/mod"
	mdpage "playcards/model/page"
	cacheroom "playcards/model/room/cache"
	dbroom "playcards/model/room/db"
	enumroom "playcards/model/room/enum"
	enumbill "playcards/model/bill/enum"
	mdroom "playcards/model/room/mod"
	mdbill "playcards/model/bill/mod"
	"playcards/model/user"
	cacheuser "playcards/model/user/cache"
	dbuser "playcards/model/user/db"
	mduser "playcards/model/user/mod"
	pbclub "playcards/proto/club"
	pbroom "playcards/proto/room"
	"playcards/utils/errors"
	"encoding/base64"
	"playcards/utils/db"
	"encoding/json"
	"playcards/utils/log"
	utilpb "playcards/utils/proto"
	utilproto "playcards/utils/proto"

	"github.com/jinzhu/gorm"
	"fmt"
)

func CreateClub(name string, creatorid int32, creatorproxy int32) error {
	checklen := len(name)
	if checklen < 3 || checklen > 60 {
		return errclub.ErrClubNameLen
	}
	if creatorid == 0 || creatorid < 100000 {
		return errclub.ErrCreatorid
	}
	//if creatorproxy == 0 {
	//	return errclub.ErrCreatorid
	//}
	//mdcs, err := GetClubsByMemberID(mdu.UserID)
	//if err != nil {
	//	return err
	//}
	count, err := dbclub.GetClubByCreatorID(db.DB(), creatorid)
	if err != nil {
		return err
	}
	if (creatorid > 0 && count >= enumclub.ProxyCreateClubLimit) || (creatorproxy == 0 && count >= enumclub.PlayerCreateClubLimit) {

		return errclub.ErrCreateClubLimit
	}

	mclub := &mdclub.Club{
		ClubName:     name,
		Status:       enumclub.ClubStatusNormal,
		CreatorID:    creatorid,
		CreatorProxy: creatorproxy,
		Setting:      &mdclub.SettingParam{1, 0, 0, 1, 0, 1},
		ClubCoin:     enumclub.ClubCoinInit,
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		err := dbclub.CreateClub(tx, mclub)
		if err != nil {
			return err
		}
		err = cacheclub.SetClub(mclub)
		if err != nil {
			return nil
		}
		return nil
	})
	if err != nil {
		return err
	}
	CreateClubMember(mclub.ClubID, creatorid)
	return nil
}

func GetClubFromDB(cid int32) (*mdclub.Club, error) {
	return dbclub.GetLockClub(db.DB(), cid)
}

func GetClubMember(clubid int32, uid int32) (*mdclub.ClubMember, error) {
	return cacheclub.GetClubMember(clubid, uid)
}

func SetClubBalance(amonut int64, amonuttype int32, clubid int32, typeid int32, foreign int64, opid int64, adminOp bool) error {
	//mClub, err := cacheclub.GetClub(clubid)
	//if err != nil {
	//	return err
	//}

	var mClub *mdclub.Club
	err := db.Transaction(func(tx *gorm.DB) error {
		if !adminOp {
			err := bill.GainGameBalance(int32(opid), clubid, enumbill.JournalTypeClubRecharge,
				&mdbill.Balance{Amount: int64(-amonut), CoinType: enumbill.TypeDiamond})
			if err != nil {
				return err
			}
		}
		c, err := dbclub.ClubBalance(tx, clubid, amonuttype, amonut, typeid, foreign, opid)
		if err != nil {
			return err
		}
		mClub = c
		return nil
	})
	if err != nil {
		return err
	}
	err = cacheclub.SetClub(mClub)
	if err != nil {
		return nil
	}
	return nil
}

func SetClubGameBalance(amonut int64, amonuttype int32, clubid int32, typeid int32, foreign int64, opid int64) error {
	//mClub, err := cacheclub.GetClub(clubid)
	//if err != nil {
	//	return err
	//}

	var mClub *mdclub.Club
	err := db.Transaction(func(tx *gorm.DB) error {
		c, err := dbclub.ClubBalance(tx, clubid, amonuttype, amonut, typeid, foreign, opid)
		if err != nil {
			return err
		}
		mClub = c
		return nil
	})
	if err != nil {
		return err
	}
	err = cacheclub.SetClub(mClub)
	if err != nil {
		return nil
	}
	return nil
}

func SetClubRoomFlag(clubid int32, rid int32) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := dbroom.SetRoomFlage(tx, clubid, rid)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return nil
}

func UpdateClub(mclub *mdclub.Club) error {
	if !cacheclub.CheckClubExists(mclub.ClubID) {
		return errclub.ErrClubNotExisted
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		mclub, err := dbclub.UpdateClub(tx, mclub)
		if err != nil {
			return err
		}
		mclub, err = dbclub.GetLockClub(tx, mclub.ClubID)
		if err != nil {
			return err
		}
		err = cacheclub.SetClub(mclub)
		if err != nil {
			return nil
		}
		return nil
	})
	if err != nil {
		return nil
	}

	return nil
}

func DeleteClub(clubID int32) ([]int32, []int32, error) {
	mdClub, err := cacheclub.GetClub(clubID)
	if err != nil {
		return nil, nil, err
	}
	mdClub.Status = enumclub.ClubStatusDel
	mcms := cacheclub.GetAllClubMember(clubID, false)
	mdrs, err := cacheroom.GetAllClubRoom(clubID)
	if err != nil {
		return nil, nil, err
	}
	if len(mdrs) > 0 {
		return nil, nil, errclub.ErrClubHasGame
	}
	var ids []int32
	var onLineIds []int32
	for _, mdMember := range mcms {
		if mdMember.ClubCoin != 0 {
			return nil, nil, errclub.ErrMemberClubCoinNonZero
		}
		ids = append(ids, mdMember.UserID)
	}
	mcmsOnLine := cacheclub.GetAllClubMember(clubID, true)
	for _, mcmsOnLine := range mcmsOnLine {
		onLineIds = append(onLineIds, mcmsOnLine.UserID)
	}
	for _, mdMember := range mcms {
		err := RemoveClubMember(clubID, mdMember.UserID, enumclub.ClubMemberStatusDissolution)
		if err != nil {
			log.Err("delete club member clubid:%d,userid:%d,err:%v", clubID, mdMember.UserID, err)
		}
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		_, err := dbclub.UpdateClub(tx, mdClub)
		if err != nil {
			return err
		}
		err = cacheclub.DeleteClub(mdClub.ClubID)
		if err != nil {
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return ids, onLineIds, nil
}

func UpdateClubMemberStatus(clubid int32, uid int32, status int32) error {
	mcm, err := cacheclub.GetClubMember(clubid, uid)
	if err != nil {
		return err
	}
	if mcm == nil {
		return errclub.ErrNotInClub
	}
	if status != enumclub.ClubMemberStatusNon && status != enumclub.ClubMemberStatusBan {
		return errclub.ErrStatus
	}
	if status == mcm.Status {
		return nil
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		mcm.Status = status
		_, err := dbclub.UpdateClubMember(tx, mcm)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil
	}
	err = cacheclub.SetClubMember(mcm)
	if err != nil {
		return nil
	}
	return nil
}

func RemoveClubMember(clubid int32, uid int32, removeType int32) error {
	mcm, err := cacheclub.GetClubMember(clubid, uid)
	if err != nil {
		return err
	}
	if mcm == nil {
		return errclub.ErrNotInClub
	}
	_, muser := cacheuser.GetUserByID(uid)
	if muser == nil {
		return errcon.ErrUserErr
	}
	if mcm.ClubCoin != 0 {
		return errclub.ErrClubCoinNegative
	}

	userInRoom := cacheroom.ExistRoomUser(uid)
	if userInRoom {
		return errclub.ErrMemberInRoom
	}

	mcm.Status = removeType
	muser.ClubID = 0
	err = db.Transaction(func(tx *gorm.DB) error {
		_, err := dbclub.UpdateClubMember(tx, mcm)
		if err != nil {
			return err
		}
		err = dbuser.ReSetUserClubID(tx, muser)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil
	}
	err = cacheuser.SimpleUpdateUser(muser)
	if err != nil {
		return nil
	}
	err = cacheclub.DeleteClubMember(mcm.ClubID, mcm.UserID)
	if err != nil {
		return nil
	}
	return nil
}

func CreateClubMember(clubid int32, uid int32) (*mdclub.Club, *mduser.User, error) { //muser *mduser.User
	_, muser := cacheuser.GetUserByID(uid)
	if muser == nil {
		return nil, nil, errcon.ErrUserErr
	}
	mClub, err := CheckUserJoinClubRoom(clubid, muser)
	if err != nil {
		return nil, nil, err
	}
	mcm := &mdclub.ClubMember{
		UserID: muser.UserID,
		ClubID: clubid,
		Status: enumclub.ClubStatusNormal,
		Role:   enumclub.ClubMember,
	}
	me, _ := cachecon.GetExamine(enumcon.TypeClub, mClub.ClubID, muser.UserID)
	muser.ClubID = mClub.ClubID
	err = db.Transaction(func(tx *gorm.DB) error {
		_, err := dbuser.UpdateUser(tx, muser)
		if err != nil {
			return err
		}
		err = dbclub.CreateClubMember(tx, mcm)
		if err != nil {
			return err
		}
		if me != nil {
			me.Status = enumcon.ExamineStatusPass
			_, err := dbcon.UpdateExamine(tx, me)
			if err != nil {
				return err
			}
			cachecon.DeleteExamine(me)
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil
	}

	err = cacheuser.SimpleUpdateUser(muser)
	if err != nil {
		return nil, nil, nil
	}
	err = cacheclub.SetClubMember(mcm)
	if err != nil {
		return nil, nil, nil
	}
	return mClub, muser, nil
}

func CheckUserJoinClubRoom(clubid int32, muser *mduser.User) (*mdclub.Club, error) {
	//if muser.ClubID > 0 {
	//	return nil, errclub.ErrAlreadyInClub
	//}

	mClub, err := cacheclub.GetClub(clubid)
	if err != nil {
		return nil, err
	}

	if mClub == nil {
		return nil, errclub.ErrClubNotExisted
	}

	if mClub.Status != enumclub.ClubStatusNormal {
		return nil, errclub.ErrStatusNoINNormal
	}

	mcm, err := cacheclub.GetClubMember(clubid, muser.UserID)
	if mcm != nil {
		return nil, errclub.ErrExistedInClub
	}

	mbl, err := cachecon.GetBlackList(enumcon.TypeClub, clubid, muser.UserID)
	if mbl != nil {
		return nil, errclub.ErrInBlackList
	}

	count, err := cacheclub.CountClubMemberHKeys(clubid)
	if count >= enumclub.ClubMemberLimit {
		return nil, errclub.ErrClubMemberLimit
	}
	return mClub, nil
}

func JoinClub(clubid int32, muser *mduser.User) error {
	me, err := cachecon.GetExamine(enumcon.TypeClub, clubid, muser.UserID)
	if err != nil {
		return err
	}
	if me != nil {
		return errcon.ErrExisted
	}

	mClub, err := CheckUserJoinClubRoom(clubid, muser)
	if err != nil {
		return err
	}

	me = &mdcon.Examine{
		Type:        enumcon.TypeClub,
		ApplicantID: muser.UserID,
		AuditorID:   clubid,
		Status:      enumcon.ExamineStatusNew,
		OpID:        mClub.CreatorID,
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		//me.Status = enumcon.ExamineStatusPass
		err := dbcon.CreateExamine(tx, me)
		if err != nil {
			return err
		}
		err = cachecon.SetExamine(me)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func SetBlackList(clubid int32, uid int32, opid int32) error {
	err := RemoveClubMember(clubid, uid, enumclub.ClubMemberStatusBlackList)
	if err != nil {
		return err
	}
	err = common.CreateBlackList(enumcon.TypeClub, clubid, uid, opid)
	if err != nil {
		return err
	}
	return nil
}

func CancelBlackList(clubid int32, uid int32, opid int32) error {
	mdc := &mdcon.BlackList{
		OriginID: clubid,
		TargetID: uid,
		Type:     enumcon.TypeClub,
	}
	err := common.CancelBlackList(mdc, opid)
	if err != nil {
		return err
	}
	return nil
}

func PageBlackListMember(page *mdpage.PageOption, clubid int32) (
	[]*pbclub.ClubMember, int64, error) {
	page.Page -= 1
	mdbl := &mdcon.BlackList{
		OriginID: clubid,
	}
	mdms, rows, err := common.PageBlackList(page, mdbl)
	if err != nil {
		return nil, 0, err
	}
	var out []*pbclub.ClubMember
	for _, mdm := range mdms {
		m := &pbclub.ClubMember{
			ClubID: mdm.OriginID,
			UserID: mdm.TargetID,
		}
		_, u := cacheuser.GetUserByID(mdm.TargetID)
		if u != nil {
			m.Nickname = u.Nickname
		}
		out = append(out, m)
	}
	return out, rows, err
}

func PageClubExamineMember(page *mdpage.PageOption, clubid int32) (
	[]*pbclub.ClubMember, int64, error) {
	page.Page -= 1
	mdexm := &mdcon.Examine{
		AuditorID: clubid,
	}
	mdexms, rows, err := common.PageExamine(page, mdexm)
	if err != nil {
		return nil, 0, err
	}
	var out []*pbclub.ClubMember
	for _, mdexm := range mdexms {
		m := &pbclub.ClubMember{
			ClubID: mdexm.AuditorID,
			UserID: mdexm.ApplicantID,
		}
		_, u := cacheuser.GetUserByID(mdexm.ApplicantID)
		if u != nil {
			m.Nickname = u.Nickname
		}
		out = append(out, m)
	}
	return out, rows, err
}

func CreateClubExamine(clubid int32, uid int32, opid int32) error {
	mdexm := &mdcon.Examine{
		Type:        enumcon.TypeClub,
		AuditorID:   clubid,
		ApplicantID: uid,
		Status:      enumcon.ExamineStatusNew,
	}
	err := common.CreateExamine(mdexm, opid)
	if err != nil {
		return err
	}
	return nil
}

func UpdateClubExamine(clubid int32, uid int32, status int32, opid int32) (*mdclub.Club, error) {
	if status != enumcon.ExamineStatusPass && status != enumcon.ExamineStatusRefuse {
		return nil, errcon.ErrParam
	}
	err := common.UpdateExamine(enumcon.TypeClub, clubid, uid, status, opid)
	if err != nil {
		return nil, err
	}
	var mdc *mdclub.Club
	if status == enumcon.ExamineStatusPass {
		maClub, _, err := CreateClubMember(clubid, uid)
		if err != nil {
			return nil, err
		}
		mdc = maClub
	}

	return mdc, nil
}

func GetClub(clubid int32, uid int32, hasClubList bool) (*pbclub.ClubInfo, error) { //muser *mduser.User,, clubID int32
	//if muser.ClubID == 0 {
	//	return nil, errclub.ErrNotJoinAnyClub
	//}

	pbclubs, err := GetClubsByMemberID(uid)
	if err != nil {
		return nil, err
	}
	if len(pbclubs) == 0 {
		return nil, errclub.ErrNotJoinAnyClub
	}
	var cid int32 = 0
	for _, mdc := range pbclubs {
		if mdc.ClubID == clubid {
			cid = clubid
		}
	}
	if cid == 0 {
		cid = pbclubs[0].ClubID
	}

	//fmt.Printf("GetClub:%d|%d|%+v|%+v\n", clubid, cid, pbclubs[0], pbclubs)
	mClub, err := cacheclub.GetClub(cid)
	if err != nil {
		cid = pbclubs[0].ClubID
		//fmt.Printf("GetClub:%d|%d|%+v|%+v\n", clubid, cid, pbclubs[0], pbclubs)
		mClub, err = cacheclub.GetClub(cid)
		if err != nil {
			return nil, err
		}
		//return nil, err
	}

	if mClub == nil {
		return nil, errclub.ErrClubNotExisted
	}

	if mClub.Status != enumclub.ClubStatusNormal {
		return nil, errclub.ErrStatusNoINNormal
	}
	ci := &pbclub.ClubInfo{
		UserID: uid,
		Club:   mClub.ToProto(),
	}

	f := func(r *mdroom.Room) bool {
		if r.Status < enumroom.RoomStatusDelay &&
			r.ClubID == cid && r.VipRoomSettingID == 0 {
			return true
		}
		return false
	}
	rooms := cacheroom.GetAllRooms(f)
	utilpb.ProtoSlice(rooms, &ci.RoomList)

	mcms := cacheclub.GetAllClubMember(cid, false)
	utilpb.ProtoSlice(mcms, &ci.ClubMemberList)
	for _, mcm := range mcms {
		if mcm.UserID == uid {
			ci.Club.ClubCoin = mcm.ClubCoin
		}
	}
	if hasClubList {
		for _, pbc := range pbclubs {
			mcms := cacheclub.GetAllClubMember(pbc.ClubID, false)
			pbc.MemberCount = int32(len(mcms))
			var onlineCount int32
			for _, mcm := range mcms {
				if mcm.Online == 1 {
					//fmt.Printf("AAAA:%+v\n",mcm)
					pbc.MemberOnline += 1
					onlineCount++
				}
			}
		}
	}

	ci.ClubList = pbclubs
	return ci, nil
}

func GetClubInfo(clubID int32) (*mdclub.Club, error) { //muser *mduser.User
	//if muser.ClubID == 0 {
	//	return nil, errclub.ErrNotJoinAnyClub
	//}
	mClub, err := cacheclub.GetClub(clubID)
	if err != nil {
		return nil, err
	}

	if mClub == nil {
		return nil, errclub.ErrClubNotExisted
	}

	if mClub.Status != enumclub.ClubStatusNormal {
		return nil, errclub.ErrStatusNoINNormal
	}
	return mClub, nil
}

func PageClub(page *mdpage.PageOption, mclub *mdclub.Club) (
	[]*mdclub.Club, int64, error) {
	page.Page -= 1
	clubs, rows, err := dbclub.PageClub(db.DB(), page, mclub)
	if err == nil {
		for _, club := range clubs {
			number, err := cacheclub.CountClubMemberHKeys(club.ClubID)
			if err != nil {
				continue
			}
			club.MemberNumber = number
		}
	}
	return clubs, rows, err
}

func PageClubMember(page *mdpage.PageOption, mem *mdclub.ClubMember) (
	[]*pbclub.ClubMember, int64, error) {
	page.Page -= 1
	mdms, rows, err := dbclub.PageClubMember(db.DB(), page, mem)
	if err != nil {
		return nil, 0, err
	}
	var out []*pbclub.ClubMember
	for _, mdm := range mdms {
		m := &pbclub.ClubMember{
			ClubID:   mdm.ClubID,
			UserID:   mdm.UserID,
			Status:   mdm.Status,
			ClubCoin: mdm.ClubCoin,
		}
		_, u := cacheuser.GetUserByID(mdm.UserID)
		if u != nil {
			m.Nickname = u.Nickname
			m.Icon = u.Icon
		}
		out = append(out, m)
	}
	return out, rows, err
}

func PageClubRoom(clubid int32, page int32, pagesize int32, flag int32) (
	[]*pbroom.Room, error) {
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = 20
	}
	if flag == 0 {
		flag = 2
	}
	rooms, err := dbroom.PageRoomList(db.DB(), clubid, page, pagesize, flag)
	if err != nil {
		return nil, err
	}
	var out []*pbroom.Room
	for _, r := range rooms {
		var pbr *pbroom.Room
		r.UnmarshalGameUserResult()
		r.GameUserResult = ""
		pbr = r.ToProto()
		var rurs []*mdroom.GameUserResult
		for _, rur := range r.UserResults {
			simpleUserResult := &mdroom.GameUserResult{
				UserID: rur.UserID,
				Score:  rur.Score,
			}
			_, u := cacheuser.GetUserByID(rur.UserID)
			if u != nil {
				simpleUserResult.Nickname = u.Nickname
			}
			rurs = append(rurs, simpleUserResult)
		}
		//pbr.ResultList = rurs.
		err := utilproto.ProtoSlice(rurs, &pbr.ResultList)
		if err != nil {
			return nil, err
		}
		out = append(out, pbr)
	}
	return out, nil
}

func PageClubJournal(page *mdpage.PageOption, clubid int32, status int32) (
	[]*mdclub.ClubJournal, int64, error) {
	//page.Page -= 1
	return dbclub.PageClubJournal(db.DB(), page, clubid, status)
}

func PageClubMemberJournal(page *mdpage.PageOption, uid int32, clubid int32) (
	[]*pbclub.ClubJournal, int64, error) {
	//page.Page -= 1
	mdcjs, count, err := dbclub.PageClubMemberJournal(db.DB(), page, uid, clubid)
	if err != nil {
		return nil, 0, err
	}
	var out []*pbclub.ClubJournal
	for _, mdcj := range mdcjs {
		str := enumclub.JouranlTypeNameMap[mdcj.Type]
		typename := base64.StdEncoding.EncodeToString([]byte(str))
		pdcj := &pbclub.ClubJournal{
			Amount:    mdcj.Amount,
			Type:      mdcj.Type,
			CreatedAt: mdcj.CreatedAt.Unix(),
			TagName:   typename,
			Foreign:   mdcj.Foreign,
		}
		if mdcj.Type == enumclub.JournalTypeClubAddMemberClubCoin || mdcj.Type == enumclub.JournalTypeClubMemberClubCoinOfferUp {
			pdcj.Amount *= -1
		}
		out = append(out, pdcj)
	}
	return out, count, nil
}

func UpdateClubJournal(cjid int32, clubid int32) error {
	return db.Transaction(func(tx *gorm.DB) error {
		_, err := dbclub.UpdateJournal(db.DB(), cjid, clubid)
		if err != nil {
			return err
		}
		return nil
	})
}

func AddClubMemberClubCoin(clubid int32, uid int32, amonut int64, ) (int64, error) {
	var (
		mdClub *mdclub.Club
		mdCm   *mdclub.ClubMember
	)
	mdClub, err := cacheclub.GetClub(clubid)
	if err != nil {
		return 0, err
	}
	mdCm, err = cacheclub.GetClubMember(clubid, uid)
	if err != nil {
		return 0, err
	}
	//if mdClub.CreatorID != CreatorID {
	//	return 0, errclub.ErrNotCreatorID
	//}
	if mdClub.ClubCoin < amonut {
		return 0, errclub.ErrClubCoinNotEnough
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		c, m, err := dbclub.GainClubMemberAndClubBalance(tx, clubid, uid, amonut, enumclub.JournalTypeClubAddMemberClubCoin)
		if err != nil {
			return err
		}
		mdClub = c
		mdCm = m
		return nil
	})
	if err != nil {
		return 0, err
	}
	err = cacheclub.SetClub(mdClub)
	if err != nil {
		return 0, err
	}
	err = cacheclub.SetClubMember(mdCm)
	if err != nil {
		return 0, err
	}
	return mdCm.ClubCoin, nil
}

func ClubMemberOfferUpClubCoin(clubid int32, uid int32, amonut int64) (int64, error) {
	var (
		mdClub *mdclub.Club
		mdCm   *mdclub.ClubMember
	)
	mdCm, err := cacheclub.GetClubMember(clubid, uid)
	if err != nil {
		return 0, err
	}
	if mdCm.ClubCoin < amonut {
		return 0, errclub.ErrClubCoinNotEnough
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		c, m, err := dbclub.GainClubMemberAndClubBalance(tx, clubid, uid, amonut, enumclub.JournalTypeClubMemberClubCoinOfferUp)
		if err != nil {
			return err
		}
		mdClub = c
		mdCm = m
		return nil
	})
	if err != nil {
		return 0, nil
	}
	err = cacheclub.SetClub(mdClub)
	if err != nil {
		return 0, nil
	}
	err = cacheclub.SetClubMember(mdCm)
	if err != nil {
		return 0, nil
	}
	return mdCm.ClubCoin, nil
}

func GainClubMemberGameBalance(amonut int64, clubid int32, uid int32, fid int64, opid int64, gameCost bool) (*mdclub.ClubMember, error) {
	var mdCm *mdclub.ClubMember
	if amonut == 0 {
		return nil, nil
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		m, err := dbclub.GainClubMemberGameBalance(tx, clubid, uid, amonut, fid, opid, gameCost)
		if err != nil {
			return err
		}
		mdCm = m
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = cacheclub.SetClubMember(mdCm)
	if err != nil {
		return nil, err
	}
	return mdCm, nil
}

func GetClubMemberCoinRank(page *mdpage.PageOption, clubid int32) ([]*pbclub.ClubMember, int64, error) {
	var pbcms []*pbclub.ClubMember
	mdcms, count, err := dbclub.PageClubMemberRank(db.DB(), page, clubid)
	if err != nil {
		return nil, 0, err
	}
	for _, mdcm := range mdcms {
		_, mdu := cacheuser.GetUserByID(mdcm.UserID)
		if mdu == nil {
			continue
		}
		pbcm := &pbclub.ClubMember{
			MemberID: mdcm.UserID,
			Nickname: mdu.Nickname,
			ClubCoin: mdcm.ClubCoin,
			Online:   cacheuser.GetUserOnlineStatus(mdcm.UserID),
		}
		pbcms = append(pbcms, pbcm)
	}
	return pbcms, count, err
}

func RefreshAllFromDB() error {
	err := db.Transaction(func(tx *gorm.DB) error {
		mClubs, err := dbclub.GetAllAlineClubList(db.DB())
		if err != nil {
			return err
		}
		cacheclub.SetAllClub(mClubs)
		mCms, err := dbclub.GetAllAlineClubMemberList(db.DB())
		if err != nil {
			return err
		}
		cacheclub.SetAllClubMember(mCms)

		mVrs, err := dbclub.GetAllAlineVipRoomSetting(db.DB())
		if err != nil {
			return err
		}
		cacheclub.SetAllVipRoomSetting(mVrs)
		//for _,mc := range mClubs{
		//	CreatorInClub :=cacheclub.CheckClubMemberExist(mc.ClubID,mc.CreatorID)
		//	if !CreatorInClub {
		//		CreateClubMember(mc.ClubID,mc.CreatorID)
		//	}
		//}

		return nil

	})
	if err != nil {
		return nil
	}

	return nil
}

func UpdateClubProxyID(uid int32, proxyID int32) error {

	mdu := &mduser.User{
		UserID:  uid,
		ProxyID: proxyID,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		mdu, err := user.UpdateUser(mdu)
		if err != nil {
			return err
		}
		cacheuser.SimpleUpdateUser(mdu)
		mdcs, err := dbclub.UpdateClubProxyID(tx, uid, proxyID)
		if err != nil {
			return err
		}

		for _, mdclub := range mdcs {
			err = cacheclub.SetClub(mdclub)
			if err != nil {
				continue
			}
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return nil
}

func GetClubsByMemberID(uid int32) ([]*pbclub.Club, error) {
	var out []*pbclub.Club
	mdcs, err := cacheclub.GetClubsByMemberID(uid)
	if err != nil {
		return nil, err
	}
	for _, mdc := range mdcs {
		pbc := &pbclub.Club{
			ClubName:  mdc.ClubName,
			ClubID:    mdc.ClubID,
			CreatorID: mdc.CreatorID,
		}
		_, mdu := cacheuser.GetUserByID(mdc.CreatorID)
		if mdu != nil {
			pbc.CreatorName = mdu.Nickname
		}

		out = append(out, pbc)
	}
	return out, nil

}

func CreateVipRoomSetting(mdu *mduser.User, mvrs *mdclub.VipRoomSetting) error {
	mdClub, err := cacheclub.GetClub(mvrs.ClubID)
	if err != nil {
		return err
	}

	if mdClub.Status != enumclub.ClubStatusNormal {
		return errclub.ErrStatusNoINNormal
	}
	//if len(mvrs.Name) == 0 {
	//	return errclub.ErrNameLen
	//}
	mvrss, err := cacheclub.GetAllVipRoomSetting(mvrs.ClubID)
	if err != nil {
		return err
	}
	count := len(mvrss)

	if (mdu.ProxyID > 0 && count >= enumclub.ProxyCreateVipRoomSettingLimit) || (mdu.ProxyID == 0 &&
		count >= enumclub.PlayerCreateVipRoomSettingLimit) {
		return errclub.ErrCreateVipRoomLimit
	}
	mvrs.MaxNumber, mvrs.RoundNumber, mvrs.GameType, mvrs.GameParam, mvrs.RoomAdvanceOptions, err =
		checkGameParam(mvrs.MaxNumber, mvrs.RoundNumber, mvrs.GameType, mvrs.GameParam, mvrs.RoomAdvanceOptions)
	if err != nil {
		return err
	}
	if mvrs.RoomType == enumroom.RoomTypeClub && (mvrs.SubRoomType != -1 && mvrs.SubRoomType != enumroom.SubTypeClubMatch) {
		return errclub.ErrGameParam
	}
	if mvrs.SubRoomType == enumroom.SubTypeClubMatch {
		var setttingParam *mdroom.SettingParam
		if err := json.Unmarshal([]byte(mvrs.SettingParam), &setttingParam); err != nil {
			log.Err("club vip room check setting param unmarshal failed, %v", err)
			return errclub.ErrGameParam
		}
		if setttingParam.ClubCoinRate != 1 && setttingParam.ClubCoinRate != 2 && setttingParam.ClubCoinRate != 5 && setttingParam.ClubCoinRate != 10 {
			return errclub.ErrGameParam
		}
	}

	mvrs.Status = enumclub.VipRoomSettingNon
	err = db.Transaction(func(tx *gorm.DB) error {
		err := dbclub.CreateVipRoomSetting(tx, mvrs)
		if err != nil {
			return err
		}
		err = cacheclub.SetVipRoomSetting(mvrs)
		if err != nil {
			return nil
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

//func chekcGameParam(maxNumber int32, maxRound int32, gtype int32, gameParam string, RoomAdvanceOptions []string) error {
//	if len(gameParam) == 0 {
//		return errclub.ErrGameParam
//	}
//	if maxNumber < 2 {
//		return errclub.ErrRoomMaxNumber
//	}
//	//TODO 更新正式服前放开限制
//	if maxRound != 10 && maxRound != 20 && maxRound != 30 {
//		return errclub.ErrRoomMaxRound
//	}
//	//fmt.Printf("ChekcGameParam:%d|%d|%d|%s\n",maxNumber,maxRound,gtype,gameParam)
//	//if JoinType != enumroom.CanJoin && JoinType != enumroom.CanNotJoin {
//	//	mderr := errors.Parse(errclub.ErrGameParam.Error())
//	//	mderr.Detail = fmt.Sprintf(mderr.Detail, "加入类型格式错误！")
//	//	return mderr
//	//}
//	if len(RoomAdvanceOptions) == 0 {
//		mderr := errors.Parse(errclub.ErrGameParam.Error())
//		mderr.Detail = fmt.Sprintf(mderr.Detail, "房间参数格式错误！")
//		return mderr
//	}
//	if RoomAdvanceOptions[0] != "1" && RoomAdvanceOptions[0] != "0" {
//		mderr := errors.Parse(errclub.ErrGameParam.Error())
//		mderr.Detail = fmt.Sprintf(mderr.Detail, "房间允许加入参数格式错误！")
//		return mderr
//	}
//
//	switch gtype {
//	case enumroom.ThirteenGameType:
//		if maxNumber > 8 {
//			return errclub.ErrRoomMaxNumber
//		}
//		var roomParam *mdroom.ThirteenRoomParam
//		mderr := errors.Parse(errclub.ErrGameParam.Error())
//		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
//			log.Err("room check thirteen clean unmarshal room param failed, %v", err)
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
//			return mderr
//		}
//		//if roomParam.BankerType != 1 && roomParam.BankerType != 2 {
//		//	return errors.ErrGameParam
//		//}
//		if roomParam.BankerAddScore < 0 || roomParam.BankerAddScore > 6 || roomParam.BankerAddScore%2 != 0 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "当庄加分格式错误！")
//			return mderr
//		}
//		if roomParam.Joke != 0 && roomParam.Joke != 1 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "大小王格式错误！")
//			return mderr
//		}
//		if roomParam.Times < 1 || roomParam.Times > 3 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "比赛模式格式错误！")
//			return mderr
//		}
//		break
//	case enumroom.NiuniuGameType:
//		if maxNumber != 4 && maxNumber != 6 && maxNumber != 8 && maxNumber != 10 {
//			return errclub.ErrRoomMaxNumber
//		}
//		var roomParam *mdroom.NiuniuRoomParam
//		mderr := errors.Parse(errclub.ErrGameParam.Error())
//		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
//			log.Err("niuniu unmarshal room param failed, %v", err)
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
//			return mderr
//		}
//		if roomParam.BankerType < 1 || roomParam.BankerType > 5 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "玩法ID错误！")
//			return mderr
//		}
//		if roomParam.Times != 1 && roomParam.Times != 2 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "倍数ID错误！")
//			return mderr
//		}
//		if roomParam.BetScore < 1 || roomParam.BetScore > 4 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "底分ID错误！")
//			return mderr
//		}
//		if len(roomParam.SpecialCards) != 7 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "特殊牌型长度错误！")
//			return mderr
//		}
//		if len(roomParam.AdvanceOptions) != 3 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "高级选项长度错误！")
//			return mderr
//		}
//
//		for _, value := range roomParam.SpecialCards {
//			if value != "1" && value != "0" {
//				mderr.Detail = fmt.Sprintf(mderr.Detail, "特殊牌型格式错误！")
//				return mderr
//			}
//		}
//
//		if roomParam.AdvanceOptions[0] != "0" && roomParam.BankerType == 5 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "不能同时选择推注和通比！")
//			return mderr
//		}
//
//		if roomParam.AdvanceOptions[0] != "0" && roomParam.AdvanceOptions[0] != "1" && roomParam.AdvanceOptions[0] != "2" && roomParam.AdvanceOptions[0] != "3" {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "推注最高倍数格式错误！")
//			return mderr
//		}
//
//		if roomParam.SpecialCards[0] == "1" && roomParam.AdvanceOptions[1] == "1" {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "不能同时选择五花牛和不发花牌！")
//			return mderr
//		}
//
//		if maxNumber == 10 && roomParam.AdvanceOptions[1] == "1" { //|| (roomParam.SpecialCards[0] == "1" && roomParam.AdvanceOptions[1] == "1")
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "不能同时选择五花牛和10人模式！")
//			return mderr
//		}
//
//		break
//	case enumroom.DoudizhuGameType:
//		if maxNumber != 4 {
//			return errclub.ErrRoomMaxNumber
//		}
//		var roomParam *mdroom.DoudizhuRoomParam
//		mderr := errors.Parse(errclub.ErrGameParam.Error())
//		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
//			log.Err("doudizhu unmarshal room param failed, %v", err)
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
//			return mderr
//		}
//		if roomParam.BaseScore != 0 && roomParam.BaseScore != 5 && roomParam.BaseScore != 10 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "基本分格式错误！")
//			return mderr
//		}
//		break
//	case enumroom.FourCardGameType:
//		if maxNumber < 2 && maxNumber > 8 {
//			return errclub.ErrRoomMaxNumber
//		}
//		mderr := errors.Parse(errclub.ErrGameParam.Error())
//		var roomParam *mdroom.FourCardRoomParam
//		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
//			log.Err("fourcard unmarshal room param failed, %v", err)
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
//			return errclub.ErrGameParam
//		}
//		if roomParam.ScoreType < 1 || roomParam.ScoreType > 2 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "计分模式格式错误！")
//			return mderr
//		}
//		if roomParam.BetType < 1 || roomParam.BetType > 2 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "下注类型格式错误！")
//			return mderr
//		}
//		break
//	case enumroom.TwoCardGameType:
//		if maxNumber < 2 && maxNumber > 10 {
//			return errclub.ErrRoomMaxNumber
//		}
//		var roomParam *mdroom.TwoCardRoomParam
//		mderr := errors.Parse(errclub.ErrGameParam.Error())
//		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
//			log.Err("towcard unmarshal room param failed, %v", err)
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
//			return mderr
//		}
//		if roomParam.ScoreType < 1 || roomParam.ScoreType > 2 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "计分模式格式错误！")
//			return errclub.ErrGameParam
//		}
//		if roomParam.BetType < 1 || roomParam.BetType > 2 {
//			mderr.Detail = fmt.Sprintf(mderr.Detail, "下注类型格式错误！")
//			return mderr
//		}
//		break
//	default:
//		return errclub.ErrGameParam
//	}
//	return nil
//}

func checkGameParam(maxNumber int32, maxRound int32, gtype int32, gameParam string, RoomAdvanceOptions []string) (
	int32, int32, int32, string, []string, error) {
	if len(gameParam) == 0 {
		return 0, 0, 0, "", nil, errclub.ErrGameParam
	}
	if maxNumber < 2 {
		maxNumber = 2
		//return errroom.ErrRoomMaxNumber
	}
	if maxRound != 10 && maxRound != 20 && maxRound != 30 {
		maxRound = 10
		//return errroom.ErrRoomMaxRound
	}
	//fmt.Printf("ChekcGameParam:%d|%d|%d|%s\n",maxNumber,maxRound,gtype,gameParam)
	//if JoinType != enumroom.CanJoin && JoinType != enumroom.CanNotJoin {
	//	mderr := errors.Parse(errroom.ErrGameParam.Error())
	//	mderr.Detail = fmt.Sprintf(mderr.Detail, "加入类型格式错误！")
	//	return mderr
	//}

	if len(RoomAdvanceOptions) == 0 || (RoomAdvanceOptions[0] != "1" && RoomAdvanceOptions[0] != "0" ) {
		//mderr := errors.Parse(errroom.ErrGameParam.Error())
		//mderr.Detail = fmt.Sprintf(mderr.Detail, "房间参数格式错误！")
		//return mderr
		RoomAdvanceOptions = []string{"0"}
	}
	//if RoomAdvanceOptions[0] != "1" && RoomAdvanceOptions[0] != "0" {
	//	mderr := errors.Parse(errroom.ErrGameParam.Error())
	//	mderr.Detail = fmt.Sprintf(mderr.Detail, "房间允许加入参数格式错误！")
	//	return mderr
	//}

	switch gtype {
	case enumroom.ThirteenGameType:
		if maxNumber > 8 {
			maxNumber = 8
			//return errroom.ErrRoomMaxNumber
		}
		var roomParam *mdroom.ThirteenRoomParam
		mderr := errors.Parse(errclub.ErrGameParam.Error())
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("room check thirteen clean unmarshal room param failed, %v", err)
			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
			return 0, 0, 0, "", nil, mderr
		}
		//if roomParam.BankerType != 1 && roomParam.BankerType != 2 {
		//	return errors.ErrGameParam
		//}
		if roomParam.BankerAddScore < 0 || roomParam.BankerAddScore > 6 || roomParam.BankerAddScore%2 != 0 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "当庄加分格式错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.Joke != 0 && roomParam.Joke != 1 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "大小王格式错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.Times < 1 || roomParam.Times > 3 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "比赛模式格式错误！")
			return 0, 0, 0, "", nil, mderr
		}
		data, _ := json.Marshal(&roomParam)
		gameParam = string(data)
		break
	case enumroom.NiuniuGameType:
		if maxNumber != 4 && maxNumber != 6 && maxNumber != 8 && maxNumber != 10 {
			return 0, 0, 0, "", nil, errclub.ErrRoomMaxNumber
		}
		var roomParam *mdroom.NiuniuRoomParam
		mderr := errors.Parse(errclub.ErrGameParam.Error())
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("niuniu unmarshal room param failed, %v", err)
			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.BankerType < 1 || roomParam.BankerType > 5 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "玩法ID错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.Times != 1 && roomParam.Times != 2 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "倍数ID错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.BetScore < 1 || roomParam.BetScore > 4 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "底分ID错误！")
			return 0, 0, 0, "", nil, mderr
		}
		//if len(roomParam.SpecialCards) == 0 {
		//	//mderr.Detail = fmt.Sprintf(mderr.Detail, "特殊牌型长度错误！")
		//	//return 0, 0, 0, "", nil, mderr
		//	roomParam.SpecialCards = []string{"0", "0", "0", "0", "0", "0", "0"}
		//}
		lenSC := len(roomParam.SpecialCards)
		if lenSC > 7 {
			roomParam.SpecialCards = roomParam.SpecialCards[:7]
		}
		for i := 0; i < 7; i++ {
			if lenSC < i+1 {
				roomParam.SpecialCards = append(roomParam.SpecialCards, "0")
			} else {
				if roomParam.SpecialCards[i] != "0" && roomParam.SpecialCards[i] != "1" {
					roomParam.SpecialCards[i] = "0"
				}
			}
		}

		lenAo := len(roomParam.AdvanceOptions)
		if lenAo > 3 {
			roomParam.AdvanceOptions = roomParam.AdvanceOptions[:3]
		}
		for i := 0; i < 3; i++ {
			if lenAo < i+1 {
				roomParam.AdvanceOptions = append(roomParam.AdvanceOptions, "0")
			} else {
				if i == 0 && roomParam.AdvanceOptions[i] != "0" && roomParam.AdvanceOptions[i] != "1" &&
					roomParam.AdvanceOptions[i] != "2" && roomParam.AdvanceOptions[i] != "3" {
					roomParam.AdvanceOptions[i] = "0"
				} else if roomParam.AdvanceOptions[i] != "0" && roomParam.AdvanceOptions[i] != "1" {
					roomParam.AdvanceOptions[i] = "0"
				}
			}
		}

		//if len(roomParam.AdvanceOptions) != 3 {
		//	mderr.Detail = fmt.Sprintf(mderr.Detail, "高级选项长度错误！")
		//	return 0, 0, 0, "", nil, mderr
		//}
		//
		//for _, value := range roomParam.SpecialCards {
		//	if value != "1" && value != "0" {
		//		mderr.Detail = fmt.Sprintf(mderr.Detail, "特殊牌型格式错误！")
		//		return 0, 0, 0, "", nil, mderr
		//	}
		//}

		if roomParam.AdvanceOptions[0] != "0" && roomParam.BankerType == 5 {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "不能同时选择推注和通比！")
			return 0, 0, 0, "", nil, mderr
		}

		//if roomParam.AdvanceOptions[0] != "0" && roomParam.AdvanceOptions[0] != "1" && roomParam.AdvanceOptions[0] != "2" && roomParam.AdvanceOptions[0] != "3" {
		//	mderr.Detail = fmt.Sprintf(mderr.Detail, "推注最高倍数格式错误！")
		//	return 0, 0, 0, "", nil, mderr
		//}

		if roomParam.SpecialCards[0] == "1" && roomParam.AdvanceOptions[1] == "1" {
			mderr.Detail = fmt.Sprintf(mderr.Detail, "不能同时选择五花牛和不发花牌！")
			return 0, 0, 0, "", nil, mderr
		}

		if maxNumber == 10 && roomParam.AdvanceOptions[1] == "1" { //|| (roomParam.SpecialCards[0] == "1" && roomParam.AdvanceOptions[1] == "1")
			mderr.Detail = fmt.Sprintf(mderr.Detail, "不能同时选择五花牛和10人模式！")
			return 0, 0, 0, "", nil, mderr
		}
		data, _ := json.Marshal(&roomParam)
		gameParam = string(data)

		break
	case enumroom.DoudizhuGameType:
		if maxNumber != 4 {
			maxNumber = 4
			//return errroom.ErrRoomMaxNumber
		}
		var roomParam *mdroom.DoudizhuRoomParam
		mderr := errors.Parse(errclub.ErrGameParam.Error())
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("doudizhu unmarshal room param failed, %v", err)
			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.BaseScore != 0 && roomParam.BaseScore != 5 && roomParam.BaseScore != 10 {
			//mderr.Detail = fmt.Sprintf(mderr.Detail, "基本分格式错误！")
			//return mderr
			roomParam.BaseScore = 0
		}
		data, _ := json.Marshal(&roomParam)
		gameParam = string(data)
		break
	case enumroom.FourCardGameType:
		if maxNumber > 8 {
			maxNumber = 8
			//return errroom.ErrRoomMaxNumber
		}
		mderr := errors.Parse(errclub.ErrGameParam.Error())
		var roomParam *mdroom.FourCardRoomParam
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("fourcard unmarshal room param failed, %v", err)
			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
			return 0, 0, 0, "", nil, errclub.ErrGameParam
		}
		if roomParam.ScoreType < 1 || roomParam.ScoreType > 2 {
			roomParam.ScoreType = 1
			//mderr.Detail = fmt.Sprintf(mderr.Detail, "计分模式格式错误！")
			//return mderr

		}
		if roomParam.BetType < 1 || roomParam.BetType > 2 {
			roomParam.BetType = 1
			//mderr.Detail = fmt.Sprintf(mderr.Detail, "下注类型格式错误！")
			//return mderr
		}
		data, _ := json.Marshal(&roomParam)
		gameParam = string(data)
		break
	case enumroom.TwoCardGameType:
		if maxNumber > 10 {
			maxNumber = 10
		}
		var roomParam *mdroom.TwoCardRoomParam
		mderr := errors.Parse(errclub.ErrGameParam.Error())
		if err := json.Unmarshal([]byte(gameParam), &roomParam); err != nil {
			log.Err("towcard unmarshal room param failed, %v", err)
			mderr.Detail = fmt.Sprintf(mderr.Detail, "json解析错误！")
			return 0, 0, 0, "", nil, mderr
		}
		if roomParam.ScoreType < 1 || roomParam.ScoreType > 2 {
			roomParam.ScoreType = 1
			//mderr.Detail = fmt.Sprintf(mderr.Detail, "计分模式格式错误！")
			//return errroom.ErrGameParam
		}
		if roomParam.BetType < 1 || roomParam.BetType > 2 {
			roomParam.BetType = 1
			//mderr.Detail = fmt.Sprintf(mderr.Detail, "下注类型格式错误！")
			//return mderr
		}
		data, _ := json.Marshal(&roomParam)
		gameParam = string(data)
		break
	default:
		return 0, 0, 0, "", nil, errclub.ErrGameParam
	}

	return maxNumber, maxRound, gtype, gameParam, RoomAdvanceOptions, nil
}

func UpdateVipRoomSetting(mvrs *mdclub.VipRoomSetting) error {
	mdClub, err := cacheclub.GetClub(mvrs.ClubID)
	if err != nil {
		return err
	}

	if mdClub.Status != enumclub.ClubStatusNormal {
		return errclub.ErrStatusNoINNormal
	}
	if mvrs.Status == enumclub.VipRoomSettingNon {
		mvrs.MaxNumber, mvrs.RoundNumber, mvrs.GameType, mvrs.GameParam, mvrs.RoomAdvanceOptions, err =
			checkGameParam(mvrs.MaxNumber, mvrs.RoundNumber, mvrs.GameType, mvrs.GameParam, mvrs.RoomAdvanceOptions)
		if err != nil {
			return err
		}
		if mvrs.RoomType == enumroom.RoomTypeClub && (mvrs.SubRoomType != 0 && mvrs.SubRoomType != enumroom.SubTypeClubMatch) {
			return errclub.ErrGameParam
		}
		if mvrs.SubRoomType == enumroom.SubTypeClubMatch {
			var settingParam *mdroom.SettingParam
			if err := json.Unmarshal([]byte(mvrs.SettingParam), &settingParam); err != nil {
				log.Err("club vip room check setting param unmarshal failed, %v", err)
				return errclub.ErrGameParam
			}
			if settingParam.ClubCoinRate != 1 && settingParam.ClubCoinRate != 2 && settingParam.ClubCoinRate != 5 && settingParam.ClubCoinRate != 10 {
				return errclub.ErrGameParam
			}
		}
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		mvrcs, err := dbclub.UpdateVipRoomSetting(tx, mvrs)
		if err != nil {
			return err
		}

		if mvrcs.Status == enumclub.VipRoomSettingDel {
			err = cacheclub.DeleteVipRoomSetting(mvrcs.ClubID, mvrcs.ID)
			if err != nil {
				return nil
			}
		} else {
			err = cacheclub.SetVipRoomSetting(mvrcs)
			if err != nil {
				return nil
			}
		}

		return nil
	})
	if err != nil {
		return nil
	}

	return nil
}

func UpdateVipRoomSettingStatus(mvrs *mdclub.VipRoomSetting) error {
	mdClub, err := cacheclub.GetClub(mvrs.ClubID)
	if err != nil {
		return err
	}

	if mdClub.Status != enumclub.ClubStatusNormal {
		return errclub.ErrStatusNoINNormal
	}
	mdvrs, err := cacheclub.GetVipRoomSetting(mvrs.ClubID, mvrs.ID)
	if err != nil {
		return err
	}
	if mvrs.Status > enumclub.VipRoomSettingDel || mvrs.Status < enumclub.VipRoomSettingNon {
		return errclub.ErrStatus
	}
	if mvrs.Status == mdvrs.Status {
		return nil
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		mvrcs, err := dbclub.UpdateVipRoomSetting(tx, mvrs)
		if err != nil {
			return err
		}

		if mvrcs.Status == enumclub.VipRoomSettingDel {
			err = cacheclub.DeleteVipRoomSetting(mvrcs.ClubID, mvrcs.ID)
			if err != nil {
				return nil
			}
		} else {
			err = cacheclub.SetVipRoomSetting(mvrcs)
			if err != nil {
				return nil
			}
		}

		return nil
	})
	if err != nil {
		return nil
	}

	return nil
}

func GetVipRoomSettingByID(clubid int32, sid int32) (*mdclub.VipRoomSetting, error) {
	return cacheclub.GetVipRoomSetting(clubid, sid)
}

func GetVipRoomSettingList(clubid int32) ([]*mdclub.VipRoomSetting, error) {
	return cacheclub.GetAllVipRoomSetting(clubid)
}

func GetClubRoomLog(clubid int32) ([]*mdroom.ClubRoomLog, error) {
	out, err := dbroom.GetClubRoomLog(db.DB(), clubid)
	if err != nil {
		return nil, err
	}
	return out, nil
}
