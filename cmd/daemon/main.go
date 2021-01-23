package main

import (
	"context"
	"log"
	"net/http"

	"github.com/julz/freeze-proxy/pkg/daemon"
	"github.com/julz/freeze-proxy/pkg/freezer"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	ctrd, err := freezer.Connect()
	if err != nil {
		log.Fatal(err)
	}

	http.ListenAndServe(":8080", &daemon.Handler{
		Freezer: ctrd,
		Thawer:  ctrd,
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
