package db

import (
	"playcards/model/config/enum"
	mdconf "playcards/model/config/mod"
	"playcards/utils/errors"
	mdpage "playcards/model/page"
	"github.com/jinzhu/gorm"
)

func CreateConfig(tx *gorm.DB, co *mdconf.Config) (*mdconf.Config, error) {
	now := gorm.NowFunc()
	co.UpdatedAt = &now
	co.CreatedAt = &now
	if err := tx.Create(co).Error; err != nil {
		return nil, errors.Internal("create config failed", err)
	}
	return co, nil
}

func UpdateConfig(tx *gorm.DB, co *mdconf.Config) (*mdconf.Config, error) {
	now := gorm.NowFunc()
	configopen := &mdconf.Config{
		Channel:     co.Channel,
		Version:     co.Version,
		MobileOs:    co.MobileOs,
		Status:      co.Status,
		ItemValue:   co.ItemValue,
		Description: co.Description,
		UpdatedAt:   &now,
	}
	if err := tx.Model(co).Updates(configopen).Error; err != nil {
		return nil, errors.Internal("update config failed", err)
	}
	return configopen, nil
}

func GetConfigs(tx *gorm.DB) ([]*mdconf.Config, error) {
	var (
		out []*mdconf.Config
	)
	if err := tx.Where("status = ?", enum.ConfigOpenStatusAble).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errors.Internal("select config list failed", err)
	}
	return out, nil
}

func PageConfigs(tx *gorm.DB, page *mdpage.PageOption,
	n *mdconf.Config) ([]*mdconf.Config, int64, error) {
	var out []*mdconf.Config
	rows, rtx := page.Find(tx.Model(n).Order("created_at desc").
		Where(n), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page config failed", rtx.Error)
	}
	return out, rows, nil
}
