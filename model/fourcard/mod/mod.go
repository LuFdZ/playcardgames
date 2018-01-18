package mod

import (
	"encoding/json"
	pbfour "playcards/proto/fourcard"
	enumfour "playcards/model/fourcard/enum"
	utilproto "playcards/utils/proto"
	"playcards/utils/tools"
	"time"

	"github.com/jinzhu/gorm"
)

type Fourcard struct {
	GameID        int32       `gorm:"primary_key"`
	RoomID        int32
	Status        int32
	Index         int32
	GameResultStr string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	OpDateAt      *time.Time
	BankerID      int32
	GameResult    *GameResult `gorm:"-"`
	PassWord      string      `gorm:"-"`
	SearchKey     string      `gorm:"-"`
	SubDateAt     *time.Time  `gorm:"-"`
	RefreshDateAt *time.Time  `gorm:"-"`
	Ids           []int32     `gorm:"-"`
}

type UserCard struct {
	Win      int32
	Score    int32
	CardType int32
	CardList []string
}

type UserInfo struct {
	UserID     int32
	Status     int32
	Bet        int32
	Role       int32
	CardList   []string
	HeadCards  *UserCard
	TailCards  *UserCard
	TotalScore int32
}

type UserDice struct {
	UserID      int32 //开始玩家id
	DiceAPoints int32 //骰子A点数
	DiceBPoints int32 //骰子B点数
}

type GameResult struct {
	UserDice *UserDice
	List     []*UserInfo
}

type RoomResultList struct {
	List []*GameResult
}

func (ud *UserDice) ToProto() *pbfour.UserDice {
	return &pbfour.UserDice{
		UserID:      ud.UserID,
		DiceAPoints: ud.DiceAPoints,
		DiceBPoints: ud.DiceBPoints,
	}
}

func (uc *UserCard) ToProto() *pbfour.UserCard {
	return &pbfour.UserCard{
		CardType: tools.IntToString(uc.CardType),
		CardList: uc.CardList,
		Score:    tools.IntToString(uc.Score),
	}
}

func (ur *UserInfo) ToProto() *pbfour.UserInfo {
	out := &pbfour.UserInfo{
		UserID:     ur.UserID,
		Status:     ur.Status,
		Bet:        enumfour.BetScoreMap[ur.Bet],
		Role:       ur.Role,
		TotalScore: tools.IntToString(ur.TotalScore),
		CardList:   ur.CardList,
	}
	if ur.HeadCards != nil && ur.TailCards != nil {
		out.HeadCards = ur.HeadCards.ToProto()
		out.TailCards = ur.TailCards.ToProto()
	}
	return out
}

func (fc *Fourcard) ToProto() *pbfour.GameResult {
	out := &pbfour.GameResult{
		RoomID:     fc.RoomID,
		GameID:     fc.GameID,
		GameStatus: fc.Status,
	}
	if fc.Status > enumfour.GameStatusAllBet {
		out.UserDice = fc.GameResult.UserDice.ToProto()
	}
	if fc.Status > enumfour.GameStatusOrdered {
		utilproto.ProtoSlice(fc.GameResult.List, &out.List)
	}

	return out
}

func (rrl *RoomResultList) ToProto() *pbfour.GameResultListReply {
	out := &pbfour.GameResultListReply{}
	utilproto.ProtoSlice(rrl.List, &out.List)
	return out
}

func (fc *Fourcard) BeforeUpdate(scope *gorm.Scope) error {
	fc.MarshalGameResult()
	scope.SetColumn("game_result_str", fc.GameResultStr)
	return nil
}

func (fc *Fourcard) BeforeCreate(scope *gorm.Scope) error {
	fc.MarshalGameResult()
	scope.SetColumn("game_result_str", fc.GameResultStr)
	return nil
}

func (fc *Fourcard) AfterFind() error {
	err := fc.UnmarshalGameResult()
	if err != nil {
		return err
	}
	return nil
}

func (fc *Fourcard) MarshalGameResult() error {
	data, _ := json.Marshal(&fc.GameResult)
	fc.GameResultStr = string(data)
	return nil
}

func (fc *Fourcard) UnmarshalGameResult() error {
	var out *GameResult
	if err := json.Unmarshal([]byte(fc.GameResultStr), &out); err != nil {
		return err
	}
	fc.GameResult = out
	return nil
}
