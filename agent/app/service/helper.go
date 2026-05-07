package service

import (
	"context"

	"github.com/masx200/multinodewatchpanel/agent/constant"
	"github.com/masx200/multinodewatchpanel/agent/global"
	"gorm.io/gorm"
)

func getTxAndContext() (tx *gorm.DB, ctx context.Context) {
	tx = global.DB.Begin()
	ctx = context.WithValue(context.Background(), constant.DB, tx)
	return
}
