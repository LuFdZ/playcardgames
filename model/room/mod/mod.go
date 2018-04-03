package mod

import (
	"encoding/base64"
	"encoding/json"
	mdtime "playcards/model/time"
	cacheuser "playcards/model/user/cache"
	enumr "playcards/model/room/enum"
	pbr "playcards/proto/room"
	"playcards/utils/log"
	utilproto "playcards/utils/proto"
	"time"

	"github.com/jinzhu/gorm"
)

type Room struct {
	RoomID           int32             `gorm:"primary_key"`
	Password         string //`reg:"required,min=6,max=32,excludesall="`
	Status           int32
	Giveup           int32
	MaxNumber        int32 //`reg:"required"`
	RoundNumber      int32 //`reg:"required"`
	RoundNow         int32
	UserList         string
	RoomType         int32 //`reg:"required"`
	PayerID          int32
	GameType         int32  //`reg:"required"`
	GameParam        string //`reg:"required,min=1,excludesall="`
	GameUserResult   string
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
	GiveupAt         *time.Time
	Shuffle          int32
	Flag             int32
	ClubID           int32
	Cost             int64
	CostType         int32
	Level            int32
	SubRoomType      int32
	SettingParam     string
	StartMaxNumber   int32
	VipRoomSettingID int32
	GameIDNow        int32             `gorm:"-"`
	ShuffleAt        *time.Time        `gorm:"-"`
	ReadyAt          *time.Time        `gorm:"-"`
	UserResults      []*GameUserResult `gorm:"-"`
	Users            []*RoomUser       `gorm:"-"`
	GiveupGame       GiveUpGameResult  `gorm:"-"`
	HasNotice        bool              `gorm:"-"`
	BankerList       []int32           `gorm:"-"`
	ReadyUserMap     map[int32]int64   `gorm:"-"`
	Ids              []int32           `gorm:"-"`
	BigWiners        []int32           `gorm:"-"`
	PlayerIds        []int32           `gorm:"-"`
	RobotIds         []int32           `gorm:"-"`
	WatchIds         []int32           `gorm:"-"`
	//SearchKey      string            `gorm:"-"`
}

type GiveUpGameResult struct {
	Status        int32
	UserStateList []*UserState
}

type UserState struct {
	State  int32
	UserID int32
}

type SettingParam struct {
	CostType          int32
	CostValue         int32
	ClubCoinBaseScore int64
	ClubCoinRate      int32
	CostRange         int32
}

type RoomUser struct {
	UserID       int32
	Nickname     string
	Ready        int32
	Position     int32
	Icon         string
	Sex          int32
	Role         int32
	UpdatedAt    *time.Time
	Location     string
	Online       int32
	Type         int32
	CoinType     int32
	Join         int32
	Gold         int64
	ResultAmount int32
	ClubCoin     int64
	RoomCost     int64
}

type GameUserResult struct {
	UserID             int32
	Nickname           string
	Win                int32
	Lost               int32
	Tie                int32
	Score              int32
	Role               int32
	GameCardCount      string
	RoundScore         int32
	RoundClubCoinScore int64
	TotalClubCoinScore int64
	ClubCoin           int64
	LastRole           int32
	LastPushOnBet      int32
	PushOnBetScore     int32
	CanPushOn          int32
}

type RoomResults struct {
	RoomID          int32
	RoundNumber     int32
	RoundNow        int32
	Status          int32
	Password        string
	GameType        int32
	CreatedAt       *time.Time
	List            []*GameUserResult
	GameParam       string
	MaxPlayerNumber int32
	PlayerNumberNow int32
	RoomType        int32
	SubRoomType     int32
	ClubCoin        int64
}

