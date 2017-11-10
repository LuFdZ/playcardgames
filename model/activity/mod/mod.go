package mod

import (
	"time"
)

type PlayerShare struct {
	UserID        int32 `gorm:"primary_key"`
	ShareTimes    int32
	TotalDiamonds int64
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}
