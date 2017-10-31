package mod

import (
	mdtime "playcards/model/time"
	pbconf "playcards/proto/config"
	"time"
)

type Config struct {
	ID          int32 `gorm:"primary_key"`
	Name        string
	Description string
	Value       string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

type ConfigOpen struct {
	OpenID      int32 `gorm:"primary_key"`
	Channel     string
	Version     string
	MobileOs    string
	ItemID      int32
	ItemValue   string
	Value       string
	Status      int32
	Description string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

func (c *Config) ToProto() *pbconf.Config {
	return &pbconf.Config{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Value:       c.Value,
		CreatedAt:   mdtime.TimeToProto(c.CreatedAt),
		UpdatedAt:   mdtime.TimeToProto(c.UpdatedAt),
	}
}

func ConfigFromProto(c *pbconf.Config) *Config {
	return &Config{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Value:       c.Value,
		CreatedAt:   mdtime.TimeFromProto(c.CreatedAt),
		UpdatedAt:   mdtime.TimeFromProto(c.UpdatedAt),
	}
}

func (co *ConfigOpen) ToProto() *pbconf.ConfigOpen {
	return &pbconf.ConfigOpen{
		OpenID:      co.OpenID,
		Channel:     co.Channel,
		Version:     co.Version,
		MobileOs:    co.MobileOs,
		ItemID:      co.ItemID,
		ItemValue:   co.ItemValue,
		Value:       co.Value,
		Status:      co.Status,
		Description: co.Description,
	}
}

func ConfigOpenFromProto(co *pbconf.ConfigOpen) *ConfigOpen {
	return &ConfigOpen{
		OpenID:      co.OpenID,
		Channel:     co.Channel,
		Version:     co.Version,
		MobileOs:    co.MobileOs,
		ItemID:      co.ItemID,
		ItemValue:   co.ItemValue,
		Value:       co.Value,
		Status:      co.Status,
		Description: co.Description,
	}
}
