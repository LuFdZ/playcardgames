package db

import (
	"fmt"
	mdpage "playcards/model/page"
	enumu "playcards/model/user/enum"
	erru "playcards/model/user/errors"
	mdu "playcards/model/user/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	"github.com/Masterminds/squirrel"
	"github.com/jinzhu/gorm"
)

func AddUser(tx *gorm.DB, u *mdu.User) (int32, error) {
	if err := tx.Create(u).Error; err != nil {
		return enumu.ErrUID, errors.Internal("add user failed", err)
	}
	return u.UserID, nil
}

func GetUser(tx *gorm.DB, u *mdu.User) (*mdu.User, error) {
	qr := &struct {
		mdu.User
	}{}

	selsql := fmt.Sprintf("%s.*", enumu.UserTableName)

	if err := tx.Select(selsql).Table(enumu.UserTableName).
		Where(u).Find(qr).Error; err != nil {
		return nil, errors.Internal("get user failed", err)
	}
	return &qr.User, nil
}

// valid user id from db , if not exist , will return invalid user error.
func ValidUserID(tx *gorm.DB, uid int32) error {
	if found, _ := db.FoundRecord(tx.Select("user_id").
		Where("user_id = ?", uid).Find(&mdu.User{}).Error); !found {
		return erru.ErrUserNotExisted
	}
	return nil
}

func UpdateUser(tx *gorm.DB, u *mdu.User) (*mdu.User, error) {
	user := &mdu.User{
		//Username: u.Username,
		Nickname:     u.Nickname,
		Email:        u.Email,
		Rights:       u.Rights,
		MobileOs:     u.MobileOs,
		Version:      u.Version,
		Mobile:       u.Mobile,
		Icon:         u.Icon,
		InviteUserID: u.InviteUserID,
		LastLoginIP:  u.LastLoginIP,
		ClubID:       u.ClubID,
	}
	if err := tx.Model(u).Updates(user).Error; err != nil {
		return nil, errors.Internal("update user failed", err)
	}
	return u, nil
}

func ReSetUserClubID(tx *gorm.DB, u *mdu.User) error {
	if err := tx.Model(u).Update("club_id", 0).Error; err != nil {
		return errors.Internal("reset user club_id failed", err)
	}
	return nil
}

func PageUserList(tx *gorm.DB, page *mdpage.PageOption, u *mdu.User) (
	[]*mdu.User, int64, error) {
	var (
		list []*mdu.User
		etx  = tx.Model(u)
	)
	str := ""
	if len(u.Username) > 0 {
		if len(str) > 0 {
			str += " and"
		}
		str += " username like '%" + u.Username + "%'"
	}
	if len(u.Nickname) > 0 {
		if len(str) > 0 {
			str += " and"
		}
		str += " nickname like '%" + u.Nickname + "%'"
	}
	if len(u.Channel) > 0 {
		if len(str) > 0 {
			str += " and"
		}
		str += " channel like '%" + u.Channel + "%'"
	}
	if page.StartAt != nil {
		if len(str) > 0 {
			str += " and"
		}
		str += fmt.Sprintf(" created_at > %s", page.StartAt.Unix())
	}
	if page.EndAt != nil {
		if len(str) > 0 {
			str += " and"
		}
		str += fmt.Sprintf(" end_at < %s", page.EndAt.Unix())
	}

	if u.UserID != 0 {
		etx = etx.Where("user_id = ?", u.UserID)
	}

	if len(u.OpenID) > 0 {
		if len(str) > 0 {
			str += " and"
		}
		str += " open_id = '" + u.OpenID + "'"
	}

	if len(u.UnionID) > 0 {
		if len(str) > 0 {
			str += " and"
		}
		str += " union_id = '" + u.UnionID + "'"
	}

	rows, res := page.Find(etx.Where(str), &list)
	if res.Error != nil {
		return nil, 0, errors.Internal("page user list error", res.Error)
	}

	return list, rows, nil
}

func GetInvitedUserCount(tx *gorm.DB, uid int32) ([]mdu.User, error) {
	var (
		out []mdu.User
	)
	sql, param, err := squirrel.
		Select(" user_id,icon "). //
		From(enumu.UserTableName).
		Where(" invite_user_id = ? ", uid).ToSql()

	if err != nil {
		return nil, errors.Internal("get user id list failed", err)
	}

	err = tx.Raw(sql, param...).Scan(&out).Error
	if err != nil {
		return nil, errors.Internal("get list failed", err)
	}
	return out, nil
}

func FindAndGetUser(tx *gorm.DB, u *mdu.User) (*mdu.User, error) {
	//fmt.Printf("WXLogin:%v", u)
	found, err := db.FoundRecord(tx.Where("open_id = ?", u.OpenID).Find(&u).Error)
	if err != nil {
		return nil, errors.Internal("get user failed ", err)
	}
	if !found {
		return nil, nil
	}
	//fmt.Printf("FindAndGetUser UserInfo:%s|%s|%s\n", u.MobileOs, u.Version, u.Channel)
	return u, nil
}
