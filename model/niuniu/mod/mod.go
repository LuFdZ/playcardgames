package mod

import (
	"encoding/json"
	pbniu "playcards/proto/niuniu"
	mdroom "playcards/model/room/mod"
	utilproto "playcards/utils/proto"
	"playcards/utils/tools"
	"time"

	"github.com/jinzhu/gorm"
)

type Niuniu struct {
	GameID        int32                   `gorm:"primary_key"`
	RoomID        int32
	Status        int32
	Index         int32
	BankerType    int32
	GameResults   string
	Result        *NiuniuRoomResult
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	OpDateAt      *time.Time
	BankerID      int32
	PassWord      string                  `gorm:"-"`
	SearchKey     string                  `gorm:"-"`
	GetBankerList []*GetBanker            `gorm:"-"`
	BroStatus     int32                   `gorm:"-"`
	SubDateAt     *time.Time              `gorm:"-"`
	HasNewBanker  bool                    `gorm:"-"`
	RefreshDateAt *time.Time              `gorm:"-"`
	RoomType      int32                   `gorm:"-"`
	RobotOpStatus int32                   `gorm:"-"`
	Ids           []int32                 `gorm:"-"`
	RobotIds      []int32                 `gorm:"-"`
	RobotOpMap    map[int32][]int32       `gorm:"-"`
	RoomParam     *mdroom.NiuniuRoomParam `gorm:"-"`
	WatchIds      []int32                 `gorm:"-"`
}

type GetBanker struct {
	UserID int32
	Key    int32
}

type BankerAndBet struct {
	BankerScore int32
	BetScore    int32
	Role        int32
}

type UserCard struct {
	CardType string
	CardList []string
}

type NiuniuUserResult struct {
	UserID        int32
	Status        int32
	Info          *BankerAndBet
	Cards         *UserCard
	Score         int32
	Type          int32
	ClubCoinScore int64
	PushOnBet     int32
	UpdateAt      *time.Time
}

type NiuniuRoomResult struct {
	RoomID int32
	List   []*NiuniuUserResult
	//RobotIds []int32
}

type NiuniuRoomResultList struct {
	List []*NiuniuRoomResult
}

//type NiuniuRoomParam struct {
//	Times       int32
//	BankerType  int32
//	PreBankerID int32
//}

func (bab *BankerAndBet) ToProto() *pbniu.BankerAndBet {
	return &pbniu.BankerAndBet{
		BankerScore: bab.BankerScore + 1,
		BetScore:    bab.BetScore,
		Role:        bab.Role,
	}
}

func (uc *UserCard) ToProto() *pbniu.UserCard {
	return &pbniu.UserCard{
		CardType: uc.CardType,
		CardList: uc.CardList,
	}
}

func (ur *NiuniuUserResult) ToProto() *pbniu.NiuniuUserResult {
	var info *pbniu.BankerAndBet
	if ur.Info == nil {
		info = nil
	} else {
		info = ur.Info.ToProto()
	}
	return &pbniu.NiuniuUserResult{
		UserID:    ur.UserID,
		Status:    ur.Status,
		Info:      info,
		Cards:     ur.Cards.ToProto(),
		Score:     tools.IntToString(ur.Score),
		PushOnBet: ur.PushOnBet,
	}
}

func (rr *NiuniuRoomResult) ToProto() *pbniu.NiuniuRoomResult {
	out := &pbniu.NiuniuRoomResult{
		RoomID: rr.RoomID,
	}
	utilproto.ProtoSlice(rr.List, &out.List)
	return out
}

func (rrl *NiuniuRoomResultList) ToProto() *pbniu.GameResultListReply {
	out := &pbniu.GameResultListReply{}
	utilproto.ProtoSlice(rrl.List, &out.List)
	return out
}

func (gb *GetBanker) ToProto() *pbniu.GetBanker {
	return &pbniu.GetBanker{
		UserID: gb.UserID,
		Key:    gb.Key,
	}
}

func (n *Niuniu) BeforeUpdate(scope *gorm.Scope) error {
	n.MarshalNiuniuRoomResult()
	scope.SetColumn("game_results", n.GameResults)
	return nil
}

func (n *Niuniu) BeforeCreate(scope *gorm.Scope) error {
	n.MarshalNiuniuRoomResult()
	scope.SetColumn("game_results", n.GameResults)
	return nil
}

func (n *Niuniu) AfterFind() error {
	err := n.UnmarshalNiuniuRoomResult()
	if err != nil {
		return err
	}
	return nil
}

func (n *Niuniu) MarshalNiuniuRoomResult() error {
	data, _ := json.Marshal(&n.Result)
	n.GameResults = string(data)
	return nil
}

func (n *Niuniu) UnmarshalNiuniuRoomResult() error {
	var out *NiuniuRoomResult
	if err := json.Unmarshal([]byte(n.GameResults), &out); err != nil {
		return err
	}
	n.Result = out
	return nil
}
