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
	GameType       int32
	GameParam      string `reg:"required,min=1,excludesall= 	"`
	GameUserResult string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time

	UserResults []*GameUserResult
	Users       []*RoomUser `gorm:"-"`
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
	UserID int32
	Win    int32
	Lost   int32
	Tie    int32
	Score  int32
}

type RoomResults struct {
	RoomID int32
	List   []*GameUserResult
}

func (r *Room) String() string {
	return fmt.Sprintf("[roomid: %d userlist: %s status: %d gametype: %d]",
		r.RoomID, r.UserList, r.Status, r.GameType)
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
		RoomID: r.RoomID,
	}
	utilproto.ProtoSlice(r.List, &out.List)
	return out
}

func (ur *GameUserResult) ToProto() *pbr.GameUserResult {
	return &pbr.GameUserResult{
		UserID: ur.UserID,
		Win:    ur.Win,
		Lost:   ur.Lost,
		Tie:    ur.Tie,
		Score:  ur.Score,
	}
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
	//fmt.Printf("MarshalGameUserResult:%+v ", r.UserResults)
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
