//go:build xpackee

package router

import (
	xpackEERouter "github.com/masx200/multinodewatchpanel/core/xpack-ee/router"
	xpackRouter "github.com/masx200/multinodewatchpanel/core/xpack/router"
)

func RouterGroups() []CommonRouter {
	baseRouter := commonGroups()
	for _, ro := range xpackRouter.XpackGroups() {
		if val, ok := ro.(CommonRouter); ok {
			baseRouter = append(baseRouter, val)
		}
	}
	for _, ro := range xpackEERouter.XpackEEGroups() {
		if val, ok := ro.(CommonRouter); ok {
			baseRouter = append(baseRouter, val)
		}
	}
	return baseRouter
}

var RouterGroupApp = RouterGroups()
