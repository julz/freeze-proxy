package mutations

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
)

var (
	_ apis.Defaultable = (*Pod)(nil)
	_ apis.Validatable = (*Pod)(nil)
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Pod extends a pod to be freezable.
type Pod struct {
	v1.Pod `json:",inline"`
}

// SetDefaults mutates the pod to add the freeze proxy.
func (p *Pod) SetDefaults(ctx context.Context) {
	fmt.Println("mutating pod", p.Name)

	// let's not worry about updates etc, for now
	if _, ok := p.Pod.Labels["freeze-pod-added"]; ok {
		return
	}

	p.Pod.Labels["freeze-pod-added"] = "true"
	p.Pod.Labels["freeze-pod-version"] = "2"
	p.Pod.Spec.Volumes = append(p.Pod.Spec.Volumes, v1.Volume{
		Name: "containerd-socket",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: "/var/run/containerd/containerd.sock",
				Type: &socketType,
			},
		},
	})

	for i, _ := range p.Pod.Spec.Containers {
		if p.Pod.Spec.Containers[i].Name == "queue-proxy" {
			for j := range p.Pod.Spec.Containers[i].Env {
				if p.Pod.Spec.Containers[i].Env[j].Name == "USER_PORT" {
					p.Pod.Spec.Containers[i].Env[j].Value = "9999"
				}
			}
		}
	}

	p.Pod.Spec.Containers = append(p.Pod.Spec.Containers, v1.Container{
		Name:  "freeze-proxy",
		Image: os.Getenv("FREEZE_PROXY_IMAGE"),
		Env: []v1.EnvVar{{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		}},
		VolumeMounts: []v1.VolumeMount{{
			Name: "containerd-socket",
			// TODO(jz):
			// rather than mounting containerd socket, mount a socket from an
			// intermediate daemon. This would greatly lower the surface area we're
			// exposing in to containers, and avoid needing to run the freeze proxy
			// as root.
			MountPath: "/var/run/containerd/containerd.sock",
		}},
	})
}

// Validate returns nil due to no need for validation
func (p *Pod) Validate(ctx context.Context) *apis.FieldError {
	// TODO(jz) reject any non-freeze containers that try to mount the socket themselves etc.
	// TODO(jz) reject any non-freeze containers that try to run as root.

	return nil
}

var socketType = v1.HostPathSocket
