package mutations

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/ptr"
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
		Name: "projected",
		VolumeSource: v1.VolumeSource{
			Projected: &v1.ProjectedVolumeSource{
				Sources: []v1.VolumeProjection{{
					ServiceAccountToken: &v1.ServiceAccountTokenProjection{
						Audience:          "freeze",
						Path:              "token",
						ExpirationSeconds: ptr.Int64(60 * 10),
					},
				},
				},
			},
		},
	})

	userContainerName := "user-container"
	for i := range p.Pod.Spec.Containers {
		if p.Pod.Spec.Containers[i].Name == "queue-proxy" {
			for j := range p.Pod.Spec.Containers[i].Env {
				if p.Pod.Spec.Containers[i].Env[j].Name == "USER_PORT" {
					p.Pod.Spec.Containers[i].Env[j].Value = "9999"
				}
			}
		} else if len(p.Pod.Spec.Containers) == 2 {
			// if we're not in the multi-container path, we can easily figure out which
			// container is the user container because it's whichever isn't the QP.
			userContainerName = p.Pod.Spec.Containers[i].Name
		}
	}

	p.Pod.Spec.Containers = append(p.Pod.Spec.Containers, v1.Container{
		Name:  "freeze-proxy",
		Image: os.Getenv("FREEZE_PROXY_IMAGE"),
		Env: []v1.EnvVar{{
			Name: "HOST_IP",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			},
		}, {
			Name:  "USER_CONTAINER",
			Value: userContainerName,
		}},
		VolumeMounts: []v1.VolumeMount{{
			Name:      "projected",
			MountPath: "/var/run/projected",
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
