package job

import (
	"github.com/masx200/multinodewatchpanel/agent/app/service"
	"github.com/masx200/multinodewatchpanel/agent/global"
)

type app struct{}

func NewAppStoreJob() *app {
	return &app{}
}

func (a *app) Run() {
	global.LOG.Info("AppStore scheduled task in progress ...")
	if err := service.NewIAppService().SyncAppListFromRemote(""); err != nil {
		global.LOG.Errorf("AppStore sync failed %s", err.Error())
	}
	global.LOG.Info("AppStore scheduled task has completed")
}
