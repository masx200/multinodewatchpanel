package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/masx200/multinodewatchpanel/agent/app/repo"
	"github.com/masx200/multinodewatchpanel/agent/constant"
	"github.com/masx200/multinodewatchpanel/agent/cron"
	"github.com/masx200/multinodewatchpanel/agent/global"
	"github.com/masx200/multinodewatchpanel/agent/i18n"
	"github.com/masx200/multinodewatchpanel/agent/init/app"
	"github.com/masx200/multinodewatchpanel/agent/init/business"
	"github.com/masx200/multinodewatchpanel/agent/init/cache"
	"github.com/masx200/multinodewatchpanel/agent/init/db"
	"github.com/masx200/multinodewatchpanel/agent/init/dir"
	"github.com/masx200/multinodewatchpanel/agent/init/firewall"
	"github.com/masx200/multinodewatchpanel/agent/init/hook"
	"github.com/masx200/multinodewatchpanel/agent/init/lang"
	"github.com/masx200/multinodewatchpanel/agent/init/log"
	"github.com/masx200/multinodewatchpanel/agent/init/migration"
	"github.com/masx200/multinodewatchpanel/agent/init/router"
	"github.com/masx200/multinodewatchpanel/agent/init/validator"
	"github.com/masx200/multinodewatchpanel/agent/init/viper"
	"github.com/masx200/multinodewatchpanel/agent/utils/encrypt"
	"github.com/masx200/multinodewatchpanel/agent/utils/re"
)

func Start() {
	re.Init()
	viper.Init()
	dir.Init()
	log.Init()
	db.Init()
	migration.Init()
	i18n.Init()
	cache.Init()
	app.Init()
	lang.Init()
	validator.Init()
	cron.Run()
	hook.Init()
	go firewall.Init()
	InitOthers()

	rootRouter := router.Routers()

	server := &http.Server{
		Handler: rootRouter,
	}

	if global.CONF.Base.Mode != "stable" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	if global.IsMaster {
		_ = os.Remove("/etc/1panel/agent.sock")
		_ = os.Mkdir("/etc/1panel", constant.DirPerm)
		listener, err := net.Listen("unix", "/etc/1panel/agent.sock")
		if err != nil {
			panic(err)
		}
		business.Init()
		_ = server.Serve(listener)
		return
	} else {
		server.Addr = fmt.Sprintf("0.0.0.0:%s", global.CONF.Base.Port)
		settingRepo := repo.NewISettingRepo()
		certItem, err := settingRepo.Get(settingRepo.WithByKey("ServerCrt"))
		if err != nil {
			panic(err)
		}
		cert, _ := encrypt.StringDecrypt(certItem.Value)
		keyItem, err := settingRepo.Get(settingRepo.WithByKey("ServerKey"))
		if err != nil {
			panic(err)
		}
		key, _ := encrypt.StringDecrypt(keyItem.Value)
		tlsCert, err := tls.X509KeyPair([]byte(cert), []byte(key))
		if err != nil {
			fmt.Printf("failed to load X.509 key pair: %s\n", err)
			return
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
		}
		caItem, _ := settingRepo.GetValueByKey("RootCrt")
		if len(caItem) != 0 {
			caCertPool := x509.NewCertPool()
			rootCrt, _ := encrypt.StringDecrypt(caItem)
			caCertPool.AppendCertsFromPEM([]byte(rootCrt))
			server.TLSConfig.ClientCAs = caCertPool
		}
		business.Init()
		global.LOG.Infof("listen at https://0.0.0.0:%s", global.CONF.Base.Port)
		if err := server.ListenAndServeTLS("", ""); err != nil {
			panic(err)
		}
	}
}
