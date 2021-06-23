package freezer

import (
	"fmt"

	"github.com/julz/freeze-proxy/pkg/daemon"
	"github.com/julz/freeze-proxy/pkg/freezer/containerd"
	docker "github.com/julz/freeze-proxy/pkg/freezer/docker"
)

const (
	RuntimeTypeDocker     string = "docker"
	RuntimeTypeContainerd string = "contaienrd"
	RuntimeTypeCriO       string = "cri-o"
)

func GetFreezer(runtimeType string) (daemon.Freezer, error) {
	switch runtimeType {
	case RuntimeTypeDocker:
		return docker.NewDockerService(docker.DefaultDockerUri)
	case RuntimeTypeContainerd:
		return containerd.Connect()
		// TODO support cri-o
	}
	return nil, fmt.Errorf("unsupported cri runtime")
}

func GetThawer(runtimeType string) (daemon.Thawer, error) {
	switch runtimeType {
	case RuntimeTypeDocker:
		return docker.NewDockerService(docker.DefaultDockerUri)
	case RuntimeTypeContainerd:
		return containerd.Connect()
		// TODO support cri-o
	}
	return nil, fmt.Errorf("unsupported cri runtime")
}
