package mod

import (
	pbddz "playcards/proto/doudizhu"
	"time"
	"github.com/jinzhu/gorm"
	"encoding/json"
)

type Doudizhu struct {
	GameID          int32 `gorm:"primary_key"`
	RoomID          int32
	Index           int32
	BankerID        int32
	DizhuCardStr    string
	UserCardInfoStr string
	GameResultStr   string
	GetBankerLogStr string
	GameCardLogStr  string
	RestartTimes    int32
	Status          int32
	BaseScore       int32
	BankerTimes     int32
	BombTimes       int32
	OpIndex         int32
	WinerID         int32
	WinerType       int32
	Spring          int32
	CommonBomb      int32
	RocketBomb      int32
	EightBomb       int32
	OpDateAt        *time.Time
	CreatedAt       *time.Time
	UpdatedAt       *time.Time

	GoldGame            int32         `gorm:"-"`
	BankerStatus        int32         `gorm:"-"`
	PassWord            string        `gorm:"-"`
	SearchKey           string        `gorm:"-"`
	OpTotalIndex        int32         `gorm:"-"`
	BankerType          int32         `gorm:"-"`
	UserCardInfoList    []*UserCard   `gorm:"-"`
	DizhuCardList       []string      `gorm:"-"`
	GameResultList      []*UserResult `gorm:"-"`
	GameCardLogList     []string      `gorm:"-"`
	GetBankerLogList    []*GetBanker  `gorm:"-"`
	GetBankerLogStrList []string      `gorm:"-"`
	SubmitCardNow       *SubmitCard   `gorm:"-"`
	SubDateAt           *time.Time    `gorm:"-"`
	OpID                int32         `gorm:"-"`
	Ids                 []int32       `gorm:"-"`
}

type GetBanker struct {
	Index   int32
	UserID  int32
	Type    int32
	OpTimes int32
}

type SubmitCard struct {
	UserID    int32
	CardType  int32
	CardList  []string
	BombTimes int32
	NextID    int32
}

type UserCard struct {
	UserID     int32
	CardLast   []string
	CardList   []string
	CardRemain []string
}

type UserResult struct {
	UserID        int32
	Score         int32
	GameCardCount string
}

type GameInit struct {
	DiZhuCardList []string
	UserCardList  []*UserCard
	GetBankerList []*GetBanker
	BaseScore     int32
	BankerTimes   int32
}

type GameRecovery struct {
	GameID            int32
	RoomID            int32
	LastGetBankerID   int32
	LastGetBankerType int32
	UserCardList      []string
	DizhuCardList     []string
}

func (ur *UserResult) ToProto() *pbddz.UserInfo {
	return &pbddz.UserInfo{
		UserID: ur.UserID,
		Score:  ur.Score,
	}
}

func (ddz *Doudizhu) ResultToProto() *pbddz.GameResultBro {
	var urls []*pbddz.UserInfo
	ur := &pbddz.GameResult{}
	for i, uci := range ddz.UserCardInfoList {
		ur := &pbddz.UserInfo{
			UserID:     uci.UserID,
			Score:      ddz.GameResultList[i].Score,
			CardRemain: uci.CardRemain,
		}
		urls = append(urls, ur)
	}
	ur.GameID = ddz.GameID
	ur.BaseScore = ddz.BaseScore
	ur.BombTimes = ddz.BankerTimes * ddz.BombTimes
	ur.UserResult = urls
	urb := &pbddz.GameResultBro{
		Content: ur,
		Ids:     ddz.Ids,
	}
	return urb
}

func (ddz *Doudizhu) GetUserCard(uid int32) *UserCard {
	for _, uci := range ddz.UserCardInfoList {
		if uci.UserID == uid {
			return uci
		}
	}
	return nil
}

