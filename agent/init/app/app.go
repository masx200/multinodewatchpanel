package app

import (
	"github.com/masx200/multinodewatchpanel/agent/utils/docker"
)

func Init() {
	go func() {
		_ = docker.CreateDefaultDockerNetwork()
	}()
}
