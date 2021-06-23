package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/julz/freeze-proxy/pkg/daemon"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/julz/freeze-proxy/pkg/freezer"
)

var runtimeType string

func init() {
	runtimeType = os.Getenv("RUNTIME_TYPE")
}
func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	fre, err := freezer.GetFreezer(runtimeType)
	if err != nil {
		log.Fatal(err)
	}

	thawer, err := freezer.GetThawer(runtimeType)
	if err != nil {
		log.Fatal(err)
	}

	http.ListenAndServe(":8080", &daemon.Handler{
		Freezer: fre,
		Thawer:  thawer,
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
