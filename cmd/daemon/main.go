package main

import (
	"context"
	"log"
	"net/http"
	"os"

	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/julz/freeze-proxy/pkg/daemon"
	"github.com/julz/freeze-proxy/pkg/freeze/containerd"
	"github.com/julz/freeze-proxy/pkg/freeze/docker"
)

const (
	runtimeTypeDocker     string = "docker"
	runtimeTypeContainerd string = "containerd"
	runtimeTypeCriO       string = "cri-o"
)

func main() {
	runtimeType := os.Getenv("RUNTIME_TYPE")

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	var freezeThaw daemon.FreezeThawer
	switch runtimeType {
	case runtimeTypeDocker:
		freezeThaw, err = docker.New()
	case runtimeTypeContainerd:
		freezeThaw, err = containerd.New()
		// TODO suport crio
	default:
		log.Fatal("unrecognised runtimeType", runtimeType)
	}
	if err != nil {
		log.Fatal(err)
	}

	http.ListenAndServe(":8080", &daemon.Handler{
		Freezer: freezeThaw,
		Thawer:  freezeThaw,
		Validator: daemon.TokenValidatorFunc(func(ctx context.Context, token string) (*authv1.TokenReview, error) {
			return clientset.AuthenticationV1().TokenReviews().CreateContext(ctx, &authv1.TokenReview{
				Spec: authv1.TokenReviewSpec{
					Token: token,
					Audiences: []string{
						// The projected token only gives the right to freeze/unfreeze.
						"freeze",
					},
				},
			})
		}),
	})
}
