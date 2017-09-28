package mod

import (
	"encoding/json"
	"fmt"
	mdtime "playcards/model/time"
	pbr "playcards/proto/room"
	utilproto "playcards/utils/proto"
	"time"

	"github.com/jinzhu/gorm"
)

type Room struct {
	RoomID         int32  `gorm:"primary_key"`
	Password       string `reg:"required,min=6,max=32,excludesall= 	"`
	Status         int32
	MaxNumber      int32 `reg:"required"`
	RoundNumber    int32 `reg:"required"`
	RoundNow       int32
	UserList       string
	RoomType       int32 `reg:"required"`
	PayerID        int32
	GameType       int32  `reg:"required"`
	GameParam      string `reg:"required,min=1,excludesall= 	"`
	GameUserResult string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time

	UserResults []*GameUserResult `gorm:"-"`
	Users       []*RoomUser       `gorm:"-"`
	GiveupGame  GiveUpGameResult  `gorm:"-"`
}

type GiveUpGameResult struct {
	RoomID        int32
	Status        int32
	UserStateList []*UserState
}

type UserState struct {
	State  int32
	UserID int32
}

type RoomUser struct {
	UserID    int32
	Nickname  string
	Ready     int32
	Position  int32
	Icon      string
	Sex       int32
	Role      int32
	UpdatedAt *time.Time
	RoomID    int32
}

type GameUserResult struct {
	UserID        int32
	Nickname      string
	Win           int32
	Lost          int32
	Tie           int32
	Score         int32
	Role          int32
	GameCardCount string //interface{}
}

type RoomResults struct {
	RoomID      int32
	RoundNumber int32
	RoundNow    int32
	Status      int32
	Password    string
	GameType    int32
	CreatedAt   *time.Time
	List        []*GameUserResult
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

func (r *Room) String() string {
	return fmt.Sprintf("[roomid: %d pwd: %s status: %d gametype: %d]",
		r.RoomID, r.Password, r.Status, r.GameType)
}

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
		RoomID:        gur.RoomID,
		Status:        gur.Status,
		UserStateList: uss,
	}
	return out
}

func (r *Room) ToProto() *pbr.Room {
	out := &pbr.Room{
		RoomID:      r.RoomID,
		Password:    r.Password,
		MaxNumber:   r.MaxNumber,
		Status:      r.Status,
		GameType:    r.GameType,
		RoundNumber: r.RoundNumber,
		RoundNow:    r.RoundNow,
		CreatedAt:   mdtime.TimeToProto(r.CreatedAt),
		UpdatedAt:   mdtime.TimeToProto(r.UpdatedAt),
	}
	utilproto.ProtoSlice(r.Users, &out.UserList)
	return out
}

func (r *RoomUser) ToProto() *pbr.RoomUser {
	return &pbr.RoomUser{
		UserID:   r.UserID,
		Nickname: r.Nickname,
		Ready:    r.Ready,
		Position: r.Position,
		Icon:     r.Icon,
		Sex:      r.Sex,
		Role:     r.Role,
		RoomID:   r.RoomID,
	}
}

func (r *RoomResults) ToProto() *pbr.RoomResults {
	out := &pbr.RoomResults{
		RoomID:      r.RoomID,
		Password:    r.Password,
		RoundNumber: r.RoundNumber,
		RoundNow:    r.RoundNow,
		CreatedAt:   mdtime.TimeToProto(r.CreatedAt),
		Status:      r.Status,
	}
	utilproto.ProtoSlice(r.List, &out.List)
	return out
}

func (ur *GameUserResult) ToProto() *pbr.GameUserResult {
	return &pbr.GameUserResult{
		UserID:        ur.UserID,
		Nickname:      ur.Nickname,
		Win:           ur.Win,
		Lost:          ur.Lost,
		Tie:           ur.Tie,
		Score:         ur.Score,
		GameCardCount: ur.GameCardCount,
	}
}

// func (rs *RoomResultList) ToProto() *pbr.RoomResultListReply {
// 	out := &pbr.RoomResultListReply{
// 		List: nil,
// 	}
// 	utilproto.ProtoSlice(r.List, &out.List)
// 	return out
// }

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
