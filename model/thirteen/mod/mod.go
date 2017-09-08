package thirteen

import (
	"encoding/json"
	pbt "playcards/proto/thirteen"
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
	SubmitCards     []*SubmitCard   `gorm:"-"`
	Cards           []*GroupCard    `gorm:"-"`
	GameLua         *lua.LState     `gorm:"-"`
	Result          *GameResultList `gorm:"-"`
}

type ThirteenUserLog struct {
	LogID      int32 `gorm:"primary_key"`
	GameID     int32
	UserID     int32
	RoomID     int32
	GameResult string
	Status     int32
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

type Card struct {
	Type  int32
	Value int32
}

// type GroupCardList struct {
// 	List []*GroupCard
// }

// type CardList struct {
// 	List []*Card
// }

type Settle struct {
	UserID      int32
	ScoreHead   int32
	ScoreMiddle int32
	ScoreTail   int32
	Score       int32
}

type GameResult struct {
	UserID     int32
	SettleList []*Settle
	CardsList  *SubmitCard
	CardType   string
	//NullUserCards CardList
}

type GameResultList struct {
	RoomID  int32
	Results []*GameResult
}

type GameParam struct {
	Joke           int32
	Time           int32
	Times          int32
	BankerAddScore int32
}

type GroupCard struct {
	UserID   int32
	Type     int32
	Weight   int32
	CardList []string
}

type SubmitCard struct {
	UserID int32
	Head   []string
	Middle []string
	Tail   []string
}

func SubmitCardFromProto(sc *pbt.SubmitCard) *SubmitCard {
	return &SubmitCard{
		Head:   sc.Head,
		Middle: sc.Middle,
		Tail:   sc.Tail,
	}
}

func (s *Settle) ToProto() *pbt.Settle {
	return &pbt.Settle{
		UserID:      s.UserID,
		ScoreHead:   s.ScoreHead,
		ScoreMiddle: s.ScoreMiddle,
		ScoreTail:   s.ScoreTail,
		Score:       s.Score,
	}
}

func (sc *SubmitCard) ToProto() *pbt.SubmitCard {
	return &pbt.SubmitCard{
		Head:   sc.Head,
		Middle: sc.Middle,
		Tail:   sc.Tail,
	}
}

func (gr *GameResult) ToProto() *pbt.GameResult {
	var settleList []*pbt.Settle
	for _, settle := range gr.SettleList {
		settleList = append(settleList, settle.ToProto())
	}

	return &pbt.GameResult{
		UserID:     gr.UserID,
		SettleList: settleList,
		CardsList:  gr.CardsList.ToProto(),
		CardType:   gr.CardType,
	}
}

func (grl *GameResultList) ToProto() *pbt.GameResultList {
	var results []*pbt.GameResult
	for _, gr := range grl.Results {
		results = append(results, gr.ToProto())
	}

	out := &pbt.GameResultList{
		RoomID:  grl.RoomID,
		Results: results,
	}
	return out
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
		UserID:   gc.UserID,
		Type:     gc.Type,
		Weight:   gc.Weight,
		CardList: gc.CardList,
	}
	//utilproto.ProtoSlice(gc.CardList, &out.CardList)
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
