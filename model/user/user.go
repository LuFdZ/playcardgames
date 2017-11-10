package user

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"playcards/model/bill"
	"playcards/model/user/enum"
	dbbill "playcards/model/bill/db"
	mdbill "playcards/model/bill/mod"
	mdpage "playcards/model/page"
	cacheuser "playcards/model/user/cache"
	dbu "playcards/model/user/db"
	enumbill "playcards/model/bill/enum"
	erru "playcards/model/user/errors"
	mdu "playcards/model/user/mod"
	gsync "playcards/utils/sync"
	"playcards/utils/auth"
	"playcards/utils/db"
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
			Diamond: enum.NewUserDiamond,
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

func Login(u *mdu.User, address string) (*mdu.User, int64, error) {
	var nu *mdu.User
	var err error

	_, err = govalidator.ValidateStruct(u)
	if err != nil {
		return nil, 0, erru.ErrInvalidUserInfo
	}

	hash := sha256.Sum256([]byte(u.Password + enum.Salt))
	u.Password = fmt.Sprintf("%x", hash)
	var diamond int64
	f := func(tx *gorm.DB) error {
		nu, err = dbu.GetUser(tx, &mdu.User{
			Username: u.Username,
			Password: u.Password,
		})

		if err != nil {
			return err
		}

		//err = tx.Model(nu).UpdateColumn("last_login_at", gorm.NowFunc()).Error
		//if err != nil {
		//	return errors.Internal("login failed", err)
		//}
		//
		//err = tx.Model(nu).UpdateColumn("last_login_ip", address).Error
		//if err != nil {
		//	return errors.Internal("login failed", err)
		//}
		now := gorm.NowFunc()
		nu.LastLoginAt = &now
		nu.LastLoginIP = address
		_, err := UpdateUser(nu)
		if err != nil {
			return err
		}
		//nu.Diamond = balance.Diamond
		balance, _ := GetUserRealBalance(nu.UserID)
		diamond = balance.Diamond
		return nil
	}

	err = db.Transaction(f)
	return nu, diamond, err
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
	user.Nickname = mdu.EncodNickName(user.Nickname)
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

func WXLogin(u *mdu.User, code string, address string) (int64, *mdu.User, error) {
	//fmt.Printf("AAAAWXLogin UserInfo:%s|%s|%s\n",u.MobileOs,u.Version,u.Channel)
	if u.OpenID == "" && code == "" {
		return 0, nil, erru.ErrWXLoginParam
	}
	atr := &mdu.AccessTokenResponse{}

	if u.OpenID == "" {
		checkatr, err := GetWXToken(code)

		if err != nil {
			return 0, nil, err
		}

		atr = checkatr
		u.OpenID = atr.OpenID
		//fmt.Printf("AAA WX Login Get New Token:%v", atr)
	} else {
		checkatr, err := GetAndCheckWXToken(u.OpenID) //cacheuser.GetAccessToken(u.OpenID)
		if err != nil {
			log.Err("user login set session failed, %v", err)
			return 0, nil, err
		}
		atr = checkatr
		//fmt.Printf("BBB WX Login Get Old Token:%v", atr)
	}

	err := cacheuser.SetUserWXInfo(atr.OpenID, atr.AccessToken, atr.RefreshToken)
	if err != nil {
		return 0, nil, err
	}
	newUser, err := dbu.FindAndGetUser(db.DB(), u)

	if err != nil {
		return 0, nil, err
	}

	if newUser == nil {
		newUser = &mdu.User{}
		newUser.LastLoginIP = address
		newUser.Channel = u.Channel
		newUser.Version = u.Version
		newUser.MobileOs = u.MobileOs
		newUser.RegIP = address
		newUser.OpenID = u.OpenID
		newUser.UnionID = u.UnionID
		u, err = CreateUserByWX(newUser, atr)
		if err != nil {
			return 0, nil, err
		}
	} else {
		newUser.MobileOs = u.MobileOs
		newUser.Version = u.Version
		newUser.LastLoginIP = address
		u, err = UpdateUserFromWX(newUser, atr)
		if err != nil {
			return 0, nil, err
		}
	}
	now := gorm.NowFunc()
	u.LastLoginAt = &now
	u, err = UpdateUser(u)
	if err != nil {
		return 0, nil, err
	}
	balance, _ := GetUserRealBalance(u.UserID)
	return balance.Diamond, u, err
}

func CreateUserByWX(u *mdu.User, atr *mdu.AccessTokenResponse) (*mdu.User, error) {
	u, err := UpdateUserFromWX(u, atr)
	if err != nil {
		return nil, err
	}
	u.Rights = auth.RightsPlayer
	f := func(tx *gorm.DB) error {
		uid, err := dbu.AddUser(tx, u)
		if err != nil {
			return err
		}
		b := &mdbill.Balance{
			Gold:    0,
			Diamond: enum.NewUserDiamond,
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
	//u.EncodNickName()
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

func GetUserRealBalance(uid int32)(*mdbill.UserBalance, error){
	balance, err := bill.GetUserBalance(uid)
	if err !=nil{
		return nil,err
	}
	lockBalance,err := cacheuser.GetUserLockBalance(uid)
	if err !=nil{
		return nil,err
	}
	if lockBalance!=nil{
		balance.Diamond -= lockBalance.Diamond
		balance.Gold -= lockBalance.Gold
	}
	return balance,nil
}

func SetUserLockBalance(uid int32,balanceType int32,amount int64,rid int32) error{
	lb := &mdu.Balance{}
	if balanceType == enumbill.TypeGold {
		lb.Gold = amount
	}else if balanceType == enumbill.TypeDiamond{
		lb.Diamond = amount
	}

	f := func() error {
		err := cacheuser.SetUserLockBalance(uid,lb)
		if err !=nil{
			return err
		}
		return nil
	}
	lock := fmt.Sprintf("playcards.room.userbalance.lock:%s", uid)
	err := gsync.GlobalTransaction(lock, f)
	if err != nil {
		log.Err("%s enter room failed: %v", lock, err)
		return err
	}
	if err != nil{
		return err
	}
	log.Debug("SetUserLockBalance rid:%d,uid:%d,balanceType:%d,amount:%d",rid,uid,balanceType,amount)
	return nil
}

