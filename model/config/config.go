package config

import (
	dbconf "playcards/model/config/db"
	mdconf "playcards/model/config/mod"
	cachec "playcards/model/config/cache"
	enumc "playcards/model/config/enum"
	mdpage "playcards/model/page"
	"playcards/utils/errors"
	"playcards/utils/db"
	"github.com/jinzhu/gorm"
)

func UpdateConfig(co *mdconf.Config) error {
	if co.ConfigID <1{
		return errors.Conflict(70001, "未找到数据ID！")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		_, err := dbconf.UpdateConfig(db.DB(), co)
		if err != nil {
			return err
		}
		//err = cachec.SetConfig(co)
		//if err != nil {
		//	return err
		//}
		return nil
	})
}

func CreateConfig(co *mdconf.Config) error {
	return db.Transaction(func(tx *gorm.DB) error {
		_, err := dbconf.CreateConfig(tx, co)
		if err != nil {
			return err
		}
		//err = cachec.SetConfig(co)
		//if err != nil {
		//	return err
		//}
		return nil
	})
}

func GetConfigs(channel string,version string,mobileOs string) map[int32]*mdconf.Config{
	f := func(co *mdconf.Config) bool {
		if co.Status == enumc.ConfigOpenStatusAble &&
			(co.Channel == channel || len(co.Channel) == 0) &&
			(co.Version == version || len(co.Version) == 0) &&
			(co.MobileOs == mobileOs || len(co.MobileOs) == 0) {
			return true
		}
		return false
	}
	return cachec.GetAllConfig(f)
}

func GetUniqueConfigByItemID(channel string,version string,mobileOs string) []*mdconf.Config {
	cm := GetConfigs(channel,version,mobileOs)
	var cos []*mdconf.Config
	for _,co :=range cm{
		cos = append(cos,co)
	}
	return cos
}

func RefreshAllConfigsFromDB() error{
	cos,err :=dbconf.GetConfigs(db.DB())
	if err != nil {
		return err
	}
	return cachec.SetConfigs(cos)
}

func PageConfigs(page *mdpage.PageOption, c *mdconf.Config) (
	[]*mdconf.Config, int64, error) {
	return dbconf.PageConfigs(db.DB(), page, c)
}

