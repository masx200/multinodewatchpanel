package service

import (
	"strconv"
	"sync"
	"time"

	"github.com/masx200/multinodewatchpanel/core/app/model"
	"github.com/masx200/multinodewatchpanel/core/app/repo"
	"github.com/masx200/multinodewatchpanel/core/global"
)

type ICollectorService interface {
	Start()
	Stop()
}

type CollectorService struct {
	nodeRepo    repo.INodeRepo
	monitorRepo repo.IMonitorRepo
	interval    time.Duration
	retention   int // 保留天数
	stopChan    chan struct{}
	once        sync.Once
}

func NewICollectorService() ICollectorService {
	retention := 30
	interval := 60

	if v, err := settingRepo.GetValueByKey("DataRetentionDays"); err == nil {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			retention = n
		}
	}
	if v, err := settingRepo.GetValueByKey("CollectInterval"); err == nil {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			interval = n
		}
	}

	return &CollectorService{
		nodeRepo:    repo.NewINodeRepo(),
		monitorRepo: repo.NewIMonitorRepo(),
		interval:    time.Duration(interval) * time.Second,
		retention:   retention,
		stopChan:    make(chan struct{}),
	}
}

func (c *CollectorService) Start() {
	c.once.Do(func() {
		go c.run()
	})
}

func (c *CollectorService) Stop() {
	close(c.stopChan)
}

func (c *CollectorService) run() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	cleanTicker := time.NewTicker(24 * time.Hour)
	defer cleanTicker.Stop()

	// 启动时立即采集一次
	c.collectAll()

	for {
		select {
		case <-ticker.C:
			c.collectAll()
		case <-cleanTicker.C:
			c.cleanup()
		case <-c.stopChan:
			return
		}
	}
}

func (c *CollectorService) collectAll() {
	nodes, err := c.nodeRepo.List()
	if err != nil {
		global.LOG.Errorf("collector: list nodes failed: %v", err)
		return
	}
	var wg sync.WaitGroup
	for _, n := range nodes {
		wg.Add(1)
		go func(node model.HostNode) {
			defer wg.Done()
			c.collectNode(node)
		}(n)
	}
	wg.Wait()
}

func (c *CollectorService) collectNode(node model.HostNode) {
	client, err := nodeServiceInstance.GetClient(node.ID)
	if err != nil {
		global.LOG.Errorf("collector: get client for node %d failed: %v", node.ID, err)
		_ = c.nodeRepo.UpdateStatus(node.ID, 2, nil)
		return
	}

	current, err := client.GetDashboardCurrent("all", "all")
	if err != nil {
		global.LOG.Warnf("collector: collect node %d failed: %v", node.ID, err)
		_ = c.nodeRepo.UpdateStatus(node.ID, 2, nil)
		return
	}

	now := time.Now()
	_ = c.nodeRepo.UpdateStatus(node.ID, 1, &now)

	records := []model.MonitorHistory{
		{NodeID: node.ID, MetricType: "cpu", MetricValue: current.CPUUsedPercent, RecordedAt: now},
		{NodeID: node.ID, MetricType: "memory", MetricValue: current.MemoryUsedPercent, RecordedAt: now},
		{NodeID: node.ID, MetricType: "load1", MetricValue: current.Load1, RecordedAt: now},
		{NodeID: node.ID, MetricType: "net_upload", MetricValue: float64(current.NetBytesSent), RecordedAt: now},
		{NodeID: node.ID, MetricType: "net_download", MetricValue: float64(current.NetBytesRecv), RecordedAt: now},
		{NodeID: node.ID, MetricType: "io_read", MetricValue: float64(current.IOReadBytes), RecordedAt: now},
		{NodeID: node.ID, MetricType: "io_write", MetricValue: float64(current.IOWriteBytes), RecordedAt: now},
	}

	if err := c.monitorRepo.BatchCreate(records); err != nil {
		global.LOG.Errorf("collector: save metrics for node %d failed: %v", node.ID, err)
	}
}

func (c *CollectorService) cleanup() {
	// 重新读取保留天数配置
	if v, err := settingRepo.GetValueByKey("DataRetentionDays"); err == nil {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			c.retention = n
		}
	}
	before := time.Now().AddDate(0, 0, -c.retention)
	if err := c.monitorRepo.DeleteBefore(before); err != nil {
		global.LOG.Errorf("collector: cleanup old data failed: %v", err)
	} else {
		global.LOG.Infof("collector: cleaned up monitor data before %s", before.Format("2006-01-02"))
	}
}
