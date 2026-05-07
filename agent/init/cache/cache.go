package cache

import (
	"github.com/masx200/multinodewatchpanel/agent/global"
	cachedb "github.com/masx200/multinodewatchpanel/agent/init/cache/db"
)

func Init() {
	global.CACHE = cachedb.NewCacheDB()
}
