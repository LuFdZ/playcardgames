package activity

import (
	dba "playcards/model/activity/db"
	enum "playcards/model/activity/enum"
	"playcards/model/activity/errors"
	mda "playcards/model/activity/mod"
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

func AddActivityConfig(ac *mda.ActivityConfig) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return dba.AddActivityConfig(tx, ac)
	})
}

func UpdateActivityConfig(ac *mda.ActivityConfig) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return dba.UpdateActivityConfig(tx, ac)
	})
}

func DeleteActivityConfig(ac *mda.ActivityConfig) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return dba.DeleteActivityConfig(tx, ac)
	})
}

func ActivityConfigList() ([]*mda.ActivityConfig, error) {
	return dba.ActivityConfigList(db.DB())
}

func Invite(u *mdu.User, inviterID int32) (int32, []*mbill.UserBalance, error) {
	var (
		isBalance = true
		balances  []*mbill.UserBalance
	)

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

	u.InviteUserID = inviterID
	var result int32
	result = 1

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

	if ps.ShareTimes == 3 || ps.TotalDiamonds == enum.ShareLimitDiamond {
		return nil, errors.ErrShareNoDiamonds
	}

	ub, err := bill.GainBalanceType(uid, time.Now().Unix(),
		&mbill.Balance{Diamond: enum.ShareDiamond}, enumbill.JournalTypeShate)
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

// message InviteReply{
//     int32 Result = 1;
//     int64 Diamond = 3;
// }
