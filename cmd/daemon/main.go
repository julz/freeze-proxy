package main

import (
	"log"
	"net/http"

	"github.com/julz/freeze-proxy/pkg/freezer"
	"go.uber.org/zap"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	sugared := logger.Sugar()
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("hello from freeze daemon, token:", r.Header.Get("Token"))
		result, err := clientset.AuthenticationV1().TokenReviews().Create(&authv1.TokenReview{
			Spec: authv1.TokenReviewSpec{
				Token: r.Header.Get("Token"),
				Audiences: []string{
					"freeze",
				},
			},
		})
		if err != nil {
			log.Fatal(err)
		}

		podName := result.Status.User.Extra["authentication.kubernetes.io/pod-name"][0]
		log.Println("Freeze pod:", podName)

		freezer, err := freezer.Connect(sugared, podName, "user-container")
		if err != nil {
			panic(err)
		}

		freezer.Freeze(r.Context())
		log.Println("freezed!")
	}))
}
