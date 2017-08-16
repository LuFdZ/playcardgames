package mod

import (
	enumbill "playcards/model/bill/enum"
	mdtime "playcards/model/time"
	pbbill "playcards/proto/bill"
	"time"
)

type Balance struct {
	Amount  int64
	Gold    int64
	Diamond int64
}

type UserBalance struct {
	Balance
	UserID                int32 `gorm:"primary_key"`
	Deposit               int64
	Freeze                int64
	GoldProfit            int64
	DiamondProfit         int64
	CumulativeRecharge    int64
	CumulativeGift        int64
	CumulativeConsumption int64
	CreatedAt             *time.Time
	UpdatedAt             *time.Time
}

type Journal struct {
	Balance
	ID        int64 `gorm:"primary_key"`
	UserID    int32
	Type      int32
	Foreign   string
	OpUserID  int64
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type Deposit struct {
	ID        int32 `gorm:"primary_key"`
	UserID    int32
	Amount    int64
	Type      int32
	CreatedAt *time.Time
}

func (UserBalance) TableName() string {
	return enumbill.UserBalanceTableName
}

func (b *Balance) IsZero() bool {
	return b.Gold == 0 && b.Diamond == 0 && b.Amount == 0
}

func (b *UserBalance) ToProto() *pbbill.Balance {
	return &pbbill.Balance{
		UserID:                b.UserID,
		Amount:                b.Amount,
		Gold:                  b.Gold,
		Diamond:               b.Diamond,
		Deposit:               b.Deposit,
		Freeze:                b.Freeze,
		GoldProfit:            b.GoldProfit,
		DiamondProfit:         b.DiamondProfit,
		CreatedAt:             mdtime.TimeToProto(b.CreatedAt),
		UpdatedAt:             mdtime.TimeToProto(b.UpdatedAt),
		CumulativeRecharge:    b.CumulativeRecharge,
		CumulativeGift:        b.CumulativeGift,
		CumulativeConsumption: b.CumulativeConsumption,
	}
}
