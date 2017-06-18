package thirteen

import "time"

type thirteen struct {
	GameID    int32 `gorm:"primary_key"`
	RoomID    int32
	Status    int32
	CreatedAt *time.Time
	UpdatedAt *time.Time
}
