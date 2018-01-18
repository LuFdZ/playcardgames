package config

import (
	cachec "playcards/model/config/cache"
	dbconf "playcards/model/config/db"
	enumc "playcards/model/config/enum"
	mdconf "playcards/model/config/mod"
	mdpage "playcards/model/page"
	"playcards/utils/db"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
	"strconv"
	"playcards/utils/log"
)

func UpdateConfig(co *mdconf.Config) error {
	if co.ConfigID < 1 {
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

func GetConfigs(channel string, version string, mobileOs string) map[int32]*mdconf.Config {
	log.Debug("AAAAGetConfigs:%s|%s|%s\n",channel,version,mobileOs)
	f := func(co *mdconf.Config) bool {
		if co.Status == enumc.ConfigOpenStatusAble &&
			(co.Channel == channel || len(co.Channel) == 0) &&
			(co.Version == version || len(co.Version) == 0) &&
			(co.MobileOs == mobileOs || len(co.MobileOs) == 0) {
				log.Debug("GetConfigsOk:%v\n",co)
			return true
		}else{
			log.Debug("GetConfigsNotOk:%v\n",co)
		}
		return false
	}
	return cachec.GetAllConfig(f)
}

func GetUniqueConfigByItemID(channel string, version string, mobileOs string) []*mdconf.Config {
	cm := GetConfigs(channel, version, mobileOs)

	var cos []*mdconf.Config
	//str := "GetUniqueConfigByItemID List "
	for _, co := range cm {
		cos = append(cos, co)
		//str += fmt.Sprintf("ID:%d ",co.ConfigID)
	}
	//fmt.Printf("%s\n",str)
	return cos
}

func RefreshAllConfigsFromDB() error {
	cos, err := dbconf.GetConfigs(db.DB())
	if err != nil {
		return err
	}
	return cachec.SetConfigs(cos)
}

func PageConfigs(page *mdpage.PageOption, c *mdconf.Config) (
	[]*mdconf.Config, int64, error) {
	return dbconf.PageConfigs(db.DB(), page, c)
}

func CheckConfigCondition(channel string, version string, mobileOs string) float64 {
	rate := 1.00
	cm := GetConfigs(channel, version, mobileOs)
	for itemID, co := range cm {
		if itemID == enumc.ConsumeOpen {
			value, _ := strconv.Atoi(co.ItemValue)
			rate = float64(value) / 100
		}
	}
	return rate
}