package v2

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/masx200/multinodewatchpanel/core/app/api/v2/helper"
	"github.com/masx200/multinodewatchpanel/core/app/dto"
	"github.com/masx200/multinodewatchpanel/core/utils"
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
// @Param request body dto.NodeTest true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /api/v2/nodes/test [post]
func (b *BaseApi) TestNodeConnection(c *gin.Context) {
	var req dto.NodeTest
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

	current, err := client.GetDashboardCurrent("all", "all")
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	base, _ := client.GetDashboardBase("all", "all")

	dashboard := dto.NodeDashboard{
		NodeID:       nodeInfo.ID,
		NodeName:     nodeInfo.Name,
		Status:       nodeInfo.Status,
		CPU:          current.CPUUsedPercent,
		Memory:       current.MemoryUsedPercent,
		MemoryTotal:  current.MemoryTotal,
		MemoryUsed:   current.MemoryUsed,
		MemoryAvail:  current.MemoryAvailable,
		SwapTotal:    current.SwapMemoryTotal,
		SwapUsed:     current.SwapMemoryUsed,
		Load1:        current.Load1,
		Load5:        current.Load5,
		Load15:       current.Load15,
	}
	// 从 DiskData 计算总体磁盘使用率
	if len(current.DiskData) > 0 {
		var totalUsed, totalAll float64
		var diskTotalSum, diskUsedSum int64
		for _, d := range current.DiskData {
			totalUsed += d.UsedPercent
			totalAll++
			diskTotalSum += d.Total
			diskUsedSum += d.Used
		}
		dashboard.Disk = totalUsed / totalAll
		dashboard.DiskTotal = diskTotalSum
		dashboard.DiskUsed = diskUsedSum
	}
	dashboard.NetUp = float64(current.NetBytesSent)
	dashboard.NetDown = float64(current.NetBytesRecv)
	dashboard.IORead = float64(current.IOReadBytes)
	dashboard.IOWrite = float64(current.IOWriteBytes)

	if base != nil {
		dashboard.Hostname = base.Hostname
		dashboard.Uptime = uint64(base.CurrentInfo.Uptime)
		dashboard.OSVersion = base.PrettyDistro
		dashboard.IPAddress = base.IPv4Addr
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
	var containers []utils.ContainerInfo
	if err := json.Unmarshal(result.Items, &containers); err != nil {
		helper.InternalServer(c, err)
		return
	}

	// 获取所有运行容器的统计信息
	statsList, err := client.GetContainerStatsList()
	if err == nil && len(statsList) > 0 {
		// 构建 containerID -> stats 映射
		statsMap := make(map[string]utils.ContainerStats)
		for _, stats := range statsList {
			statsMap[stats.ContainerID] = stats
		}
		// 合并统计信息到容器列表
		for i := range containers {
			if stats, ok := statsMap[containers[i].ContainerID]; ok {
				containers[i].CPUPercent = stats.CPUPercent
				containers[i].MemUsage = float64(stats.MemoryUsage)
				containers[i].MemLimit = float64(stats.MemoryLimit)
			}
		}
	}

	helper.SuccessWithData(c, containers)
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

// @Tags Node
// @Summary Get node heatmap data
// @Success 200 {object} dto.NodeHeatmapData
// @Security ApiKeyAuth
// @Router /api/v2/nodes/heatmap [get]
func (b *BaseApi) GetNodeHeatmapData(c *gin.Context) {
	data, err := nodeService.GetHeatmapData()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}
