//go:build xpack

package router

import (
	xpackRouter "github.com/masx200/multinodewatchpanel/core/xpack/router"
)

func RouterGroups() []CommonRouter {
	baseRouter := commonGroups()
	for _, ro := range xpackRouter.XpackGroups() {
		if val, ok := ro.(CommonRouter); ok {
			baseRouter = append(baseRouter, val)
		}
	}
	return baseRouter
}

var RouterGroupApp = RouterGroups()
