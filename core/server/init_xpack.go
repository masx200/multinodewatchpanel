//go:build xpack

package server

import (
	xpack "github.com/masx200/multinodewatchpanel/core/xpack"
)

func InitOthers() {
	xpack.Init()
}
