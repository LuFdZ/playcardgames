package thirteen

import (
	"encoding/json"
	"fmt"
	pbt "playcards/proto/thirteen"
	utilproto "playcards/utils/proto"
	"time"

	"github.com/jinzhu/gorm"
	lua "github.com/yuin/gopher-lua"
)

type Thirteen struct {
	GameID          int32 `gorm:"primary_key"`
	RoomID          int32
	Status          int32
	Index           int32
	UserCards       string
	UserSubmitCards string
	GameResults     string
	CreatedAt       *time.Time
	UpdatedAt       *time.Time

	SubmitCards *UserSubmitCards `gorm:"-"`
	Cards       *GroupCardList   `gorm:"-"`
	GameLua     *lua.LState      `gorm:"-"`
	Result      *GameResult      `gorm:"-"`
}

type ThirteenUserLog struct {
	LogID      int32 `gorm:"primary_key"`
	GameID     int32
	UserID     int32
	RoomID     int32
	GameResult string
	status     int32
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

type Card struct {
	Type  int32
	Value int32
}

type GroupCard struct {
	UserID   int32
	Type     int32
	Weight   int32
	CardList []*Card
}

type UserSubmitCards struct {
	SubmitCardList []*UserSubmitCard
}
type UserSubmitCard struct {
	UserID     int32
	HeadList   []*Card
	MiddkeList []*Card
	TailList   []*Card
}

type GroupCardList struct {
	List []*GroupCard
}

type CardList struct {
	List []*Card
}

type Settle struct {
	UserID      int32
	ScoreHead   int32
	ScoreMiddke int32
	ScoreTail   int32
	Score       int32
}

type GameResult struct {
	List          []*Settle
	GroupCards    GroupCardList
	NullUserCards CardList
}

func (c *Card) ToProto() *pbt.Card {
	out := &pbt.Card{
		Type:  c.Type,
		Value: c.Value,
	}
	return out
}

func (cl *CardList) ToProto() *pbt.CardList {
	out := &pbt.CardList{
		List: nil,
	}
	utilproto.ProtoSlice(cl.List, &out.List)
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
	t.MarshalUserSubmitCards()
	//fmt.Printf("BeforeUpdate user_cards:%s|user_submit_cards:%s |", t.UserCards, t.UserCards)
	scope.SetColumn("user_cards", t.UserCards)
	scope.SetColumn("user_submit_cards", t.UserSubmitCards)
	return nil
}

func (t *Thirteen) AfterFind() error {
	return t.UnmarshalUserCards()
}

func (t *Thirteen) MarshaUserCards() error {
	data, _ := json.Marshal(&t.Cards)
	t.UserCards = string(data)
	fmt.Printf("MarshaUserCards: %s", t.UserCards)
	return nil
}

func (t *Thirteen) MarshalUserSubmitCards() error {
	data, _ := json.Marshal(&t.SubmitCards)
	t.UserSubmitCards = string(data)
	return nil
}

func (t *Thirteen) MarshalGameResults() error {
	data, _ := json.Marshal(&t.Result)
	t.GameResults = string(data)
	return nil
}

func (t *Thirteen) UnmarshalUserCards() error {
	var out *GroupCardList
	if err := json.Unmarshal([]byte(t.UserCards), &out); err != nil {
		return err
	}
	t.Cards = out
	return nil
}

func (t *Thirteen) UnmarshalUserSubmitCards() error {
	var out *UserSubmitCards
	if err := json.Unmarshal([]byte(t.UserSubmitCards), &out); err != nil {
		return err
	}
	t.SubmitCards = out
	return nil
}
