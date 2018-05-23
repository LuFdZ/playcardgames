package mod

import (
	"encoding/json"
	pbrun "playcards/proto/runcard"
	utilproto "playcards/utils/proto"
	"playcards/utils/tools"
	"time"

	"github.com/jinzhu/gorm"
)

type Runcard struct {
	GameID         int32       `gorm:"primary_key"`
	RoomID         int32
	Status         int32
	Index          int32
	GameResultStr  string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	OpDateAt       *time.Time
	FirstCardType  string      `gorm:"-"`
	RoomParam      [][]string  `gorm:"-"`
	LastSubmitCard *SubmitCard `gorm:"-"`
	OpUserID       int32       `gorm:"-"`
	GameResult     *GameResult `gorm:"-"`
	PassWord       string      `gorm:"-"`
	SearchKey      string      `gorm:"-"`
	SubDateAt      *time.Time  `gorm:"-"`
	RefreshDateAt  *time.Time  `gorm:"-"`
	SubmitIndex    int32       `gorm:"-"`
	//Ids           []int32     `gorm:"-"`
	//WatchIds      []int32     `gorm:"-"`
}

type SubmitCard struct {
	UserID     int32
	NextUserID int32
	CardType   string
	CardList   []string
	Result     string //比较结果
}

type UserInfo struct {
	UserID        int32
	Status        int32
	CardNum       int32
	CardList      []string
	LastCards     *SubmitCard
	Score         int32
	BoomNum       int32
	BoomScore     int32
	ClubCoinScore int64
	Index         int32
	HasSubmitCard int32
	IsSpring      int32 //是否春天 :1是，2否， 3关门
}

type GameResult struct {
	List        []*UserInfo
	IsSpring    int32 //是否春天 :1是，2否
	SettleTimes int32 //结算倍数
}

type RoomResultList struct {
	List []*GameResult
}

func (rc *Runcard) GetUserInfo(uid int32) *UserInfo {
	for _, ui := range rc.GameResult.List {
		if ui.UserID == uid {
			return ui
		}
	}
	return nil
}

func (rc *Runcard) GetUserInfoWithIndex(uid int32) int {
	for i, ui := range rc.GameResult.List {
		if ui.UserID == uid {
			return i
		}
	}
	return 0
}

func (rc *Runcard) GetUserMapByIndex(uid int32) map[int32][]string {
	m := make(map[int32][]string)
	for _, ur := range rc.GameResult.List {
		if ur.UserID != uid {
			m[ur.UserID] = ur.CardList
		}
	}
	return m
}

func (rc *Runcard) GetNextUserInfo(index int32) *UserInfo {
	if index == int32(len(rc.GameResult.List)) {
		index = 1
	} else {
		index += 1
	}
	for _, ui := range rc.GameResult.List {
		if ui.Index == index {
			return ui
		}
	}
	return nil
}

func (rc *Runcard) GetUserInfoListByID(uid int32) []*pbrun.UserInfo {
	var uis []*pbrun.UserInfo
	for _, ui := range rc.GameResult.List {
		pbui := &pbrun.UserInfo{
			UserID:        ui.UserID,
			Status:        ui.Status,
			CardNum:       ui.CardNum,
			Score:         tools.IntToString(ui.Score),
			BoomNum:       ui.BoomNum,
			BoomScore:     tools.IntToString(ui.BoomScore),
			ClubCoinScore: tools.Int64ToString(ui.ClubCoinScore),
		}
		if ui.UserID == uid {
			pbui.CardList = ui.CardList
		}
		uis = append(uis, pbui)
	}
	return uis
}

func (sc *SubmitCard) ToProto() *pbrun.SubmitCard {
	return &pbrun.SubmitCard{
		CurUserID:  sc.UserID,
		NextUserID: sc.NextUserID,
		CardType:   sc.CardType,
		CardList:   sc.CardList,
	}
}

func (ui *UserInfo) ToProto() *pbrun.UserInfo {
	out := &pbrun.UserInfo{
		UserID:        ui.UserID,
		Status:        ui.Status,
		CardNum:       ui.CardNum,
		CardList:      ui.CardList,
		Score:         tools.IntToString(ui.Score),
		BoomNum:       ui.BoomNum,
		ClubCoinScore: tools.Int64ToString(ui.ClubCoinScore),
		BoomScore:     tools.IntToString(ui.BoomScore),
		Index:         ui.Index,
	}
	if ui.LastCards != nil {
		out.LastCards = ui.LastCards.ToProto()
	}
	return out
}

func (rc *Runcard) ToProto() *pbrun.GameResult {
	out := &pbrun.GameResult{
		RoomID:     rc.RoomID,
		GameID:     rc.GameID,
		GameStatus: rc.Status,
		OpUserID:   rc.OpUserID,
	}
	utilproto.ProtoSlice(rc.GameResult.List, &out.List)
	return out
}

func (rrl *RoomResultList) ToProto() *pbrun.GameResultListReply {
	out := &pbrun.GameResultListReply{}
	utilproto.ProtoSlice(rrl.List, &out.List)
	return out
}

func (rc *Runcard) BeforeUpdate(scope *gorm.Scope) error {
	rc.MarshalGameResult()
	scope.SetColumn("game_result_str", rc.GameResultStr)
	return nil
}

func (rc *Runcard) BeforeCreate(scope *gorm.Scope) error {
	rc.MarshalGameResult()
	scope.SetColumn("game_result_str", rc.GameResultStr)
	return nil
}

func (rc *Runcard) AfterFind() error {
	err := rc.UnmarshalGameResult()
	if err != nil {
		return err
	}
	return nil
}

func (rc *Runcard) MarshalGameResult() error {
	data, err := json.Marshal(&rc.GameResult)
	if err != nil {
		return err
	}
	rc.GameResultStr = string(data)
	return nil
}

func (rc *Runcard) UnmarshalGameResult() error {
	var out *GameResult
	if err := json.Unmarshal([]byte(rc.GameResultStr), &out); err != nil {
		return err
	}
	rc.GameResult = out
	return nil
}
