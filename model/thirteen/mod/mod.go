package thirteen

import (
	"encoding/json"
	pbt "playcards/proto/thirteen"
	utilproto "playcards/utils/proto"
	"time"

	"github.com/jinzhu/gorm"
	lua "github.com/yuin/gopher-lua"
)

type Thirteen struct {
	GameID          int32           `gorm:"primary_key"`
	RoomID          int32
	BankerID        int32
	Status          int32
	Index           int32
	UserCards       string
	UserSubmitCards string
	GameResults     string
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
	PassWord        string          `gorm:"-"`
	SubmitCards     []*SubmitCard   `gorm:"-"`
	Cards           []*GroupCard    `gorm:"-"`
	GameLua         *lua.LState     `gorm:"-"`
	Result          *GameResultList `gorm:"-"`
	SearchKey       string          `gorm:"-"`
	Ids             []int32         `gorm:"-"`
	WatchIds        []int32         `gorm:"-"`
}

type Card struct {
	Type  int32
	Value int32
}

type ThirteenSettle struct {
	Head       string
	Middle     string
	Tail       string
	AddScore   string
	TotalScore string
}

type ThirteenResult struct {
	UserID        int32
	Role          int32
	Settle        ThirteenSettle
	Result        ThirteenGroupResult
	ClubCoinScore int64
}

type ThirteenGroupResult struct {
	Head   ResGroup
	Middle ResGroup
	Tail   ResGroup
	Shoot  []int32
}

type ResGroup struct {
	GroupType string
	CardList  []string
}

type GameParam struct {
	Joke           int32
	Time           int32
	Times          int32
	BankerAddScore int32
}

type GroupCard struct {
	UserID     int32
	CardList   []string
	RoomStatus int32
}

type SubmitCard struct {
	UserID int32
	Role   int32
	Head   []string
	Middle []string
	Tail   []string
}

type GameResultList struct {
	Result []*ThirteenResult
}

type ThirteenRecovery struct {
	Result     int32
	Status     int32
	BankerID   int32
	Cards      GroupCard
	ReadyUser  []int32
	GameResult GameResultList
}

//type ThirteenRoomParam struct {
//	BankerAddScore int32
//	BankerType     int32
//	Time           int32
//	Joke           int32
//	Times          int32
//}

func SubmitCardFromProto(sc *pbt.SubmitCard, uid int32) *SubmitCard {
	return &SubmitCard{
		UserID: uid,
		Head:   sc.Head,
		Middle: sc.Middle,
		Tail:   sc.Tail,
	}
}

func (rg *ResGroup) ToProto() *pbt.ResGroup {
	return &pbt.ResGroup{
		GroupType: rg.GroupType,
		CardList:  rg.CardList,
	}
}

func (tgr *ThirteenGroupResult) ToProto() *pbt.ThirteenGroupResult {
	return &pbt.ThirteenGroupResult{
		Head:   tgr.Head.ToProto(),
		Middle: tgr.Middle.ToProto(),
		Tail:   tgr.Tail.ToProto(),
		Shoot:  tgr.Shoot,
	}
}

func (ts *ThirteenSettle) ToProto() *pbt.ThirteenSettle {
	return &pbt.ThirteenSettle{
		Head:       ts.Head,
		Middle:     ts.Middle,
		Tail:       ts.Tail,
		AddScore:   ts.AddScore,
		TotalScore: ts.TotalScore,
	}
}

func (sc *SubmitCard) ToProto() *pbt.SubmitCard {
	return &pbt.SubmitCard{
		Head:   sc.Head,
		Middle: sc.Middle,
		Tail:   sc.Tail,
	}
}

func (ts *ThirteenResult) ToProto() *pbt.ThirteenResult {
	return &pbt.ThirteenResult{
		UserID:        ts.UserID,
		Settle:        ts.Settle.ToProto(),
		Result:        ts.Result.ToProto(),
		Role:          ts.Role,
		ClubCoinScore: ts.ClubCoinScore,
	}
}

