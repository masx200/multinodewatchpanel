package service

import (
	"errors"
	"sync"
	"time"

	"github.com/masx200/multinodewatchpanel/core/app/dto"
	"github.com/masx200/multinodewatchpanel/core/app/model"
	"github.com/masx200/multinodewatchpanel/core/app/repo"
	"github.com/masx200/multinodewatchpanel/core/utils"
	"gorm.io/gorm"
)

type INodeService interface {
	Create(req dto.NodeCreate) error
	Update(id uint, req dto.NodeUpdate) error
	Delete(id uint) error
	List() ([]dto.NodeInfo, error)
	GetByID(id uint) (dto.NodeInfo, error)
	TestConnection(req dto.NodeTest) error
	GetClient(nodeID uint) (*utils.PanelClient, error)
	RefreshStatus()
}

type NodeService struct {
	nodeRepo    repo.INodeRepo
	monitorRepo repo.IMonitorRepo
	clients     sync.Map // map[uint]*utils.PanelClient
}

func NewINodeService() INodeService {
	return &NodeService{
		nodeRepo:    repo.NewINodeRepo(),
		monitorRepo: repo.NewIMonitorRepo(),
	}
}

func (s *NodeService) Create(req dto.NodeCreate) error {
	node := &model.HostNode{
		Name:                 req.Name,
		Host:                 req.Host,
		Port:                 req.Port,
		APIKey:               req.APIKey,
		Tags:                 req.Tags,
		Status:               0,
		ServerName:           req.ServerName,
		Security:             req.Security,
		AllowInsecure:        req.AllowInsecure,
		PinnedPeerCertSHA256: req.PinnedPeerCertSHA256,
	}
	if err := s.nodeRepo.Create(node); err != nil {
		return err
	}
	// 立即后台测试连接并更新状态
	go s.checkAndUpdateStatus(node.ID)
	return nil
}

func (s *NodeService) Update(id uint, req dto.NodeUpdate) error {
	updates := map[string]interface{}{
		"name":                    req.Name,
		"host":                    req.Host,
		"port":                    req.Port,
		"tags":                    req.Tags,
		"server_name":             req.ServerName,
		"security":                req.Security,
		"allow_insecure":          req.AllowInsecure,
		"pinned_peer_cert_sha256": req.PinnedPeerCertSHA256,
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if err := s.nodeRepo.Update(id, updates); err != nil {
		return err
	}
	// 清除旧客户端缓存
	s.clients.Delete(id)
	return nil
}

func (s *NodeService) Delete(id uint) error {
	if err := s.nodeRepo.Delete(id); err != nil {
		return err
	}
	// 删除该节点的监控历史
	_ = s.monitorRepo.DeleteByNodeID(id)
	s.clients.Delete(id)
	return nil
}

func (s *NodeService) List() ([]dto.NodeInfo, error) {
	nodes, err := s.nodeRepo.List()
	if err != nil {
		return nil, err
	}
	var result []dto.NodeInfo
	for _, n := range nodes {
		result = append(result, toNodeInfo(n))
	}
	return result, nil
}

func (s *NodeService) GetByID(id uint) (dto.NodeInfo, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.NodeInfo{}, errors.New("node not found")
		}
		return dto.NodeInfo{}, err
	}
	return toNodeInfo(node), nil
}

func (s *NodeService) TestConnection(req dto.NodeTest) error {
	apiKey := req.APIKey
	// 编辑时如果 APIKey 为空，从数据库取已保存的值
	if apiKey == "" && req.ID > 0 {
		node, err := s.nodeRepo.GetByID(req.ID)
		if err != nil {
			return errors.New("node not found, cannot retrieve saved APIKey")
		}
		apiKey = node.APIKey
	}
	if apiKey == "" {
		return errors.New("apiKey is required")
	}
	client := utils.NewPanelClientWithOptions(req.Host, req.Port, apiKey, utils.PanelClientOptions{
		Security:             req.Security,
		AllowInsecure:        req.AllowInsecure,
		PinnedPeerCertSHA256: req.PinnedPeerCertSHA256,
		ServerName:           req.ServerName,
	})
	return client.Ping()
}

func (s *NodeService) GetClient(nodeID uint) (*utils.PanelClient, error) {
	if v, ok := s.clients.Load(nodeID); ok {
		return v.(*utils.PanelClient), nil
	}
	node, err := s.nodeRepo.GetByID(nodeID)
	if err != nil {
		return nil, err
	}
	client := utils.NewPanelClientWithOptions(node.Host, node.Port, node.APIKey, utils.PanelClientOptions{
		Security:             node.Security,
		AllowInsecure:        node.AllowInsecure,
		PinnedPeerCertSHA256: node.PinnedPeerCertSHA256,
		ServerName:           node.ServerName,
	})
	s.clients.Store(nodeID, client)
	return client, nil
}

func (s *NodeService) RefreshStatus() {
	nodes, err := s.nodeRepo.List()
	if err != nil {
		return
	}
	for _, n := range nodes {
		go s.checkAndUpdateStatus(n.ID)
	}
}

func (s *NodeService) checkAndUpdateStatus(id uint) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return
	}
	client := utils.NewPanelClientWithOptions(node.Host, node.Port, node.APIKey, utils.PanelClientOptions{
		Security:             node.Security,
		AllowInsecure:        node.AllowInsecure,
		PinnedPeerCertSHA256: node.PinnedPeerCertSHA256,
		ServerName:           node.ServerName,
	})
	err = client.Ping()
	now := time.Now()
	status := 1
	if err != nil {
		status = 2
	}
	_ = s.nodeRepo.UpdateStatus(id, status, &now)
	s.clients.Store(id, client)
}

func toNodeInfo(n model.HostNode) dto.NodeInfo {
	info := dto.NodeInfo{
		ID:                   n.ID,
		Name:                 n.Name,
		Host:                 n.Host,
		Port:                 n.Port,
		Status:               n.Status,
		Tags:                 n.Tags,
		IsDefault:            n.IsDefault,
		ServerName:           n.ServerName,
		Security:             n.Security,
		AllowInsecure:        n.AllowInsecure,
		PinnedPeerCertSHA256: n.PinnedPeerCertSHA256,
		CreatedAt:            n.CreatedAt,
		UpdatedAt:            n.UpdatedAt,
	}
	if n.LastSeen != nil {
		info.LastSeen = n.LastSeen
	}
	return info
}
