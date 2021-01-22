package daemon

import (
	"context"
	"net/http"

	authv1 "k8s.io/api/authentication/v1"
)

const TokenHeaderKey = "Token"

type TokenValidator interface {
	Validate(ctx context.Context, token string) (*authv1.TokenReview, error)
}

type Freezer interface {
	Freeze(ctx context.Context, podName, containerName string) error
}

type Thawer interface {
	Thaw(ctx context.Context, podName, containerName string) error
}

type Handler struct {
	Validator TokenValidator
	Freezer   Freezer
	Thawer    Thawer
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(TokenHeaderKey)
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := h.Validator.Validate(r.Context(), token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !resp.Status.Authenticated {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	switch r.URL.Path {
	case "/freeze":
		h.Freezer.Freeze(r.Context(), resp.Status.User.Extra["authentication.kubernetes.io/pod-name"][0], "user-container")
	case "/thaw":
		h.Thawer.Thaw(r.Context(), resp.Status.User.Extra["authentication.kubernetes.io/pod-name"][0], "user-container")
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

type TokenValidatorFunc func(ctx context.Context, token string) (*authv1.TokenReview, error)

func (fn TokenValidatorFunc) Validate(ctx context.Context, token string) (*authv1.TokenReview, error) {
	return fn(ctx, token)
}
