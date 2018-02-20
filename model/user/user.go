package user

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	dbbill "playcards/model/bill/db"
	mdpage "playcards/model/page"
	cacheuser "playcards/model/user/cache"
	dbu "playcards/model/user/db"
	"playcards/model/user/enum"
	erru "playcards/model/user/errors"
	mdu "playcards/model/user/mod"
	"math/rand"
	"playcards/utils/auth"
	"playcards/utils/db"
	"playcards/utils/log"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v8"
	"playcards/utils/tools"
)

var registerValid *validator.Validate

func init() {
	registerValid = validator.New(&validator.Config{TagName: "reg"})
}

func GetRobotNickName() string {
	rand.Seed(time.Now().UnixNano())
	nickname := enum.FirsnameList[tools.GenerateRangeNum(0, len(enum.FirsnameList))] +
		enum.NameList[tools.GenerateRangeNum(0, len(enum.NameList))]
	nickname = mdu.EncodNickName(nickname)
	return nickname
}

func RegisterRobotUser(count int32) {
	var i int32 = 0

	f := func(tx *gorm.DB) error {
		for ; i < count; i++ {
			index := fmt.Sprintf("Robot:%d:%d", i, time.Now().Nanosecond()/1000)
			nickname := GetRobotNickName()
			exist := cacheuser.CheckExistRobot(nickname)
			for i := 0; exist && i < 3; i++ {
				nickname = GetRobotNickName()
				exist = cacheuser.CheckExistRobot(nickname)
			}
			u := &mdu.User{
				Username: index,
				Nickname: nickname,
				Type:     enum.Robot,
				OpenID:   index,
				Sex:      i%2 + 1,
			}
			_, err := dbu.AddUser(tx, u)
			cacheuser.SetRobot(u)
			if err != nil {
				log.Err("register robot user index:%d,robot:%+v,err:%v\n", i, u, err)
				continue
			}
		}
		return nil
	}
	go db.Transaction(f)
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
	u.OpenID = u.Username
	u.Type = enum.Player
	f := func(tx *gorm.DB) error {
		uid, err = dbu.AddUser(tx, u)
		if err != nil {
			return err
		}
		err = dbbill.CreateAllBalance(tx, uid)
		if err != nil {
			return err
		}
		return nil
	}

	err = db.Transaction(f)
	return uid, err
}

func Login(u *mdu.User, address string) (*mdu.User, error) {
	var nu *mdu.User
	var err error
	//if u.Username == "liufangzhou"{
	//	TestRegisterUser()
	//}
	_, err = govalidator.ValidateStruct(u)
	if err != nil {
		return nil,erru.ErrInvalidUserInfo
	}

	hash := sha256.Sum256([]byte(u.Password + enum.Salt))
	u.Password = fmt.Sprintf("%x", hash)
	//var diamond int64
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
		nu.Version = u.Version
		nu.MobileOs = u.MobileOs
		nu.Channel = u.Channel
		_, err := UpdateUser(nu)
		if err != nil {
			return err
		}
		//nu.Diamond = balance.Diamond
		//balance, _ := bill.GetUserBalance(nu.UserID, 2)
		//diamond = balance.Balance
		return nil
	}

	err = db.Transaction(f)
	if err != nil {
		return nil,err
	}
	return nu, nil
}

func GetUser(u *mdu.User) (*mdu.User, error) {
	return dbu.GetUser(db.DB(), u)
}

