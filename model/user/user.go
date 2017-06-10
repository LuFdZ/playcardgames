package user

import (
	"crypto/sha256"
	"fmt"

	dbbill "playcards/model/bill/db"
	mdbill "playcards/model/bill/mod"
	mdpage "playcards/model/page"
	dbu "playcards/model/user/db"
	"playcards/model/user/enum"
	erru "playcards/model/user/errors"
	mdu "playcards/model/user/mod"
	"playcards/utils/auth"
	"playcards/utils/db"

	"playcards/utils/errors"

	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v8"
)

var registerValid *validator.Validate

func init() {
	registerValid = validator.New(&validator.Config{TagName: "reg"})
}

func Register(u *mdu.User) (int32, error) {
	var uid int32
	var err error

	err = registerValid.Struct(u)
	if err != nil {
		return enum.ErrUID, erru.ErrInvalidUserInfo
	}

	hash := sha256.Sum256([]byte(u.Password + enum.Salt))
	u.Password = fmt.Sprintf("%x", hash)

	// TODO: delete this before release
	u.Rights = auth.RightsAdmin

	f := func(tx *gorm.DB) error {
		uid, err = dbu.AddUser(tx, u)
		if err != nil {
			return err
		}

		b := &mdbill.Balance{
			Gold:    0,
			Diamond: 0,
			Amount:  enum.RegisterBalance,
		}
		err = dbbill.CreateBalance(tx, uid, b)
		if err != nil {
			return err
		}

		return nil
	}

	err = db.Transaction(f)
	return uid, err
}

func Login(u *mdu.User) (*mdu.User, error) {
	var nu *mdu.User
	var err error

	_, err = govalidator.ValidateStruct(u)
	if err != nil {
		return nil, erru.ErrInvalidUserInfo
	}

	hash := sha256.Sum256([]byte(u.Password + enum.Salt))
	u.Password = fmt.Sprintf("%x", hash)

	f := func(tx *gorm.DB) error {
		nu, err = dbu.GetUser(tx, &mdu.User{
			Username: u.Username,
			Password: u.Password,
		})

		if err != nil {
			return err
		}

		err = tx.Model(nu).UpdateColumn("last_login_at", gorm.NowFunc()).Error
		if err != nil {
			return errors.Internal("login failed", err)
		}

		return nil
	}

	err = db.Transaction(f)
	return nu, err
}

func PageUserList(page *mdpage.PageOption, u *mdu.User) ([]*mdu.User, int64,
	error) {
	return dbu.PageUserList(db.DB(), page, u)
}

func UpdateUser(u *mdu.User) (*mdu.User, error) {
	return dbu.UpdateUser(db.DB(), u)
}
