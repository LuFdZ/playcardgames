package mod

import (
	"encoding/json"
	pbtow "playcards/proto/twocard"
	enumtow "playcards/model/twocard/enum"
	utilproto "playcards/utils/proto"
	"playcards/utils/tools"
	"time"

	"github.com/jinzhu/gorm"
)

type Twocard struct {
	GameID        int32       `gorm:"primary_key"`
	RoomID        int32
	Status        int32
	Index         int32
	GameResultStr string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	OpDateAt      *time.Time
	BankerID      int32
	BetType       int32       `gorm:"-"`
	GameResult    *GameResult `gorm:"-"`
	PassWord      string      `gorm:"-"`
	SearchKey     string      `gorm:"-"`
	SubDateAt     *time.Time  `gorm:"-"`
	RefreshDateAt *time.Time  `gorm:"-"`
	Ids           []int32     `gorm:"-"`
	WatchIds        []int32         `gorm:"-"`
}

type UserCard struct {
	Win      int32 //输赢分
	Score    int32 //牌组分值
	CardType int32 //牌型值
	Value    int32 //比较值 牌组为牌型值*100 散牌为牌型值*10+牌组值
	CardList []string
}

type UserInfo struct {
	UserID        int32
	Status        int32
	Bet           int32
	Role          int32
	Type          int32
	CardList      []string
	Cards         *UserCard
	TotalScore    int32
	ClubCoinScore int64
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

func (ud *UserDice) ToProto() *pbtow.UserDice {
	return &pbtow.UserDice{
		UserID:      ud.UserID,
		DiceAPoints: ud.DiceAPoints,
		DiceBPoints: ud.DiceBPoints,
	}
}

func (uc *UserCard) ToProto() *pbtow.UserCard {
	return &pbtow.UserCard{
		CardType: tools.IntToString(uc.CardType),
		CardList: uc.CardList,
		Value:    tools.IntToString(uc.Score),
		ScoreWin: tools.IntToString(uc.Win),
	}
}

func (ur *UserInfo) ToProto() *pbtow.UserInfo {
	out := &pbtow.UserInfo{
		UserID:        ur.UserID,
		Status:        ur.Status,
		Bet:           ur.Bet, //enumfour.BetScoreMap[ur.Bet],
		Role:          ur.Role,
		TotalScore:    tools.IntToString(ur.TotalScore),
		CardList:      ur.CardList,
		ClubCoinScore: ur.ClubCoinScore,
	}
	if ur.Cards != nil {
		out.Cards = ur.Cards.ToProto()
	}
	return out
}

func (tc *Twocard) ToProto() *pbtow.GameResult {
	out := &pbtow.GameResult{
		RoomID:     tc.RoomID,
		GameID:     tc.GameID,
		GameStatus: tc.Status,
	}
	if tc.Status > enumtow.GameStatusAllBet {
		out.UserDice = tc.GameResult.UserDice.ToProto()
	}
	utilproto.ProtoSlice(tc.GameResult.List, &out.List)
	return out
}

func (rrl *RoomResultList) ToProto() *pbtow.GameResultListReply {
	out := &pbtow.GameResultListReply{}
	utilproto.ProtoSlice(rrl.List, &out.List)
	return out
}

func (tc *Twocard) BeforeUpdate(scope *gorm.Scope) error {
	tc.MarshalGameResult()
	scope.SetColumn("game_result_str", tc.GameResultStr)
	return nil
}

func (tc *Twocard) BeforeCreate(scope *gorm.Scope) error {
	tc.MarshalGameResult()
	scope.SetColumn("game_result_str", tc.GameResultStr)
	return nil
}

func (tc *Twocard) AfterFind() error {
	err := tc.UnmarshalGameResult()
	if err != nil {
		return err
	}
	return nil
}

func (tc *Twocard) MarshalGameResult() error {
	data, _ := json.Marshal(&tc.GameResult)
	tc.GameResultStr = string(data)
	return nil
}

func (tc *Twocard) UnmarshalGameResult() error {
	var out *GameResult
	if err := json.Unmarshal([]byte(tc.GameResultStr), &out); err != nil {
		return err
	}
	tc.GameResult = out
	return nil
}