func PageUserList(page *mdpage.PageOption, u *mdu.User,sort int32) ([]*mdu.User, int64,
	error) {
	return dbu.PageUserList(db.DB(), page, u,sort)
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
		log.Err("wx login http get Err:%s|%+v", code, err)
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
		log.Err("access_token refresh WX token :%v \n", ater)
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
		log.Err("wx login http get Err:%s|%+v", refreshtoken, err)
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
		log.Err("refresh_token refresh WX token :%v \n", ater)
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

func WXLogin(u *mdu.User, code string, address string) (*mdu.User, error) {

	if u.OpenID == "" && code == "" {
		return nil, erru.ErrWXLoginParam
	}
	atr := &mdu.AccessTokenResponse{}

	if u.OpenID == "" {
		checkatr, err := GetWXToken(code)

		if err != nil {
			return nil, err
		}

		atr = checkatr
		u.OpenID = atr.OpenID
		//fmt.Printf("AAA WX Login Get New Token:%v", atr)
	} else {
		checkatr, err := GetAndCheckWXToken(u.OpenID) //cacheuser.GetAccessToken(u.OpenID)
		if err != nil {
			log.Err("user login set session failed, %v", err)
			return nil, err
		}
		atr = checkatr
		//fmt.Printf("BBB WX Login Get Old Token:%v", atr)
	}

	err := cacheuser.SetUserWXInfo(atr.OpenID, atr.RefreshToken)
	if err != nil {
		return nil, err
	}
	newUser, err := dbu.FindAndGetUser(db.DB(), u.OpenID)

	if err != nil {
		return nil, err
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
		newUser.Type = enum.Player
		u, err = CreateUserByWX(newUser, atr)
		if err != nil {
			return nil, err
		}
	} else {
		newUser.MobileOs = u.MobileOs
		newUser.Version = u.Version
		newUser.LastLoginIP = address
		u, err = UpdateUserFromWX(newUser, atr)
		if err != nil {
			return nil, err
		}
	}
	now := gorm.NowFunc()
	u.LastLoginAt = &now

	u, err = UpdateUser(u)
	if err != nil {
		return nil, err
	}

	return u, err
}

func CreateUserByWX(u *mdu.User, atr *mdu.AccessTokenResponse) (*mdu.User, error) {
	u, err := UpdateUserFromWX(u, atr)
	if err != nil {
		return nil, err
	}
	u.Rights = auth.RightsPlayer
	u.Username = u.OpenID
	f := func(tx *gorm.DB) error {
		uid, err := dbu.AddUser(tx, u)
		if err != nil {
			return err
		}
		err = dbbill.CreateAllBalance(tx, uid)
		if err != nil {
			return err
		}
		return nil

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

//func GetUserRealBalance(uid int32) (*mdbill.UserBalance, error) {
//	balance, err := bill.GetUserBalance(uid)
//	if err != nil {
//		return nil, err
//	}
//	if balance == nil {
//		return nil, erru.ErrBillNotExisted
//	}
//	lockBalance, err := cacheuser.GetUserLockBalance(uid)
//	if err != nil {
//		return nil, err
//	}
//	if lockBalance != nil {
//		balance.Diamond -= lockBalance.Diamond
//		balance.Gold -= lockBalance.Gold
//	}
//	return balance, nil
//}

//func SetUserLockBalance(uid int32, balanceType int32, amount int64, rid int32) error {
//	lb := &mdu.Balance{}
//	if balanceType == enumbill.TypeGold {
//		lb.Gold = amount
//	} else if balanceType == enumbill.TypeDiamond {
//		lb.Diamond = amount
//	}
//
//	f := func() error {
//		err := cacheuser.SetUserLockBalance(uid, lb)
//		if err != nil {
//			return err
//		}
//		return nil
//	}
//	lock := fmt.Sprintf("playcards.room.userbalance.lock:%s", uid)
//	err := gsync.GlobalTransaction(lock, f)
//	if err != nil {
//		log.Err("%s enter room failed: %v", lock, err)
//		return err
//	}
//	if err != nil {
//		return err
//	}
//	log.Debug("SetUserLockBalance rid:%d,uid:%d,balanceType:%d,amount:%d", rid, uid, balanceType, amount)
//	return nil
//}

func DayActiveUserList(page int32) ([]*mdu.User, *mdpage.PageReply) {
	timeStr := time.Now().Format("2006-01-02")
	nowData, _ := time.Parse("2006-01-02", timeStr)
	sub8h, _ := time.ParseDuration("-1h")
	nowData = nowData.Add(8 * sub8h)
	f := func(u *mdu.User) bool {
		if u.LastLoginAt.Unix() > nowData.Unix() {
			return true
		}
		return false
	}
	us, count := cacheuser.GetUserList(f, page)
	mpr := &mdpage.PageReply{
		PageNow:   page,
		PageTotal: count,
	}
	return us, mpr
}

func GetUserOnlineCount() (int32, error) {
	count := cacheuser.GetAllOnlineCount()
	return int32(count), nil
}

func DayActiveUserCount() int32 {
	count := dbu.GetDayActiveUserConut()
	return count
}

func DayNewUserCount() int32 {
	count := dbu.GetNewUserConut()
	return count
}

func RefreshUserCount() error {
	count := dbu.GetUserConut()
	err := cacheuser.SetUserNumber(count)
	if err != nil {
		return err
	}
	return nil
}

func SetLocation(user *mdu.User, Json string) error {
	if len(Json) > 300 {
		return erru.ErrParamTooLong
	}
	user.Location = Json
	err := cacheuser.SimpleUpdateUser(user)
	if err != nil {
		return err
	}
	return nil
}

func RefreshAllRobotsFromDB() error {
	rs, err := dbu.GetRobots(db.DB())
	if err != nil {
		return err
	}
	return cacheuser.SetRobots(rs)
}
