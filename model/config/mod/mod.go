package mod

import (
	mdtime "playcards/model/time"
	pbconf "playcards/proto/config"
	"time"
	"strconv"
)

type Config struct {
	ConfigID    int32 `gorm:"primary_key"`
	Channel     string
	Version     string
	MobileOs    string
	ItemID      string
	ItemValue   string
	Status      int32
	Description string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	Hkey        string `gorm:"-"`
}

func (co *Config) ToProto() *pbconf.Config {
	out := &pbconf.Config{
		ItemValue:   co.ItemValue,
	}
	i, err := strconv.ParseInt(co.ItemID, 10, 32)
	if err == nil{
		out.ItemID = int32(i)
		out.ItemName = co.ItemID
	}else{
		out.ItemName = co.ItemID
	}
	return out
}


func (co *Config) ToDetailProto() *pbconf.ConfigNew {
	return &pbconf.ConfigNew{
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


func ConfigFromProto(co *pbconf.ConfigNew) *Config {
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
