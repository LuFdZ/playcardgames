package activity

import (
	dba "playcards/model/activity/db"
	enumact "playcards/model/activity/enum"
	"playcards/model/activity/errors"
	//"playcards/model/bill"
	//enumbill "playcards/model/bill/enum"
	//enumcom "playcards/model/common/enum"
	//mbill "playcards/model/bill/mod"
	dbu "playcards/model/user/db"
	mdu "playcards/model/user/mod"
	cacheuser "playcards/model/user/cache"
	"playcards/utils/date"
	"playcards/utils/db"
	"time"

	"github.com/jinzhu/gorm"
	"playcards/utils/tools"
)

func Invite(u *mdu.User, inviterID int32) (int32, error) {//[]*mbill.Balance,
	var (
		//isBalance = true
		//balances  []*mbill.Balance
		err       error
	)
	err = nil
	//checktd, err := dbu.GetUser(db.DB(), &mdu.User{UserID: u.UserID})
	//if err != nil {
	//	return 0, nil, err
	//}

	if u.InviteUserID > 0 {
		return 0,errors.ErrHadInviter
	}
	if u.UserID == inviterID {
		return 0,errors.ErrInviterSelf
	}

	err = dbu.ValidUserID(db.DB(), inviterID)
	if err != nil {
		return 0, errors.ErrInviterNoExist
	}

	list, err := dbu.GetInvitedUserCount(db.DB(), u.UserID)
	//fmt.Printf("Invite List:%v", list)
	if err != nil {
		return 0, err
	}

	for _, inviter := range list {
		if inviter.UserID == inviterID {
			return 0, errors.ErrInviterConflict
		}
	}

	u.InviteUserID = inviterID
	var result int32
	result = 1

	//fmt.Printf("Invite date Count:%d|%d|%v\n", date.TimeSubDays(time.Now(), *u.CreatedAt), u.UserID,u.CreatedAt)
	if date.TimeSubDays(time.Now(), *u.CreatedAt) > enumact.InviteCreateDaysLimit {
		//isBalance = false
		result = 2
	} else {
		number, err := dbu.GetInvitedUserCount(db.DB(), inviterID)
		if err != nil {
			return 0, err
		}
		//fmt.Printf("Invite Number:%v", number)
		if len(number) > enumact.InviteTimesLimit {
			//isBalance = false
			result = 3
		}
	}
	// _, err = cacheu.SetUser(u)
	// if err != nil {
	// 	return 0, nil, err
	// }
	f := func(tx *gorm.DB) error {
		user, err := dbu.UpdateUser(tx, u)
		if err != nil {
			return err
		}
		u = user
		err = cacheuser.SimpleUpdateUser(u)
		if err != nil {
			return nil
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return 0, err
	}

	//if isBalance {
	//	//JournalTypeInvite	JournalTypeShate
	//	invite, err := bill.GainBalance(u.UserID, time.Now().Unix(),enumbill.JournalTypeInvite,
	//		&mbill.Balance{Amount: enumact.InviteDiamond, CoinType: enumcom.Diamond},
	//		)
	//	if err != nil {
	//		return 0, nil, err
	//	}
	//	balances = append(balances, invite)
	//
	//	invited, err := bill.GainBalance(inviterID, time.Now().Unix(),enumbill.JournalTypeInvited,
	//		&mbill.Balance{Amount: enumact.InviteDiamond, CoinType: enumcom.Diamond},
	//		)
	//	if err != nil {
	//		return 0, nil, err
	//	}
	//	balances = append(balances, invited)
	//}
	return result, nil

}

func Share(uid int32) error { //(*mbill.Balance, error) *mda.PlayerShare,
	ps, err := dba.GetPlayerShare(db.DB(), uid)
	if err != nil {
		return err
	}
	if ps == nil {
		f := func(tx *gorm.DB) error {
			ps, err = dba.CreatePlayerShare(db.DB(), uid)
			if err != nil {
				return err
			}
			return nil
		}
		if err := db.Transaction(f); err != nil {
			return err
		}
	}
	if date.TimeSubDays(time.Now(), *ps.UpdatedAt) > 1 {
		ps.ShareTimes = 0
	}

	if ps.ShareTimes >= enumact.ShareLimitTimes || ps.TotalDiamonds >= enumact.ShareLimitDiamond {
		return errors.ErrShareNoDiamonds
	}

	//ub, err := bill.GainBalance(uid, time.Now().Unix(),enumbill.JournalTypeShare,
	//	&mbill.Balance{Amount: enumact.ShareDiamond, CoinType: enumcom.Diamond})
	//if err != nil {
	//	return nil, err
	//}
	ps.TotalDiamonds += enumact.ShareDiamond
	ps.ShareTimes++
	f := func(tx *gorm.DB) error {
		err := dba.UpdatePlayerShare(tx, ps)
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

func InviteUserInfo(uid int32) (string, error) {
	list, err := dbu.GetInvitedUserCount(db.DB(), uid)
	if err != nil {
		return "0", nil
	}
	return tools.IntToString(int32(len(list))), nil
}

// message InviteReply{
//     int32 Result = 1;
//     int64 Diamond = 3;
// }
