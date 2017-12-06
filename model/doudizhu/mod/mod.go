package mod

import (
	"time"
)

type Doudizhu struct {
	GameID          int32       `gorm:"primary_key"`
	RoomID          int32
	Index           int32
	BankerType      int32
	BankerID        int32
	UserCardsRemain string
	UserCards       string
	DiZhuCards      string
	LastOpID        int32
	GameResults     string
	Status          int32
	BombTimes       int32
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
	OpDateAt        *time.Time
	GetBanker       *GetBanker  `gorm:"-"`
	SubmitCardNow   *SubmitCard `gorm:"-"`
	SubDateAt       *time.Time  `gorm:"-"`
	HasNewBanker    bool        `gorm:"-"`
	Ids             []int32     `gorm:"-"`
}

type GetBanker struct {
	UserID int32
	Type   int32
}

type SubmitCard struct {
	Index    int32
	UserID   int32
	CardType int32
	CardList []string
	NextId   int32
}

type UserCard struct {
	UserCard   []string
	CardRemain []string

}
}
