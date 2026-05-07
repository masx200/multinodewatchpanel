package v2

import (
	"fmt"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags Node
// @Summary List nodes
// @Success 200 {array} dto.NodeInfo
// @Security ApiKeyAuth
// @Router /api/v2/nodes [get]
func (b *BaseApi) ListNodes(c *gin.Context) {
	nodes, err := nodeService.List()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, nodes)
}

// @Tags Node
// @Summary Create node
// @Accept json
// @Param request body dto.NodeCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /api/v2/nodes [post]
func (b *BaseApi) CreateNode(c *gin.Context) {
	var req dto.NodeCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := nodeService.Create(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Node
// @Summary Update node
// @Accept json
// @Param id path int true "Node ID"
// @Param request body dto.NodeUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /api/v2/nodes/:id [put]
func (b *BaseApi) UpdateNode(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	var req dto.NodeUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := nodeService.Update(id, req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Node
// @Summary Delete node
// @Param id path int true "Node ID"
// @Success 200
// @Security ApiKeyAuth
// @Router /api/v2/nodes/:id [delete]
func (b *BaseApi) DeleteNode(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	if err := nodeService.Delete(id); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Node
// @Summary Test node connection
// @Accept json
// @Param request body dto.NodeCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /api/v2/nodes/test [post]
func (b *BaseApi) TestNodeConnection(c *gin.Context) {
	var req dto.NodeCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := nodeService.TestConnection(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Node
// @Summary Get node dashboard
// @Param id path int true "Node ID"
// @Success 200 {object} dto.NodeDashboard
// @Security ApiKeyAuth
// @Router /api/v2/nodes/:id/dashboard [get]
func (b *BaseApi) GetNodeDashboard(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	nodeInfo, err := nodeService.GetByID(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	client, err := nodeService.GetClient(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	current, err := client.GetDashboardCurrent("", "")
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	osInfo, _ := client.GetDashboardOS()

	dashboard := dto.NodeDashboard{
		NodeID:   nodeInfo.ID,
		NodeName: nodeInfo.Name,
		Status:   nodeInfo.Status,
		CPU:      current.CPUPercent,
		Memory:   current.MemoryPercent,
		Disk:     current.DiskPercent,
		NetUp:    current.NetworkUpload,
		NetDown:  current.NetworkDownload,
		IORead:   current.IORead,
		IOWrite:  current.IOWrite,
		Load1:    current.Load1,
		Load5:    current.Load5,
		Load15:   current.Load15,
	}
	if osInfo != nil {
		dashboard.Hostname = osInfo.Hostname
		dashboard.Uptime = osInfo.Uptime
		dashboard.OSVersion = fmt.Sprintf("%s %s", osInfo.Platform, osInfo.PlatformFamily)
	}
	helper.SuccessWithData(c, dashboard)
}

// @Tags Node
// @Summary Get node containers
// @Param id path int true "Node ID"
// @Success 200 {array} utils.ContainerInfo
// @Security ApiKeyAuth
// @Router /api/v2/nodes/:id/containers [get]
func (b *BaseApi) GetNodeContainers(c *gin.Context) {
	id, err := helper.GetParamID(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	client, err := nodeService.GetClient(id)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	result, err := client.ListContainersSimple(1, 100)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Node
// @Summary Search monitor history
// @Accept json
// @Param request body dto.MonitorSearchReq true "request"
// @Success 200 {array} dto.MonitorHistoryItem
// @Security ApiKeyAuth
// @Router /api/v2/nodes/monitor/search [post]
func (b *BaseApi) SearchMonitorHistory(c *gin.Context) {
	var req dto.MonitorSearchReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	layout := "2006-01-02 15:04:05"
	start, err := time.Parse(layout, req.StartTime)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	end, err := time.Parse(layout, req.EndTime)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	records, err := monitorRepoInstance.ListByNodeAndType(req.NodeID, req.MetricType, start, end)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	var items []dto.MonitorHistoryItem
	for _, r := range records {
		items = append(items, dto.MonitorHistoryItem{
			Time:  r.RecordedAt.Format(layout),
			Value: r.MetricValue,
		})
	}
	helper.SuccessWithData(c, items)
}
