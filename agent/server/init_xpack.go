//go:build xpack

package server

import (
	xpack "github.com/masx200/multinodewatchpanel/agent/xpack"
)

func InitOthers() {
	xpack.Init()
}
