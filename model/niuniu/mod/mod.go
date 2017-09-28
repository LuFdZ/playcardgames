package mod

import (
	"encoding/json"
	pbniu "playcards/proto/niuniu"
	utilproto "playcards/utils/proto"
	"playcards/utils/tools"
	"time"

	"github.com/jinzhu/gorm"
)

type Niuniu struct {
	GameID        int32 `gorm:"primary_key"`
	RoomID        int32
	Status        int32
	Index         int32
	BankerType    int32
	GameResults   string
	Result        *NiuniuRoomResult
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	OpDateAt      *time.Time
	BankerID      int32        `gorm:"-"`
	GetBankerList []*GetBanker `gorm:"-"`
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
	CardType int32
	CardList []string
}

type NiuniuUserResult struct {
	UserID int32
	Status int32
	Info   *BankerAndBet
	Cards  *UserCard
	Score  int32
}

type NiuniuRoomParam struct {
	Times        int32
	BankerType   int32
	BankerHasNiu int32
	PreBankerID  int32
}

type NiuniuRoomResult struct {
	RoomID int32
	List   []*NiuniuUserResult
}

type NiuniuRoomResultList struct {
	List []*NiuniuRoomResult
}

func (bab *BankerAndBet) ToProto() *pbniu.BankerAndBet {
	return &pbniu.BankerAndBet{
		BankerScore: tools.String2int(bab.BankerScore),
		BetScore:    tools.String2int(bab.BetScore),
		Role:        tools.String2int(bab.Role),
	}
}

func (uc *UserCard) ToProto() *pbniu.UserCard {
	return &pbniu.UserCard{
		CardType: tools.String2int(uc.CardType),
		CardList: uc.CardList,
	}
}

func (ur *NiuniuUserResult) ToProto() *pbniu.NiuniuUserResult {
	return &pbniu.NiuniuUserResult{
		UserID: ur.UserID,
		Status: ur.Status,
		Info:   ur.Info.ToProto(),
		Cards:  ur.Cards.ToProto(),
		Score:  tools.String2int(ur.Score),
	}
}

func (rr *NiuniuRoomResult) ToProto() *pbniu.NiuniuRoomResult {
	out := &pbniu.NiuniuRoomResult{
		RoomID: rr.RoomID,
	}
	utilproto.ProtoSlice(rr.List, &out.List)
	return out
}

func (rrl *NiuniuRoomResultList) ToProto() *pbniu.NiuniuRoomResultList {
	out := &pbniu.NiuniuRoomResultList{}
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
	scope.SetColumn("GameResults", n.GameResults)
	return nil
}

func (n *Niuniu) BeforeCreate(scope *gorm.Scope) error {
	n.MarshalNiuniuRoomResult()
	scope.SetColumn("GameResults", n.GameResults)
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