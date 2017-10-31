package config

import (
	dbconf "playcards/model/config/db"
	mdconf "playcards/model/config/mod"
	cachec "playcards/model/config/cache"
	enumc "playcards/model/config/enum"
	mdpage "playcards/model/page"
	"playcards/utils/db"
	"github.com/jinzhu/gorm"
)

func UpdateConfig(c *mdconf.Config) error {
	return dbconf.UpdateConfig(db.DB(), c)
}

func ConfigList() ([]*mdconf.Config, error) {
	return dbconf.ConfigList(db.DB())
}

func GetConfigByID(cid int32) (*mdconf.Config, error) {
	return dbconf.GetConfigByID(db.DB(), cid)
}

func UpdateConfigOpen(co *mdconf.ConfigOpen) error {
	return db.Transaction(func(tx *gorm.DB) error {
		co, err := dbconf.UpdateConfigOpen(db.DB(), co)
		if err != nil {
			return err
		}
		err = cachec.SetConfigOpen(co)
		if err != nil {
			return err
		}
		return nil
	})
}

func CreateConfigOpen(co *mdconf.ConfigOpen) error {
	return db.Transaction(func(tx *gorm.DB) error {
		co, err := dbconf.CreateConfigOpen(tx, co)
		if err != nil {
			return err
		}
		err = cachec.SetConfigOpen(co)
		if err != nil {
			return err
		}
		return nil
	})
}

func GetConfigOpens(channel string,version string,mobileOs string) ([]*mdconf.ConfigOpen){
	f := func(co *mdconf.ConfigOpen) bool {
		if co.Status == enumc.ConfigOpenStatusAble &&
			(co.Channel == channel || len(co.Channel) == 0) &&
			(co.Version == version || len(co.Version) == 0) &&
			(co.MobileOs == mobileOs || len(co.MobileOs) == 0) {
			return true
		}
		return false
	}
	return cachec.GetAllConfigOpen(f)
}

func RefreshAllConfigOpensFromDB() error{
	cos,err :=dbconf.GetConfigOpens(db.DB())
	if err != nil {
		return err
	}
	return cachec.SetConfigOpens(cos)
}

func PageConfigOpens(page *mdpage.PageOption, n *mdconf.ConfigOpen) (
	[]*mdconf.ConfigOpen, int64, error) {
	return dbconf.PageConfigOpens(db.DB(), page, n)
}
