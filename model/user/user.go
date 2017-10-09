package user

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	dbbill "playcards/model/bill/db"
	mdbill "playcards/model/bill/mod"
	mdpage "playcards/model/page"
	cacheuser "playcards/model/user/cache"
	dbu "playcards/model/user/db"
	"playcards/model/user/enum"
	erru "playcards/model/user/errors"
	mdu "playcards/model/user/mod"
	"playcards/utils/auth"
	"playcards/utils/db"
	"playcards/utils/errors"
	"playcards/utils/log"

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
		balance, _ := dbbill.GetUserBalance(tx, nu.UserID)
		nu.Diamond = balance.Diamond

		return nil
	}

	err = db.Transaction(f)
	return nu, err
}

func GetUser(u *mdu.User) (*mdu.User, error) {
	return dbu.GetUser(db.DB(), u)
}

func PageUserList(page *mdpage.PageOption, u *mdu.User) ([]*mdu.User, int64,
	error) {
	return dbu.PageUserList(db.DB(), page, u)
}

func UpdateUser(u *mdu.User) (*mdu.User, error) {
	f := func(tx *gorm.DB) error {
		user, err := dbu.UpdateUser(tx, u)
		if err != nil {
			return err
		}
		u = user
		return nil
	}
	err := db.Transaction(f)
	if err != nil {
		return nil, err
	}
	return u, err
}

func GetUserInfoSimple(userID int32) *mdu.User {
	_, u := cacheuser.GetUserByID(userID)
	return u
}

func GetWXToken(code string) (*mdu.AccessTokenResponse, error) {
	requestLine := strings.Join([]string{enum.GetTokenUrl, "?appid=",
		enum.AppId, "&Secret=", enum.Secret, "&code=", code,
		"&&grant_type=authorization_code"}, "")
	resp, err := http.Get(requestLine)

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Err("wx login http get Err:%s|%v", code, err)
		return nil, erru.ErrWXRequest
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err("wx login http get resp Err:%s|%v", code, err)
		return nil, erru.ErrWXResponse
	}

	atr := &mdu.AccessTokenResponse{}
	if bytes.Contains(body, []byte("access_token")) {
		err = json.Unmarshal(body, &atr)
		if err != nil {
			return nil, erru.ErrWXResponseJson
		}
	} else {
		ater := &mdu.AccessTokenErrorResponse{}
		err = json.Unmarshal(body, &ater)
		if err != nil {
			return nil, erru.ErrWXResponseJson
		}
		fmt.Printf("Refresh WX Token :%v \n", ater)
		log.Err("Refresh WX Token :%v \n", ater)
		return nil, erru.ErrWXParam
	}
	return atr, nil
}

func RefreshWXToken(refreshtoken string) (*mdu.AccessTokenResponse, error) {
	//appid=APPID&grant_type=refresh_token&refresh_token=REFRESH_TOKEN
	requestLine := strings.Join([]string{enum.RefreshTokenUrl, "?appid=",
		enum.AppId, "&grant_type=refresh_token&&refresh_token=", refreshtoken}, "")
	resp, err := http.Get(requestLine)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Err("wx login http get Err:%s|%v", refreshtoken, err)
		return nil, erru.ErrWXRequest
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err("wx login http get resp Err:%s|%v", refreshtoken, err)
		return nil, erru.ErrWXResponse
	}

	atr := &mdu.AccessTokenResponse{}
	if bytes.Contains(body, []byte("access_token")) {
		err = json.Unmarshal(body, &atr)
		if err != nil {
			return nil, erru.ErrWXResponseJson
		}
	} else {
		ater := &mdu.AccessTokenErrorResponse{}
		err = json.Unmarshal(body, &ater)
		if err != nil {
			return nil, erru.ErrWXResponseJson
		}
		fmt.Printf("Refresh WX Token :%v \n", ater)
		log.Err("Refresh WX Token :%v \n", ater)
		return nil, erru.ErrWXParam
	}
	return atr, nil
}

