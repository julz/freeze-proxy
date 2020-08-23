package main

import (
	"context"

	"github.com/julz/pauseme/pkg/mutations"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
)

var types = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	v1.SchemeGroupVersion.WithKind("Pod"): &mutations.Pod{},
}

func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return defaulting.NewAdmissionController(ctx,
		// Name of the resource webhook.
		"webhook.freeze.extensions.knative.dev",
		// The path on which to serve the webhook.
		"/defaulting",
		// The resources to validate and default.
		types,
		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},
		// Whether to disallow unknown fields.
		true,
	)
}

func main() {
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: "freezer-webhook",
		SecretName:  "freezer-webhook-certs",
		Port:        8443,
	})

	sharedmain.WebhookMainWithContext(
		ctx, "freezer-webhook",
		certificates.NewController,
		NewController,
	)
}