package db

import (
	"playcards/model/config/enum"
	mdconf "playcards/model/config/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jinzhu/gorm"
)

func UpdateConfig(tx *gorm.DB, c *mdconf.Config) error {
	now := gorm.NowFunc()

	m := make(map[string]interface{})
	m["id"] = c.ID
	m["name"] = c.Name
	m["description"] = c.Description
	m["value"] = c.Value
	m["created_at"] = now
	m["updated_at"] = now

	usql := "set name = ?, description = ?, value = ?, updated_at = ?"
	uargs := []interface{}{
		c.Name,
		c.Description,
		c.Value,
		now,
	}

	sql, args, _ := sq.Insert(enum.ConfigTableName).SetMap(m).
		Suffix("ON DUPLICATE KEY UPDATE "+usql, uargs...).ToSql()

	if err := tx.Exec(sql, args...).Error; err != nil {
		return errors.Internal("update config failed", err)
	}

	return nil
}

func ConfigList(tx *gorm.DB) ([]*mdconf.Config, error) {
	var out []*mdconf.Config

	if err := tx.Find(&out).Error; err != nil {
		return nil, errors.Internal("config list failed", err)
	}

	return out, nil
}

func GetConfigByID(tx *gorm.DB, cid int32) (*mdconf.Config, error) {
	var out mdconf.Config

	err := tx.Where("id = ?", cid).Find(&out).Error
	found, err := db.FoundRecord(err)

	if err != nil {
		return nil, errors.Internal("get specific config failed", err)
	}

	if !found {
		return nil, nil
	}

	return &out, nil
}
