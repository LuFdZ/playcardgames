package mod

import (
	"encoding/json"
	"fmt"
	mdtime "playcards/model/time"
	pbr "playcards/proto/room"
	utilproto "playcards/utils/proto"
	"time"
)

type Room struct {
	RoomID    int32  `gorm:"primary_key"`
	Password  string `reg:"required,min=6,max=32,excludesall= 	"`
	Status    int32
	MaxNumber int32
	UserList  string
	GameType  int32
	CreatedAt *time.Time
	UpdatedAt *time.Time

	Users []*RoomUser `gorm:"-"`
}

type RoomUser struct {
	UserID    int32
	Ready     int32
	Position  int32
	Icon      string
	Sex       int32
	Role      int32
	UpdatedAt *time.Time
}

func (r *Room) String() string {
	return fmt.Sprintf("[roomid: %d userlist: %s status: %d gametype: %d]",
		r.RoomID, r.UserList, r.Status, r.GameType)
}

func (r *Room) ToProto() *pbr.Room {
	out := &pbr.Room{
		RoomID:    r.RoomID,
		Password:  r.Password,
		MaxNumber: r.MaxNumber,
		Status:    r.Status,
		GameType:  r.GameType,
		CreatedAt: mdtime.TimeToProto(r.CreatedAt),
		UpdatedAt: mdtime.TimeToProto(r.UpdatedAt),
	}
	utilproto.ProtoSlice(r.Users, &out.UserList)
	return out
}

func (r *RoomUser) ToProto() *pbr.RoomUser {
	return &pbr.RoomUser{
		UserID:   r.UserID,
		Ready:    r.Ready,
		Position: r.Position,
		Icon:     r.Icon,
		Sex:      r.Sex,
		Role:     r.Role,
	}
}

func (r *Room) BeforeUpdate() error {
	r.MarshalUsers()
	return nil
}

func (r *Room) BeforeCreate() error {
	r.MarshalUsers()
	return nil
}

func (r *Room) AfterFind() error {
	return r.UnmarshalUsers()
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
