package db

import (
	"playcards/model/config/enum"
	mdconf "playcards/model/config/mod"
	"playcards/utils/db"
	"playcards/utils/errors"
	mdpage "playcards/model/page"
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

func CreateConfigOpen(tx *gorm.DB, co *mdconf.ConfigOpen) (*mdconf.ConfigOpen,error) {
	now := gorm.NowFunc()
	co.UpdatedAt = &now
	co.CreatedAt = &now
	if err := tx.Create(co).Error; err != nil {
		return nil,errors.Internal("create configopen failed", err)
	}
	return co,nil
}

func UpdateConfigOpen(tx *gorm.DB, co *mdconf.ConfigOpen) (*mdconf.ConfigOpen, error) {
	now := gorm.NowFunc()
	configopen := &mdconf.ConfigOpen{
		Channel:   co.Channel,
		Version:   co.Version,
		MobileOs:  co.MobileOs,
		UpdatedAt: &now,
	}
	if err := tx.Model(co).Updates(configopen).Error; err != nil {
		return nil, errors.Internal("update configopen failed", err)
	}
	return configopen, nil
}

func GetConfigOpens(tx *gorm.DB) ([]*mdconf.ConfigOpen, error) {
	var (
		out []*mdconf.ConfigOpen
	)
	if err := tx.Where("status = ?", enum.ConfigOpenStatusAble).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("select configopen list failed", err)
	}
	return out, nil
}

func PageConfigOpens(tx *gorm.DB, page *mdpage.PageOption,
	n *mdconf.ConfigOpen) ([]*mdconf.ConfigOpen, int64, error) {
	var out []*mdconf.ConfigOpen
	rows, rtx := page.Find(tx.Model(n).Order("created_at desc").
		Where(n), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page configopens failed", rtx.Error)
	}
	return out, rows, nil
}
