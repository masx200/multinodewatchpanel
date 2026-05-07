package repo

import (
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
)

type MonitorRepo struct{}

type IMonitorRepo interface {
	BatchCreate(records []model.MonitorHistory) error
	ListByNodeAndType(nodeID uint, metricType string, start, end time.Time) ([]model.MonitorHistory, error)
	DeleteBefore(before time.Time) error
	DeleteByNodeID(nodeID uint) error
}

func NewIMonitorRepo() IMonitorRepo {
	return &MonitorRepo{}
}

func (r *MonitorRepo) BatchCreate(records []model.MonitorHistory) error {
	if len(records) == 0 {
		return nil
	}
	return global.DB.CreateInBatches(records, 100).Error
}

func (r *MonitorRepo) ListByNodeAndType(nodeID uint, metricType string, start, end time.Time) ([]model.MonitorHistory, error) {
	var records []model.MonitorHistory
	err := global.DB.
		Where("node_id = ? AND metric_type = ? AND recorded_at BETWEEN ? AND ?", nodeID, metricType, start, end).
		Order("recorded_at ASC").
		Find(&records).Error
	return records, err
}

func (r *MonitorRepo) DeleteBefore(before time.Time) error {
	return global.DB.Where("recorded_at < ?", before).Delete(&model.MonitorHistory{}).Error
}

func (r *MonitorRepo) DeleteByNodeID(nodeID uint) error {
	return global.DB.Where("node_id = ?", nodeID).Delete(&model.MonitorHistory{}).Error
}
