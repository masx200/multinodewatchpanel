package migrations

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// AddNodeMonitorTables 添加多节点监控相关表
var AddNodeMonitorTables = &gormigrate.Migration{
	ID: "20260507-add-node-monitor-tables",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.HostNode{},
			&model.MonitorHistory{},
		)
	},
}

// AddNodeMonitorSettings 添加节点监控相关配置项
var AddNodeMonitorSettings = &gormigrate.Migration{
	ID: "20260507-add-node-monitor-settings",
	Migrate: func(tx *gorm.DB) error {
		settings := []model.Setting{
			{Key: "DataRetentionDays", Value: "30"},
			{Key: "CollectInterval", Value: "60"},
			{Key: "DashboardRefresh", Value: "10"},
		}
		for _, s := range settings {
			var existing model.Setting
			if err := tx.Where("key = ?", s.Key).First(&existing).Error; err == nil {
				continue // 已存在则跳过
			}
			if err := tx.Create(&s).Error; err != nil {
				return err
			}
		}
		return nil
	},
}

// AddNodeSecurityFields 为 host_nodes 表添加安全字段
var AddNodeSecurityFields = &gormigrate.Migration{
	ID: "20260507-add-node-security-fields",
	Migrate: func(tx *gorm.DB) error {
		// AutoMigrate 会自动补全新字段（不删除已有列）
		return tx.AutoMigrate(&model.HostNode{})
	},
}
