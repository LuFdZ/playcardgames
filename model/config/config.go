package config

import (
	dbconf "playcards/model/config/db"
	mdconf "playcards/model/config/mod"
	"playcards/utils/db"
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
