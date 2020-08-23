package freezer

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/namespaces"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	cri "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

// Containerd freezes and unfreezes containers via containerd.
type Containerd struct {
	conn        *grpc.ClientConn
	containerId string
}

// Connect connects to containerd and looks up the containerId.
// Requires /var/run/containerd/containerd.sock to be mounted.
func Connect(logger *zap.SugaredLogger, podName, containerName string) (*Containerd, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "/var/run/containerd/containerd.sock", grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*16)), grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", addr)
	}))
	if err != nil {
		return nil, err
	}

	containerId, err := lookupContainerId(ctx, conn, podName, containerName)
	if err != nil {
		return nil, err
	}

	return &Containerd{
		conn:        conn,
		containerId: containerId,
	}, nil
}

// Freeze freezes the user container via the freezer cgroup.
func (f *Containerd) Freeze(ctx context.Context) error {
	ctrd, err := containerd.NewWithConn(f.conn)
	if err != nil {
		return err
	}

	ctx = namespaces.WithNamespace(ctx, "k8s.io")
	if _, err := ctrd.TaskService().Pause(ctx, &tasks.PauseTaskRequest{ContainerID: f.containerId}); err != nil {
		return err
	}

	return nil
}

// Thaw thats a container which was freezed via the Freeze method.
func (f *Containerd) Thaw(ctx context.Context) error {
	ctrd, err := containerd.NewWithConn(f.conn)
	if err != nil {
		return err
	}

	ctx = namespaces.WithNamespace(ctx, "k8s.io")
	if _, err := ctrd.TaskService().Resume(ctx, &tasks.ResumeTaskRequest{ContainerID: f.containerId}); err != nil {
		return err
	}

	return nil
}

func lookupContainerId(ctx context.Context, conn *grpc.ClientConn, podName, containerName string) (string, error) {
	client := cri.NewRuntimeServiceClient(conn)
	pods, err := client.ListPodSandbox(context.Background(), &cri.ListPodSandboxRequest{
		Filter: &cri.PodSandboxFilter{
			LabelSelector: map[string]string{
				"io.kubernetes.pod.name": podName,
			},
		},
	})
	if err != nil {
		return "", err
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("pod %s not found", podName)
	}
	pod := pods.Items[0]

	ctrs, err := client.ListContainers(ctx, &cri.ListContainersRequest{Filter: &cri.ContainerFilter{
		PodSandboxId: pod.Id,
		LabelSelector: map[string]string{
			"io.kubernetes.container.name": containerName,
		},
	}})
	if err != nil {
		return "", err
	}

	if len(ctrs.Containers) == 0 {
		return "", fmt.Errorf("pod %s not found", podName)
	}
	return ctrs.Containers[0].Id, nil
}
