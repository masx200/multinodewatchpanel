package run

import (
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
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