func (ddz *Doudizhu) BeforeUpdate(scope *gorm.Scope) error {
	ddz.MarshalDoudizhuResult()
	ddz.MarshalGetBankerLog()
	ddz.MarshalDizhuCard()
	ddz.MarshalUserCardInfo()
	ddz.MarshalGameCardLogStr()
	scope.SetColumn("dizhu_card_str", ddz.DizhuCardStr)
	scope.SetColumn("user_card_info_str", ddz.UserCardInfoStr)
	scope.SetColumn("game_result_str", ddz.GameResultStr)
	scope.SetColumn("game_card_log_str", ddz.GameCardLogStr)
	scope.SetColumn("get_banker_log_str", ddz.GetBankerLogStr)
	return nil
}

func (ddz *Doudizhu) BeforeCreate(scope *gorm.Scope) error {
	ddz.MarshalDoudizhuResult()
	ddz.MarshalGetBankerLog()
	ddz.MarshalDizhuCard()
	ddz.MarshalUserCardInfo()
	ddz.MarshalGameCardLogStr()
	scope.SetColumn("dizhu_card_str", ddz.DizhuCardStr)
	scope.SetColumn("user_card_info_str", ddz.UserCardInfoStr)
	scope.SetColumn("game_result_str", ddz.GameResultStr)
	scope.SetColumn("game_card_log_str", ddz.GameCardLogStr)
	scope.SetColumn("get_banker_log_str", ddz.GetBankerLogStr)
	return nil
}

func (ddz *Doudizhu) AfterFind() error {
	err := ddz.UnmarshalDoudizhuResult()
	if err != nil {
		return err
	}
	err = ddz.UnmarshalDizhuCard()
	if err != nil {
		return err
	}
	err = ddz.UnmarshalUserCardInfo()
	if err != nil {
		return err
	}
	err = ddz.UnmarshalGetBankerLog()
	if err != nil {
		return err
	}
	err = ddz.UnmarshalGameCardLogStr()
	if err != nil {
		return err
	}
	return nil
}

func (ddz *Doudizhu) MarshalDoudizhuResult() error {
	data, _ := json.Marshal(&ddz.GameResultList)
	ddz.GameResultStr = string(data)
	return nil
}

func (ddz *Doudizhu) UnmarshalDoudizhuResult() error {
	var out []*UserResult
	if err := json.Unmarshal([]byte(ddz.GameResultStr), &out); err != nil {
		return err
	}
	ddz.GameResultList = out
	return nil
}

func (ddz *Doudizhu) MarshalGetBankerLog() error {
	data, _ := json.Marshal(&ddz.GetBankerLogStrList)
	ddz.GetBankerLogStr = string(data)
	return nil
}

func (ddz *Doudizhu) UnmarshalGetBankerLog() error {
	var out []string
	if err := json.Unmarshal([]byte(ddz.GetBankerLogStr), &out); err != nil {
		return err
	}
	ddz.GetBankerLogStrList = out
	return nil
}

func (ddz *Doudizhu) MarshalDizhuCard() error {
	data, _ := json.Marshal(&ddz.DizhuCardList)
	ddz.DizhuCardStr = string(data)
	return nil
}

func (ddz *Doudizhu) UnmarshalDizhuCard() error {
	var out []string
	if err := json.Unmarshal([]byte(ddz.DizhuCardStr), &out); err != nil {
		return err
	}
	ddz.DizhuCardList = out
	return nil
}

func (ddz *Doudizhu) MarshalUserCardInfo() error {
	data, _ := json.Marshal(&ddz.UserCardInfoList)
	ddz.UserCardInfoStr = string(data)
	return nil
}

func (ddz *Doudizhu) UnmarshalUserCardInfo() error {
	var out []*UserCard
	if err := json.Unmarshal([]byte(ddz.UserCardInfoStr), &out); err != nil {
		return err
	}
	ddz.UserCardInfoList = out
	return nil
}

func (ddz *Doudizhu) MarshalGameCardLogStr() error {
	data, _ := json.Marshal(&ddz.GameCardLogList)
	ddz.GameCardLogStr = string(data)
	return nil
}

func (ddz *Doudizhu) UnmarshalGameCardLogStr() error {
	var out []string
	if err := json.Unmarshal([]byte(ddz.GameCardLogStr), &out); err != nil {
		return err
	}
	ddz.GameCardLogList = out
	return nil
}