type PlayerRoom struct {
	LogID     int32
	UserID    int32
	RoomID    int32
	GameType  int32
	PlayTimes int32
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type RoomResultList struct {
	List []*RoomResults
}

type CheckRoomExist struct {
	Result       int32
	Status       int32
	Room         Room
	GiveupResult GiveUpGameResult
	GameResult   RoomResults
}

type ThirteenRoomParam struct {
	BankerAddScore int32
	Time           int32
	Joke           int32
	Times          int32
}

type NiuniuRoomParam struct {
	Times          int32
	BankerType     int32 //上庄类型
	PreBankerID    int32
	BetScore       int32 //底分选择ID 1 2 3 4
	AdvanceOptions []string
	SpecialCards   []string //特殊牌型
}

type DoudizhuRoomParam struct {
	BaseScore      int32
	PreBankerID    int32
	PreBankerIndex int32
}

type FourCardRoomParam struct {
	ScoreType int32
	BetType   int32
}

type TwoCardRoomParam struct {
	ScoreType int32
	BetType   int32
}

type PlayerSpecialGameRecord struct {
	GameID     int32
	RoomID     int32
	GameType   int32
	RoomType   int32
	Password   string
	UserID     int32
	GameResult string
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

type ClubRoomLog struct {
	Date              string
	VipRoomCount      int32
	ClubRoomCount     int32
	ClubCoinRoomCount int32
	TotalRoomCount    int32
}

//func (r *Room) String() string {
//	return fmt.Sprintf("[roomid: %d pwd: %s status: %d gametype: %d]",
//		r.RoomID, r.Password, r.Status, r.GameType)
//}

func (us *UserState) ToProto() *pbr.UserState {
	out := &pbr.UserState{
		State:  us.State,
		UserID: us.UserID,
	}
	return out
}

func (gur *GiveUpGameResult) ToProto() *pbr.GiveUpGameResult {
	var uss []*pbr.UserState
	for _, us := range gur.UserStateList {
		uss = append(uss, us.ToProto())
	}
	out := &pbr.GiveUpGameResult{
		//RoomID:        gur.RoomID,
		Status:        gur.Status,
		UserStateList: uss,
	}
	return out
}

func (r *Room) ToProto() *pbr.Room {
	out := &pbr.Room{
		RoomID:           r.RoomID,
		Password:         r.Password,
		MaxNumber:        r.MaxNumber,
		Status:           r.Status,
		Giveup:           r.Giveup,
		GameType:         r.GameType,
		RoundNumber:      r.RoundNumber,
		RoundNow:         r.RoundNow,
		RoomType:         r.RoomType,
		Flag:             r.Flag,
		ClubID:           r.ClubID,
		Shuffle:          r.Shuffle,
		StartMaxNumber:   r.StartMaxNumber,
		BankerList:       r.BankerList,
		CreatedAt:        mdtime.TimeToProto(r.CreatedAt),
		UpdatedAt:        mdtime.TimeToProto(r.UpdatedAt),
		GameParam:        r.GameParam,
		SubRoomType:      r.SubRoomType,
		SettingParam:     r.SettingParam,
		VipRoomSettingID: r.VipRoomSettingID,
		//GameUserResult:r.GameUserResult,
	}
	utilproto.ProtoSlice(r.Users, &out.UserList)
	return out
}

func RoomFromProto(r *pbr.Room) *Room {
	out := &Room{
		RoomID:      r.RoomID,
		Password:    r.Password,
		MaxNumber:   r.MaxNumber,
		Status:      r.Status,
		Giveup:      r.Giveup,
		GameType:    r.GameType,
		RoundNumber: r.RoundNumber,
		RoundNow:    r.RoundNow,
		RoomType:    r.RoomType,
		ClubID:      r.ClubID,
		CreatedAt:   mdtime.TimeFromProto(r.CreatedAt),
		UpdatedAt:   mdtime.TimeFromProto(r.UpdatedAt),
		GameParam:   r.GameParam,
	}
	return out
}

func (r *RoomUser) ToProto() *pbr.RoomUser {
	mRu := &pbr.RoomUser{
		UserID:   r.UserID,
		Ready:    r.Ready,
		Position: r.Position,
		Role:     r.Role,
		Gold:     r.Gold,
		ClubCoin: r.ClubCoin,
	}
	_, u := cacheuser.GetUserByID(r.UserID)
	if u != nil {
		mRu.Nickname = u.Nickname
		mRu.Icon = u.Icon
		mRu.Sex = u.Sex
		mRu.Location = u.Location
		mRu.Online = cacheuser.GetUserOnlineStatus(r.UserID)
	}
	if r.Type == enumr.Robot {
		mRu.Nickname = r.Nickname
		mRu.Sex = r.Sex
		mRu.Online = 1
	}
	return mRu
}

func (r *RoomUser) SimplyToProto() *pbr.RoomUser {
	mRu := &pbr.RoomUser{
		UserID:   r.UserID,
		Ready:    r.Ready,
		Position: r.Position,
		Role:     r.Role,
	}
	return mRu
}

func (r *RoomResults) ToProto() *pbr.RoomResults {
	out := &pbr.RoomResults{
		RoomID:          r.RoomID,
		Password:        r.Password,
		RoundNumber:     r.RoundNumber,
		RoundNow:        r.RoundNow,
		CreatedAt:       mdtime.TimeToProto(r.CreatedAt),
		Status:          r.Status,
		GameType:        r.GameType,
		GameParam:       r.GameParam,
		MaxPlayerNumber: r.MaxPlayerNumber,
		PlayerNumberNow: r.PlayerNumberNow,
		RoomType:        r.RoomType,
		SubRoomType:     r.SubRoomType,
	}
	utilproto.ProtoSlice(r.List, &out.List)
	return out
}

func (ur *GameUserResult) ToProto() *pbr.GameUserResult {
	mGur := &pbr.GameUserResult{
		UserID:             ur.UserID,
		Nickname:           ur.Nickname, //DecodNickName(ur.Nickname),
		Win:                ur.Win,
		Lost:               ur.Lost,
		Tie:                ur.Tie,
		Score:              ur.Score,
		GameCardCount:      ur.GameCardCount,
		TotalClubCoinScore: ur.TotalClubCoinScore,
		ClubCoin:           ur.ClubCoin,
	}
	_, u := cacheuser.GetUserByID(ur.UserID)
	if u != nil {
		mGur.Nickname = u.Nickname
		mGur.Icon = u.Icon
	}
	return mGur
}

func (cre *CheckRoomExist) ToProto() *pbr.CheckRoomExistReply {
	out := &pbr.CheckRoomExistReply{}
	out.Result = cre.Result
	out.Status = cre.Status
	if &cre.Room == nil {
		out.Room = nil
	} else {
		out.Room = cre.Room.ToProto()
	}
	if &cre.GiveupResult == nil {
		out.GiveupResult = nil
	} else {
		out.GiveupResult = cre.GiveupResult.ToProto()
		out.GiveupResult.CountDown = &pbr.CountDown{
			ServerTime: cre.Room.GiveupAt.Unix(),
			Count:      enumr.RoomGiveupCleanMinutes * 60,
		}
	}
	if &cre.GameResult == nil {
		out.GameResult = nil
	} else {
		out.GameResult = cre.GameResult.ToProto()
	}
	return out
	//return &pbr.CheckRoomExistReply{
	//	Result:       cre.Result,
	//	Status:       cre.Status,
	//	Room:         cre.Room.ToProto(),
	//	GiveupResult: cre.GiveupResult.ToProto(),
	//	GameResult:   cre.GameResult.ToProto(),
	//}
}

func (psgr *PlayerSpecialGameRecord) ToProto() *pbr.PlayerSpecialGameRecord {
	out := &pbr.PlayerSpecialGameRecord{
		GameID:     psgr.GameID,
		RoomID:     psgr.RoomID,
		GameType:   psgr.GameType,
		RoomType:   psgr.RoomType,
		Password:   psgr.Password,
		UserID:     psgr.UserID,
		GameResult: psgr.GameResult,
		CreatedAt:  mdtime.TimeToProto(psgr.CreatedAt),
		UpdatedAt:  mdtime.TimeToProto(psgr.UpdatedAt),
	}
	return out
}

func (crl *ClubRoomLog) ToProto() *pbr.ClubRoomLog {
	out := &pbr.ClubRoomLog{
		Date:           crl.Date,
		VipRoomCount:   crl.VipRoomCount,
		ClubRoomCount:  crl.ClubRoomCount,
		TotalRoomCount: crl.TotalRoomCount,
	}
	return out
}

func GameRecordFromProto(psgr *pbr.PlayerSpecialGameRecord) *PlayerSpecialGameRecord {
	out := &PlayerSpecialGameRecord{
		GameID:   psgr.GameID,
		RoomID:   psgr.RoomID,
		GameType: psgr.GameType,
		RoomType: psgr.RoomType,
		Password: psgr.Password,
		UserID:   psgr.UserID,
	}
	return out
}

func (r *Room) BeforeUpdate(scope *gorm.Scope) error {
	r.MarshalUsers()
	r.MarshalGameUserResult()
	scope.SetColumn("user_list", r.UserList)
	scope.SetColumn("game_user_result", r.GameUserResult)
	return nil
}

func (r *Room) BeforeCreate(scope *gorm.Scope) error {
	r.MarshalUsers()
	r.MarshalGameUserResult()
	scope.SetColumn("user_list", r.UserList)
	scope.SetColumn("game_user_result", r.GameUserResult)
	return nil
}

func (r *Room) AfterFind() error {
	err := r.UnmarshalGameUserResult()
	if err != nil {
		return err
	}
	err = r.UnmarshalUsers()
	if err != nil {
		return err
	}
	return nil
}

func (r *Room) MarshalUsers() error {
	data, _ := json.Marshal(&r.Users)
	r.UserList = string(data)
	return nil
}

func (r *Room) UnmarshalUsers() error {
	var out []*RoomUser
	if err := json.Unmarshal([]byte(r.UserList), &out); err != nil {
		return err
	}
	r.Users = out
	return nil
}

func (r *Room) MarshalGameUserResult() error {
	data, _ := json.Marshal(&r.UserResults)
	r.GameUserResult = string(data)
	return nil
}

func (r *Room) UnmarshalGameUserResult() error {
	if len(r.GameUserResult) == 0 {
		return nil
	}
	var out []*GameUserResult
	if err := json.Unmarshal([]byte(r.GameUserResult), &out); err != nil {
		return err
	}
	r.UserResults = out
	return nil
}

type Feedback struct {
	FeedbackID    int32 `gorm:"primary_key"`
	UserID        int32
	Channel       string
	Version       string
	Content       string
	MobileModel   string
	MobileNetWork string
	MobileOs      string
	LoginIP       string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

func (fb *Feedback) ToProto() *pbr.Feedback {
	return &pbr.Feedback{
		FeedbackID:    fb.FeedbackID,
		UserID:        fb.UserID,
		Channel:       fb.Channel,
		Version:       fb.Version,
		Content:       fb.Content,
		MobileModel:   fb.MobileModel,
		MobileNetWork: fb.MobileNetWork,
		MobileOs:      fb.MobileOs,
		LoginIP:       fb.LoginIP,
		CreatedAt:     mdtime.TimeToProto(fb.CreatedAt),
		UpdatedAt:     mdtime.TimeToProto(fb.UpdatedAt),
	}
}

func FeedbackFromProto(fb *pbr.Feedback) *Feedback {
	return &Feedback{
		FeedbackID:    fb.FeedbackID,
		UserID:        fb.UserID,
		Channel:       fb.Channel,
		Version:       fb.Version,
		Content:       fb.Content,
		MobileModel:   fb.MobileModel,
		MobileNetWork: fb.MobileNetWork,
		MobileOs:      fb.MobileOs,
		LoginIP:       fb.LoginIP,
		CreatedAt:     mdtime.TimeFromProto(fb.CreatedAt),
		UpdatedAt:     mdtime.TimeFromProto(fb.UpdatedAt),
	}
}

func DecodNickName(nikename string) string {
	uDec, err := base64.StdEncoding.DecodeString(nikename)
	if err != nil {
		log.Err("EncodNickName nickname:%s,err:%v", nikename, err)
	}
	//nikename = string(uDec)
	return string(uDec)
}
