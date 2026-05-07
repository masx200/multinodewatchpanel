package run

import (
	"github.com/masx200/multinodewatchpanel/core/app/dto"
	"github.com/masx200/multinodewatchpanel/core/app/repo"
	"github.com/masx200/multinodewatchpanel/core/app/service"
	"github.com/masx200/multinodewatchpanel/core/constant"
	"github.com/masx200/multinodewatchpanel/core/global"
)

func Init() {
	scriptSync, _ := repo.NewISettingRepo().GetValueByKey("ScriptSync")
	if !global.CONF.Base.IsOffLine && scriptSync == constant.StatusEnable {
		if err := service.NewIScriptService().Sync(dto.OperateByTaskID{}); err != nil {
			global.LOG.Errorf("sync scripts from remote failed, err: %v", err)
		}
	}

	// 启动多节点监控数据采集器
	collector := service.NewICollectorService()
	collector.Start()
	global.LOG.Info("Multi-node collector started")
}
