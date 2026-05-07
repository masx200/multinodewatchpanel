package model

import "time"

type MonitorHistory struct {
	ID          uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	NodeID      uint      `gorm:"not null;index:idx_node_time" json:"nodeId"`
	MetricType  string    `gorm:"not null;size:32;index:idx_type_time" json:"metricType"` // cpu/memory/disk/io/network
	MetricValue float64   `gorm:"type:decimal(10,2);not null" json:"metricValue"`
	RecordedAt  time.Time `gorm:"not null;index:idx_node_time;index:idx_type_time" json:"recordedAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (m MonitorHistory) TableName() string {
	return "monitor_history"
}
