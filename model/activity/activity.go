package activity

import (
	dba "playcards/model/activity/db"
	enum "playcards/model/activity/enum"
	"playcards/model/activity/errors"
	"playcards/model/bill"
	enumbill "playcards/model/bill/enum"
	mbill "playcards/model/bill/mod"
	dbu "playcards/model/user/db"
	mdu "playcards/model/user/mod"
	"playcards/utils/date"
	"playcards/utils/db"
	"time"

	"github.com/jinzhu/gorm"
)

func Invite(u *mdu.User, inviterID int32) (int32, []*mbill.UserBalance, error) {
	var (
		isBalance = true
		balances  []*mbill.UserBalance
		err       error
	)
	err = nil
	checktd, err := dbu.GetUser(db.DB(), &mdu.User{UserID: u.UserID})
	if err != nil {
		return 0, nil, err
	}

	if checktd.InviteUserID > 0 {
		return 0, nil, errors.ErrHadInviter
	}
	if u.UserID == inviterID {
		return 0, nil, errors.ErrInviterSelf
	}

	err = dbu.ValidUserID(db.DB(), inviterID)
	if err != nil {
		return 0, nil, errors.ErrInviterNoExist
	}

	list, err := dbu.GetInvitedUserCount(db.DB(), u.UserID)
	//fmt.Printf("Invite List:%v", list)
	if err != nil {
		return 0, nil, err
	}

	for _, inviter := range list {
		if inviter.UserID == inviterID {
			return 0, nil, errors.ErrInviterConflict
		}
	}

	u.InviteUserID = inviterID
	var result int32
	result = 1

	//fmt.Printf("Invite date Count:%d|%d|%v\n", date.TimeSubDays(time.Now(), *u.CreatedAt), u.UserID,u.CreatedAt)
	if date.TimeSubDays(time.Now(), *u.CreatedAt) > enum.InviteCreateDaysLimit {
		isBalance = false
		result = 2
	} else {
		number, err := dbu.GetInvitedUserCount(db.DB(), inviterID)
		if err != nil {
			return 0, nil, err
		}
		//fmt.Printf("Invite Number:%v", number)
		if len(number) > enum.InviteTimesLimit {
			isBalance = false
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
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return 0, nil, err
	}

	if isBalance {
		//JournalTypeInvite	JournalTypeShate
		invite, err := bill.GainBalanceType(u.UserID, time.Now().Unix(),
			&mbill.Balance{Diamond: enum.InviteDiamond},
			enumbill.JournalTypeInvite)
		if err != nil {
			return 0, nil, err
		}
		balances = append(balances, invite)

		invited, err := bill.GainBalanceType(inviterID, time.Now().Unix(),
			&mbill.Balance{Diamond: enum.InviteDiamond},
			enumbill.JournalTypeInvited)
		if err != nil {
			return 0, nil, err
		}
		balances = append(balances, invited)
	}

	return result, balances, nil

}

func Share(uid int32) (*mbill.UserBalance, error) { //*mda.PlayerShare,
	ps, err := dba.GetPlayerShare(db.DB(), uid)
	if err != nil {
		return nil, err
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
			return nil, err
		}
	}
	if date.TimeSubDays(time.Now(), *ps.UpdatedAt) > 1 {
		ps.ShareTimes = 0
	}

	if ps.ShareTimes >= enum.ShareLimitTimes || ps.TotalDiamonds >= enum.ShareLimitDiamond {
		return nil, nil //errors.ErrShareNoDiamonds
	}

	ub, err := bill.GainBalanceType(uid, time.Now().Unix(),
		&mbill.Balance{Diamond: enum.ShareDiamond}, enumbill.JournalTypeShare)
	if err != nil {
		return nil, err
	}
	ps.TotalDiamonds += enum.ShareDiamond
	ps.ShareTimes++
	f := func(tx *gorm.DB) error {
		err := dba.UpdatePlayerShare(tx, ps)
		if err != nil {
			return nil
		}
		return nil
	}
	if err := db.Transaction(f); err != nil {
		return nil, err
	}
	return ub, nil
}

func InviteUserInfo(uid int32) (int32, error) {
	list, err := dbu.GetInvitedUserCount(db.DB(), uid)
	if err != nil {
		return 0, nil
	}
	return int32(len(list)), nil
}

// message InviteReply{
//     int32 Result = 1;
//     int64 Diamond = 3;
// }
