package club

import (
	cacheclub "playcards/model/club/cache"
	dbclub "playcards/model/club/db"
	enumclub "playcards/model/club/enum"
	errclub "playcards/model/club/errors"
	mdclub "playcards/model/club/mod"
	"playcards/model/common"
	cachecon "playcards/model/common/cache"
	dbcon "playcards/model/common/db"
	enumcon "playcards/model/common/enum"
	errcon "playcards/model/common/errors"
	mdcon "playcards/model/common/mod"
	mdpage "playcards/model/page"
	cacheroom "playcards/model/room/cache"
	dbroom "playcards/model/room/db"
	enumroom "playcards/model/room/enum"
	mdroom "playcards/model/room/mod"
	cacheuser "playcards/model/user/cache"
	dbuser "playcards/model/user/db"
	mduser "playcards/model/user/mod"
	pbclub "playcards/proto/club"
	pbroom "playcards/proto/room"
	"playcards/utils/db"
	utilpb "playcards/utils/proto"
	utilproto "playcards/utils/proto"

	"github.com/jinzhu/gorm"
)

func CreateClub(name string, creatorid int32, creatorproxy int32) error {
	checklen := len(name)
	if checklen < 3 || checklen > 60 {
		return errclub.ErrClubNameLen
	}
	if creatorid == 0 || creatorid < 100000 {
		return errclub.ErrCreatorid
	}
	if creatorproxy == 0 {
		return errclub.ErrCreatorid
	}
	mclub := &mdclub.Club{
		ClubName:     name,
		Status:       enumclub.ClubStatusExamine,
		CreatorID:    creatorid,
		CreatorProxy: creatorproxy,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
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
	return nil
}
func SetClubBalance(amonut int64, amonuttype int32, clubid int32, typeid int32, foreign int64, opid int64) error {
	mClub, err := cacheclub.GetClub(clubid)
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		_, err = dbclub.ClubBalance(tx, mClub, amonuttype, amonut, typeid, foreign, opid)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil
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

func RemoveClubMember(clubid int32, uid int32, removeType int32) error {
	mcm, err := cacheclub.GetClubMember(clubid, uid)
	if mcm == nil {
		return errclub.ErrNotInClub
	}
	_, muser := cacheuser.GetUserByID(uid)
	if muser == nil {
		return errcon.ErrUserErr
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

func CreateClubMember(clubid int32, uid int32) (*mduser.User, error) { //muser *mduser.User
	_, muser := cacheuser.GetUserByID(uid)
	if muser == nil {
		return nil, errcon.ErrUserErr
	}
	mClub, err := CheckUserJoinClubRoom(clubid, muser)
	if err != nil {
		return nil, err
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
		return nil, nil
	}

	err = cacheuser.SimpleUpdateUser(muser)
	if err != nil {
		return nil, nil
	}
	err = cacheclub.SetClubMember(mcm)
	if err != nil {
		return nil, nil
	}

	return muser, nil
}

func CheckUserJoinClubRoom(clubid int32, muser *mduser.User) (*mdclub.Club, error) {
	if muser.ClubID > 0 {
		return nil, errclub.ErrAlreadyInClub
	}

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

func UpdateClubExamine(clubid int32, uid int32, status int32, opid int32) error {
	err := common.UpdateExamine(enumcon.TypeClub, clubid, uid, status, opid)
	if err != nil {
		return err
	}
	_, err = CreateClubMember(clubid, uid)
	if err != nil {
		return err
	}
	return nil
}

func GetClub(muser *mduser.User) (*pbclub.ClubInfo, error) {
	if muser.ClubID == 0 {
		return nil, errclub.ErrNotJoinAnyClub
	}
	mClub, err := cacheclub.GetClub(muser.ClubID)
	if err != nil {
		return nil, err
	}

	if mClub == nil {
		return nil, errclub.ErrClubNotExisted
	}

	if mClub.Status != enumclub.ClubStatusNormal {
		return nil, errclub.ErrStatusNoINNormal
	}
	ci := &pbclub.ClubInfo{
		UserID: muser.UserID,
		Club:   mClub.ToProto(),
	}
	mcms := cacheclub.GetAllClubMember(muser.ClubID, false)
	utilpb.ProtoSlice(mcms, &ci.ClubMemberList)

	f := func(r *mdroom.Room) bool {
		if r.Status < enumroom.RoomStatusDelay &&
			r.ClubID == muser.ClubID {
			return true
		}
		return false
	}
	rooms := cacheroom.GetAllRoom(f)
	utilpb.ProtoSlice(rooms, &ci.RoomList)
	return ci, nil
}

func GetClubInfo(muser *mduser.User) (*mdclub.Club, error) {
	if muser.ClubID == 0 {
		return nil, errclub.ErrNotJoinAnyClub
	}
	mClub, err := cacheclub.GetClub(muser.ClubID)
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
	[]*mdclub.ClubMember, int64, error) {
	page.Page -= 1
	return dbclub.PageClubMember(db.DB(), page, mem)
}

func PageClubRoom(clubid int32, page int32, pagesize int32,flag int32) (
	[]*pbroom.Room, error) {
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = 20
	}
	if flag == 0{
		flag = 2
	}
	rooms, err := dbroom.PageRoomList(db.DB(), clubid, page, pagesize,flag)
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

func RefreshAllFromDB() error {
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
	return nil
}
