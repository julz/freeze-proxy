module github.com/julz/freeze-proxy

go 1.15

require (
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/Microsoft/hcsshim/test v0.0.0-20200818230740-94556e86d3db // indirect
	github.com/containerd/cgroups v0.0.0-20200817152742-7a3c009711fb // indirect
	github.com/containerd/containerd v1.4.0
	github.com/containerd/continuity v0.0.0-20200710164510-efbc4488d8fe // indirect
	github.com/containerd/fifo v0.0.0-20200410184934-f15a3290365b // indirect
	github.com/containerd/ttrpc v1.0.1 // indirect
	github.com/containerd/typeurl v1.0.1 // indirect
	github.com/docker/docker v20.10.5+incompatible
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/selinux v1.6.0 // indirect
	github.com/rogpeppe/go-internal v1.6.0 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	go.uber.org/atomic v1.6.0
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/sys v0.0.0-20200821140526-fda516888d29 // indirect
	google.golang.org/grpc v1.31.0
	gotest.tools/v3 v3.0.2 // indirect
	k8s.io/api v0.18.7-rc.0
	k8s.io/apimachinery v0.18.7-rc.0
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/cri-api v0.18.8
	k8s.io/klog v1.0.0
	knative.dev/pkg v0.0.0-20200822052046-d5c09d2aef18
)

replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/code-generator => k8s.io/code-generator v0.17.6
)
