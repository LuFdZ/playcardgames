package mod

import (
	"fmt"
	mdtime "playcards/model/time"
	pbr "playcards/proto/room"
	"strings"
	"time"
)

type Room struct {
	RoomID       int32  `gorm:"primary_key"`
	Password     string `reg:"required,min=6,max=32,excludesall= 	"`
	UserList     string
	PlayerMaxNum int32
	Status       int32
	GameType     int32
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

func (r *Room) String() string {
	return fmt.Sprintf("[rid: %d,uidLidt: %d,status: %d,gametype: %d]",
		r.RoomID, r.UserList, r.Status, r.GameType)
}

func (r *Room) ToProto() *pbr.Room {
	return &pbr.Room{
		RoomID:       r.RoomID,
		Password:     r.Password,
		UserList:     strings.Split(r.UserList, ","),
		PlayerMaxNum: r.PlayerMaxNum,
		Status:       r.Status,
		GameType:     r.GameType,
		CreatedAt:    mdtime.TimeToProto(r.CreatedAt),
		UpdatedAt:    mdtime.TimeToProto(r.UpdatedAt),
	}
}
