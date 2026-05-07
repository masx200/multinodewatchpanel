//go:build xpackee

package server

import (
	xpack "github.com/masx200/multinodewatchpanel/core/xpack"
	xpackEE "github.com/masx200/multinodewatchpanel/core/xpack-ee"
)

func InitOthers() {
	xpack.Init()
	xpackEE.Init()
}
