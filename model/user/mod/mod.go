package mod

import (
	"encoding/base64"
	"fmt"
	mdtime "playcards/model/time"
	pbu "playcards/proto/user"
	"playcards/utils/log"
	"time"
	"unicode/utf8"
)

type User struct {
	UserID        int32  `gorm:"primary_key"`
	Username      string `reg:"required,min=6,max=32,excludesall= 	"`
	Password      string `reg:"required,min=6,max=32,excludesall= 	"`
	Nickname      string `reg:"required,min=6,max=16,excludesall= 	"`
	Mobile        string `reg:"omitempty,min=6,max=16,excludesall= 	"`
	Email         string `reg:"min=6,max=30,email,excludesall= 	"`
	Avatar        string
	Status        int32
	Channel       string `reg:"omitempty,min=6,max=64,excludesall= 	"`
	Version       string `reg:"omitempty,min=6,max=64,excludesall= 	"`
	Icon          string
	Sex           int32
	Rights        int32
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	LastLoginAt   *time.Time
	InviteUserID  int32
	ClubID        int32
	MobileUuID    string
	MobileModel   string
	MobileNetWork string
	MobileOs      string `reg:"omitempty,min=1,max=64,excludesall= 	"`
	LastLoginIP   string
	RegIP         string
	OpenID        string
	UnionID       string
	AccessToken   string `gorm:"-"`
	Location      string `gorm:"-"`
}

type UserInfo struct {
	UserID   int32
	Username string
	Nickname string
	Icon     string
	Sex      int32
}

type AccessTokenResponse struct {
	AccessToken  string  `json:"access_token"`
	ExpiresIn    float64 `json:"expires_in"`
	RefreshToken string  `json:"refresh_token"`
	OpenID       string  `json:"openid"`
	Scope        string  `json:"scope"`
}

type AccessTokenErrorResponse struct {
	Errcode float64 `json:"errcode"`
	Errmsg  string  `json:"errmsg"`
}

type UserInfoResponse struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int32  `json:"sex"`
	Headimgurl string `json:"headimgurl"`
	UnionID    string `json:"unionid"`
}

func (u *User) String() string {
	return fmt.Sprintf("[uid: %d, name: %s, rights: %d,openid:%s,mobile:%s,nickname:%s]", u.UserID,
		u.Username, u.Rights, u.OpenID, u.Mobile, u.Nickname)
}

func (u *User) ToProto() *pbu.User {
	return &pbu.User{
		UserID:        u.UserID,
		Username:      u.Username,
		Password:      u.Password,
		Nickname:      u.Nickname, //DecodNickName(u.Nickname),
		Mobile:        u.Mobile,
		Email:         u.Email,
		Avatar:        u.Avatar,
		Status:        u.Status,
		Channel:       u.Channel,
		Version:       u.Version,
		Rights:        u.Rights,
		CreatedAt:     mdtime.TimeToProto(u.CreatedAt),
		UpdatedAt:     mdtime.TimeToProto(u.UpdatedAt),
		LastLoginAt:   mdtime.TimeToProto(u.LastLoginAt),
		InviteUserID:  u.InviteUserID,
		MobileUuID:    u.MobileUuID,
		MobileModel:   u.MobileModel,
		MobileNetWork: u.MobileNetWork,
		MobileOs:      u.MobileOs,
		LastLoginIP:   u.LastLoginIP,
		RegIP:         u.RegIP,
		Icon:          u.Icon,
		Sex:           u.Sex,
		OpenID:        u.OpenID,
		UnionID:       u.UnionID,
		ClubID:        u.ClubID,
	}
}
func (u *UserInfo) ToProto() *pbu.UserInfo {
	return &pbu.UserInfo{
		UserID:   u.UserID,
		Username: u.Username,
		Nickname: u.Nickname,
		Icon:     u.Icon,
		Sex:      u.Sex,
	}
}

func UserFromPageRequestProto(u *pbu.PageUserListRequest) *User {
	return &User{
		UserID:   u.UserID,
		Username: u.Username,
		Nickname: EncodNickName(u.Nickname),
		Rights:   u.Rights,
		OpenID:   u.OpenID,
		UnionID:  u.UnionID,
	}
}

func UserFromProto(u *pbu.User) *User {
	return &User{
		UserID:        u.UserID,
		Username:      u.Username,
		Password:      u.Password,
		Nickname:      EncodNickName(u.Nickname),
		Mobile:        u.Mobile,
		Email:         u.Email,
		Avatar:        u.Avatar,
		Status:        u.Status,
		Channel:       u.Channel,
		Version:       u.Version,
		Rights:        u.Rights,
		CreatedAt:     mdtime.TimeFromProto(u.CreatedAt),
		UpdatedAt:     mdtime.TimeFromProto(u.UpdatedAt),
		LastLoginAt:   mdtime.TimeFromProto(u.LastLoginAt),
		InviteUserID:  u.InviteUserID,
		MobileUuID:    u.MobileUuID,
		MobileModel:   u.MobileModel,
		MobileNetWork: u.MobileNetWork,
		MobileOs:      u.MobileOs,
		LastLoginIP:   u.LastLoginIP,
		RegIP:         u.RegIP,
		Icon:          u.Icon,
		Sex:           u.Sex,
		OpenID:        u.OpenID,
		UnionID:       u.UnionID,
	}
}

func UserFromWXLoginRequestProto(u *pbu.WXLoginRequest) *User {
	return &User{
		OpenID:        u.OpenID,
		Mobile:        u.Mobile,
		Channel:       u.Channel,
		Version:       u.Version,
		MobileUuID:    u.MobileUuID,
		MobileModel:   u.MobileModel,
		MobileNetWork: u.MobileNetWork,
		MobileOs:      u.MobileOs,
	}
}

func (u *User) AfterFind() error {
	u.Password = ""
	return nil
}

func EncodNickName(nikename string) string {
	input := []byte(nikename)
	uEnc := base64.StdEncoding.EncodeToString([]byte(input))
	return string(uEnc)
}

func DecodNickName(nikename string) string {
	uDec, err := base64.StdEncoding.DecodeString(nikename)
	if err != nil {
		log.Err("EncodNickName nickname:%s,err:%v", nikename, err)
	}
	//nikename = string(uDec)
	return string(uDec)
}

/***
过滤坑爹的Emoji表情
*/
func FilterEmoji(content string) string {
	new_content := ""
	for _, value := range content {
		_, size := utf8.DecodeRuneInString(string(value))
		if size <= 3 {
			new_content += string(value)
		}
	}
	return new_content
}
