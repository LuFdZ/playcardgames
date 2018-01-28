package mod

import (
	mdtime "playcards/model/time"
	pbconf "playcards/proto/config"
	"time"
)

type Config struct {
	ConfigID    int32 `gorm:"primary_key"`
	Channel     string
	Version     string
	MobileOs    string
	ItemID      int32
	ItemValue   string
	Status      int32
	Description string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	Hkey        string `gorm:"-"`
}

func (co *Config) ToProto() *pbconf.Config {
	return &pbconf.Config{
		ItemID:      co.ItemID,
		ItemValue:   co.ItemValue,
	}
}


func (co *Config) ToDetailProto() *pbconf.Config {
	return &pbconf.Config{
		ConfigID:    co.ConfigID,
		Channel:     co.Channel,
		Version:     co.Version,
		MobileOs:    co.MobileOs,
		ItemID:      co.ItemID,
		ItemValue:   co.ItemValue,
		Status:      co.Status,
		Description: co.Description,
		CreatedAt:   mdtime.TimeToProto(co.CreatedAt),
		UpdatedAt:   mdtime.TimeToProto(co.UpdatedAt),
	}
}


func ConfigFromProto(co *pbconf.Config) *Config {
	if co == nil {
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
