package thirteen

import (
	pbt "playcards/proto/thirteen"
	"time"
)

type Thirteen struct {
	GameID        int32 `gorm:"primary_key"`
	RoomID        int32
	Status        int32
	UserCardList  string
	UserScoreList string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

type ThirteenUserLog struct {
	GameID       int32
	UserID       int32
	RoomID       int32
	UserCardList string
	score        int32
	status       int32
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
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
	out := &CardList{
		List: nil,
	}
	utilproto.ProtoSlice(cl.List, &out.List)
	return out
}
