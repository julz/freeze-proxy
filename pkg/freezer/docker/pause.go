package docker

import (
	"context"
	"fmt"
	"k8s.io/klog"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerapi "github.com/docker/docker/client"
)

const DefaultDockerUri = "unix:///var/run/docker.sock"

type Docker struct {
	runtimeURI string

	client *dockerapi.Client
}

func NewDockerService(runtimeURI string) (*Docker, error) {
	r := &Docker{runtimeURI: runtimeURI}
	if err := r.createRuntimeClientIfNecessary(); err != nil {
		return nil, err
	}
	return r, nil
}

func (d *Docker) createRuntimeClientIfNecessary() error {

	if d.client != nil {
		return nil
	}
	c, err := dockerapi.NewClientWithOpts(dockerapi.WithVersion("1.19"))
	if err != nil {
		return err
	}
	d.client = c
	return nil
}

func (d *Docker) Freeze(ctx context.Context, podUID, containerName string) error {
	klog.Infof("Start to freeze container %s/%s",podUID,containerName)
	containerID, err := d.lookupContainerID(ctx, podUID, containerName)
	if err != nil {
		return err
	}
	err = d.client.ContainerPause(ctx, containerID)
	if err != nil {
		return fmt.Errorf("error when pause container, err: %s", err.Error())
	}
	klog.Infof("Freeze container %s/%s success !",podUID,containerName)
	return nil
}

// Thaw thats a container which was freezed via the Freeze method.
func (d *Docker) Thaw(ctx context.Context, podUID, containerName string) error {
	klog.Infof("Start to thaw container %s/%s",podUID,containerName)
	containerID, err := d.lookupContainerID(ctx, podUID, containerName)
	if err != nil {
		return err
	}

	err = d.client.ContainerUnpause(ctx, containerID)
	if err != nil {
		return fmt.Errorf("error when unpause container, err: %s", err.Error())
	}
	klog.Infof("Thaw container %s/%s success !",podUID,containerName)
	return nil
}

func (d *Docker) lookupContainerID(ctx context.Context, podUID, containerName string) (string, error) {
	filter := filters.NewArgs()
	filter.Add("label", fmt.Sprintf("io.kubernetes.pod.uid=%s", podUID))
	filter.Add("label", fmt.Sprintf("io.kubernetes.container.name=%s", containerName))
	containers, err := d.client.ContainerList(context.Background(), types.ContainerListOptions{Filters: filter})
	if err != nil {
		return "", err
	}

	if len(containers) == 0 {
		return "", fmt.Errorf("container %q in pod %q not found", containerName, podUID)
	}

	return containers[0].ID, nil
}
