package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julz/pauseme/pkg/freezer"
	"github.com/julz/pauseme/pkg/handler"
	"github.com/prometheus/common/log"
	"go.uber.org/zap"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	sugared := logger.Sugar()
	freezer, err := freezer.Connect(sugared, os.Getenv("POD_NAME"), "user-container")
	if err != nil {
		panic(err)
	}

	pause := func() {
		if err := freezer.Freeze(context.Background()); err != nil {
			panic(err)
		}

		log.Info("paused")
	}

	resume := func() {
		if err := freezer.Thaw(context.Background()); err != nil {
			panic(err)
		}

		log.Info("resumed")
	}

	// make sure we resume when we're going to be killed so that the user
	// container can be killed normally.
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		resume()
		os.Exit(0)
	}()

	// start paused.
	pause()

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: "localhost:8080"})
	proxy.FlushInterval = 25 * time.Millisecond

	http.ListenAndServe(":9999", handler.New(proxy, pause, resume))
}
