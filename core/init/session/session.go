package session

import (
	"github.com/masx200/multinodewatchpanel/core/global"
	"github.com/masx200/multinodewatchpanel/core/init/session/psession"
)

func Init() {
	global.SESSION = psession.NewPSession("")
	global.LOG.Info("init in-memory session successfully")
}
