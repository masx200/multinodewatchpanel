package repo

import (
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	"gorm.io/gorm"
)

type NodeRepo struct{}

type INodeRepo interface {
	Create(node *model.HostNode) error
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error
	GetByID(id uint) (model.HostNode, error)
	List(opts ...global.DBOption) ([]model.HostNode, error)
	UpdateStatus(id uint, status int, lastSeen *time.Time) error
}

func NewINodeRepo() INodeRepo {
	return &NodeRepo{}
}

func getNodeDB(opts ...global.DBOption) *gorm.DB {
	db := global.DB.Model(&model.HostNode{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}

func (r *NodeRepo) Create(node *model.HostNode) error {
	return global.DB.Create(node).Error
}

func (r *NodeRepo) Update(id uint, updates map[string]interface{}) error {
	return global.DB.Model(&model.HostNode{}).Where("id = ?", id).Updates(updates).Error
}

func (r *NodeRepo) Delete(id uint) error {
	return global.DB.Delete(&model.HostNode{}, id).Error
}

func (r *NodeRepo) GetByID(id uint) (model.HostNode, error) {
	var node model.HostNode
	err := global.DB.Where("id = ?", id).First(&node).Error
	return node, err
}

func (r *NodeRepo) List(opts ...global.DBOption) ([]model.HostNode, error) {
	var nodes []model.HostNode
	err := getNodeDB(opts...).Find(&nodes).Error
	return nodes, err
}

func (r *NodeRepo) UpdateStatus(id uint, status int, lastSeen *time.Time) error {
	updates := map[string]interface{}{"status": status}
	if lastSeen != nil {
		updates["last_seen"] = lastSeen
	}
	return global.DB.Model(&model.HostNode{}).Where("id = ?", id).Updates(updates).Error
}