func (grl *GameResultList) ToProto() *pbt.GameResultList {
	var results []*pbt.ThirteenResult
	for _, gr := range grl.Result {
		results = append(results, gr.ToProto())
	}

	out := &pbt.GameResultList{
		Result: results,
	}
	return out
}

func (tr *ThirteenRecovery) ToProto() *pbt.GameRecovery {
	return &pbt.GameRecovery{
		Result:     tr.Result,
		Status:     tr.Status,
		Cards:      tr.Cards.ToProto(),
		GameResult: tr.GameResult.ToProto(),
		ReadyUser:  tr.ReadyUser,
	}
}

// func (c *Card) ToProto() *pbt.Card {
// 	out := &pbt.Card{
// 		Type:  c.Type,
// 		Value: c.Value,
// 	}
// 	return out
// }

func (gc *GroupCard) ToProto() *pbt.GroupCard {
	out := &pbt.GroupCard{
		UserID:     gc.UserID,
		CardList:   gc.CardList,
		RoomStatus: gc.RoomStatus,
	}
	utilproto.ProtoSlice(gc.CardList, &out.CardList)
	//fmt.Printf("GroupCard ToProto %v", out.CardList)
	return out
}

func (t *Thirteen) BeforeUpdate(scope *gorm.Scope) error {
	//t.MarshaUserCards()
	t.MarshalUserSubmitCards()
	t.MarshalGameResults()

	//scope.SetColumn("user_cards", t.UserCards)
	scope.SetColumn("user_submit_cards", t.UserSubmitCards)
	scope.SetColumn("game_results", t.GameResults)
	return nil
}

func (t *Thirteen) BeforeCreate(scope *gorm.Scope) error {
	//fmt.Printf("AAAAAAA ", checkpwd)
	t.MarshaUserCards()
	//t.MarshalUserSubmitCards()
	//fmt.Printf("BeforeUpdate user_cards:%s|user_submit_cards:%s |", t.UserCards, t.UserCards)
	scope.SetColumn("user_cards", t.UserCards)
	//scope.SetColumn("user_submit_cards", t.UserSubmitCards)
	return nil
}

func (t *Thirteen) AfterFind() error {
	err := t.UnmarshalGameResults()
	if err != nil {
		return err
	}
	err = t.UnmarshalUserSubmitCards()
	if err != nil {
		return err
	}
	err = t.UnmarshalUserCards()
	if err != nil {
		return err
	}
	return nil
}

func (t *Thirteen) MarshaUserCards() error {
	data, _ := json.Marshal(&t.Cards)
	t.UserCards = string(data)
	return nil
}

func (t *Thirteen) MarshalUserSubmitCards() error {
	//fmt.Printf(" MarshalUserSubmitCards:%+v",t.SubmitCards)
	data, _ := json.Marshal(&t.SubmitCards)
	t.UserSubmitCards = string(data)
	return nil
}

func (t *Thirteen) MarshalGameResults() error {
	data, _ := json.Marshal(&t.Result)
	t.GameResults = string(data)
	//fmt.Printf("MarshalGameResults:%+v ", t.Result)
	return nil
}

func (t *Thirteen) UnmarshalUserCards() error {
	var out []*GroupCard
	if err := json.Unmarshal([]byte(t.UserCards), &out); err != nil {
		return err
	}
	t.Cards = out
	return nil
}

func (t *Thirteen) UnmarshalUserSubmitCards() error {
	var out []*SubmitCard
	if len(t.UserSubmitCards) > 0 {
		if err := json.Unmarshal([]byte(t.UserSubmitCards), &out); err != nil {
			return err
		}
	}
	t.SubmitCards = out
	return nil
}

func (t *Thirteen) UnmarshalGameResults() error {
	if len(t.GameResults) == 0 {
		return nil
	}
	var out GameResultList
	if err := json.Unmarshal([]byte(t.GameResults), &out); err != nil {
		return err
	}
	t.Result = &out
	return nil
}