func GetWXUserInfo(token string, openID string) (*mdu.UserInfoResponse, error) {
	requestLine := strings.Join([]string{enum.GetUserUrl, "?access_token=",
		token, "&openid=", openID}, "")
	//fmt.Printf("Get WXUser Info:%s", requestLine)
	resp, err := http.Get(requestLine)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Err("wx get user info http get Err:%s|%v", openID, err)
		return nil, erru.ErrWXRequest
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err("wx get user http get resp Err:%s|%v", openID, err)
		return nil, erru.ErrWXResponse
	}

	user := &mdu.UserInfoResponse{}
	if bytes.Contains(body, []byte("openid")) {
		err = json.Unmarshal(body, &user)
		if err != nil {
			return nil, erru.ErrWXResponseJson
		}
	} else {
		ater := &mdu.AccessTokenErrorResponse{}
		err = json.Unmarshal(body, &ater)
		if err != nil {
			return nil, erru.ErrWXResponseJson
		}
		//msg := "微信查询用户提示：" + ater.Errmsg
		//return nil, errors.BadRequest(1009, msg)
		return nil, erru.ErrWXLoginParam
	}
	return user, nil
}

func GetAndCheckWXToken(openid string) (*mdu.AccessTokenResponse, error) {
	refreshtoken, err := cacheuser.GetRefreshToken(openid)

	if err != nil {
		log.Err("user login get refresh session failed, %v", err)
		return nil, err
	}

	accesstoken, err := RefreshWXToken(refreshtoken)
	if err != nil {
		log.Err("refresh wx token failed, %s|%v", refreshtoken, err)
		return nil, err
	}
	return accesstoken, nil
}

//GetRefreshToken
func WXLogin(u *mdu.User, code string) (int32, *mdu.User, error) {

	if u.OpenID == "" && code == "" {
		return enum.ResultStatusFail, nil, erru.ErrWXLoginParam
	}
	atr := &mdu.AccessTokenResponse{}
	if u.OpenID == "" {
		checkatr, err := GetWXToken(code)

		if err != nil {
			return enum.ResultStatusFail, nil, err
		}

		atr = checkatr
		u.OpenID = atr.OpenID
		//fmt.Printf("AAA WX Login Get New Token:%v", atr)
	} else {
		checkatr, err := GetAndCheckWXToken(u.OpenID) //cacheuser.GetAccessToken(u.OpenID)
		if err != nil {
			log.Err("user login set session failed, %v", err)
			return enum.ResultStatusFail, nil, err
		}
		atr = checkatr
		//fmt.Printf("BBB WX Login Get Old Token:%v", atr)
	}

	err := cacheuser.SetUserWXInfo(atr.OpenID, atr.AccessToken, atr.RefreshToken)
	if err != nil {
		return enum.ResultStatusSuccess, nil, err
	}

	u, err = dbu.FindAndGetUser(db.DB(), u)

	if err != nil {
		return enum.ResultStatusFail, nil, err
	}

	if u == nil {
		u = &mdu.User{}
		u, err = CreateUserByWX(u, atr)
		if err != nil {
			return enum.ResultStatusFail, nil, err
		}
	} else {
		u, err = UpdateUserFromWX(u, atr)
		if err != nil {
			return enum.ResultStatusFail, nil, err
		}
	}

	f := func(tx *gorm.DB) error {
		err = tx.Model(u).UpdateColumn("last_login_at", gorm.NowFunc()).Error
		if err != nil {
			return errors.Internal("login failed", err)
		}
		balance, _ := dbbill.GetUserBalance(tx, u.UserID)
		u.Diamond = balance.Diamond

		return nil
	}

	err = db.Transaction(f)
	if err != nil {
		return enum.ResultStatusFail, nil, err
	}

	return enum.ResultStatusSuccess, u, err
}

func CreateUserByWX(u *mdu.User, atr *mdu.AccessTokenResponse) (*mdu.User, error) {
	u, err := UpdateUserFromWX(u, atr)
	if err != nil {
		return nil, err
	}

	f := func(tx *gorm.DB) error {
		fmt.Printf("AAA Create User ByWX:%v", u)
		uid, err := dbu.AddUser(tx, u)
		if err != nil {
			return err
		}
		fmt.Printf("BBB Create User ByWX:%v", u)
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
	if err != nil {
		return nil, err
	}
	return u, nil
}

func UpdateUserFromWX(u *mdu.User, atr *mdu.AccessTokenResponse) (*mdu.User, error) {
	var err error

	ui, err := GetWXUserInfo(atr.AccessToken, atr.OpenID)
	if err != nil {
		return nil, err
	}

	u.Nickname = ui.Nickname
	u.OpenID = ui.OpenID
	u.UnionID = ui.UnionID
	u.Sex = ui.Sex
	u.Icon = ui.Headimgurl

	return u, err
}
