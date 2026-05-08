package migrations

import (
	"encoding/json"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/masx200/multinodewatchpanel/core/app/dto"
	"github.com/masx200/multinodewatchpanel/core/app/model"
	"github.com/masx200/multinodewatchpanel/core/init/migration/helper"
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

// AddMultiMonitorMenu 添加多节点监控菜单到 HideMenu
var AddMultiMonitorMenu = &gormigrate.Migration{
	ID: "20260508-add-multi-monitor-menu",
	Migrate: func(tx *gorm.DB) error {
		var menuJSON string
		if err := tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Pluck("value", &menuJSON).Error; err != nil {
			return err
		}
		if menuJSON == "" {
			return tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Update("value", helper.LoadMenus()).Error
		}

		var menus []dto.ShowMenu
		if err := json.Unmarshal([]byte(menuJSON), &menus); err != nil {
			return tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Update("value", helper.LoadMenus()).Error
		}

		// Check if Monitor-Menu already exists
		for _, menu := range menus {
			if menu.ID == "14" && menu.Label == "Monitor-Menu" {
				return nil
			}
		}

		// Build the Monitor-Menu entry
		monitorMenu := dto.ShowMenu{
			ID:       "14",
			Disabled: false,
			Title:    "menu.multiMonitor",
			IsShow:   true,
			Label:    "Monitor-Menu",
			Path:     "/monitor/dashboard",
			Sort:     150,
			Children: []dto.ShowMenu{
				{ID: "141", Disabled: false, Title: "menu.monitorDashboard", IsShow: true, Label: "MonitorDashboard", Path: "/monitor/dashboard", Sort: 100},
				{ID: "142", Disabled: false, Title: "menu.monitorHosts", IsShow: true, Label: "MonitorHosts", Path: "/monitor/hosts", Sort: 200},
				{ID: "143", Disabled: false, Title: "menu.monitorContainers", IsShow: true, Label: "MonitorContainers", Path: "/monitor/containers", Sort: 300},
				{ID: "144", Disabled: false, Title: "menu.monitorHistory", IsShow: true, Label: "MonitorHistory", Path: "/monitor/history", Sort: 400},
			},
		}

		// Insert after Home-Menu (Sort 100), before everything else
		inserted := false
		for i, menu := range menus {
			if menu.ID == "1" && menu.Label == "Home-Menu" {
				menus = append(menus[:i+1], append([]dto.ShowMenu{monitorMenu}, menus[i+1:]...)...)
				inserted = true
				break
			}
		}
		if !inserted {
			menus = append([]dto.ShowMenu{monitorMenu}, menus...)
		}

		updatedJSON, err := json.Marshal(menus)
		if err != nil {
			return tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Update("value", helper.LoadMenus()).Error
		}
		return tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Update("value", string(updatedJSON)).Error
	},
}
