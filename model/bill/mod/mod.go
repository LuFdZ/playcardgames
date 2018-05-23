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
	Foreign      string
	OpUserID     int64
	Channel      string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

func JournalFromProto(pJournal *pbbill.Journal) *Journal {
	out := &Journal{
		CoinType: pJournal.CoinType,
		UserID:   pJournal.UserID,
		Type:     pJournal.Type,
		OpUserID: pJournal.OpUserID,
		Channel:  pJournal.Channel,
	}
	return out
}

func (Journal *Journal) ToProto() *pbbill.Journal {
	//clubMemberNumber, _ := cacheclub.CountClubMemberHKeys(club.ClubID)
	//fmt.Printf("Club ToProto:%d\n",club.Setting)
	return &pbbill.Journal{
		ID           :Journal.ID,
		CoinType     :Journal.CoinType,
		Amount       :Journal.Amount,
		AmountBefore :Journal.AmountBefore,
		AmountAfter  :Journal.AmountAfter,
		UserID       :Journal.UserID,
		Type         :Journal.Type,
		Foreign      :Journal.Foreign,
		OpUserID     :Journal.OpUserID,
		Channel      :Journal.Channel,
		CreatedAt    :mdtime.TimeToProto(Journal.CreatedAt),
		UpdatedAt    :mdtime.TimeToProto(Journal.UpdatedAt),
	}
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
