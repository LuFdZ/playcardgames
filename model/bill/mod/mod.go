package mod

import (
	enumbill "playcards/model/bill/enum"
	mdtime "playcards/model/time"
	pbbill "playcards/proto/bill"
	"time"
)

type Balance struct {
	ID        int32 `gorm:"primary_key"`
	UserID    int32
	CoinType  int32
	Deposit   int64
	Freeze    int64
	Amount    int64
	Balance   int64
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type Journal struct {
	ID           int32 `gorm:"primary_key"`
	CoinType     int32
	Amount       int64
	AmountBefore int64
	AmountAfter  int64
	UserID       int32
	Type         int32
	Foreign      int64
	OpUserID     int64
	Channel      string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

func (Balance) TableName() string {
	return enumbill.UserBalanceTableName
}

func (b *Balance) IsZero() bool {
	return b.Amount == 0
}

func (b *Balance) ToProto() *pbbill.Balance {
	return &pbbill.Balance{
		UserID:    b.UserID,
		CreatedAt: mdtime.TimeToProto(b.CreatedAt),
		UpdatedAt: mdtime.TimeToProto(b.UpdatedAt),
	}
}
