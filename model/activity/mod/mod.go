package mod

import (
	pba "playcards/proto/activity"
	"time"
)

type ActivityConfig struct {
	ConfigID         int32 `gorm:"primary_key"`
	Description      string
	Parameter        string
	LastModifyUserID int32
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
}

func (a *ActivityConfig) ToProto() *pba.ActivityConfig {
	return &pba.ActivityConfig{
		ConfigID:         a.ConfigID,
		Description:      a.Description,
		Parameter:        a.Parameter,
		LastModifyUserID: a.LastModifyUserID,
	}
}

func ActivityConfigFormProto(ac *pba.ActivityConfig) *ActivityConfig {
	return &ActivityConfig{
		ConfigID:    ac.ConfigID,
		Description: ac.Description,
		Parameter:   ac.Parameter,
	}
}
