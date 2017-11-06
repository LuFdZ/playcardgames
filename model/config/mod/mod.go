package mod

import (
	pbconf "playcards/proto/config"
	mdtime "playcards/model/time"
	"time"
)

type Config struct {
	ConfigID      int32 `gorm:"primary_key"`
	Channel     string
	Version     string
	MobileOs    string
	ItemID      int32
	ItemValue   string
	Status      int32
	Description string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}


func (co *Config) ToProto() *pbconf.Config {
	return &pbconf.Config{
		ConfigID:      co.ConfigID,
		Channel:     co.Channel,
		Version:     co.Version,
		MobileOs:    co.MobileOs,
		ItemID:      co.ItemID,
		ItemValue:   co.ItemValue,
		Status:      co.Status,
		Description: co.Description,
		CreatedAt:     mdtime.TimeToProto(co.CreatedAt),
		UpdatedAt:     mdtime.TimeToProto(co.UpdatedAt),
	}
}

func ConfigFromProto(co *pbconf.Config) *Config {
	if co == nil{
		return &Config{}
	}
	return &Config{
		ConfigID:    co.ConfigID,
		Channel:     co.Channel,
		Version:     co.Version,
		MobileOs:    co.MobileOs,
		ItemID:      co.ItemID,
		ItemValue:   co.ItemValue,
		Status:      co.Status,
		Description: co.Description,
		CreatedAt:   mdtime.TimeFromProto(co.CreatedAt),
		UpdatedAt:   mdtime.TimeFromProto(co.UpdatedAt),
	}
}
