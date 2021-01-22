package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julz/freeze-proxy/pkg/handler"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func main() {
	hostIP := os.Getenv("HOST_IP")

	log.Println("Connect to freeze daemon on:", hostIP)

	// todo: reload every few minutes
	token, err := ioutil.ReadFile("/var/run/projected/token")
	if err != nil {
		log.Fatal("could not read token", err)
	}

	log.Println("token:", string(token))

	pause := func() {
		req, err := http.NewRequest("POST", "http://"+hostIP+":9696/freeze", nil)
		if err != nil {
			panic(err)
		}

		req.Header.Add("Token", string(token))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}

		log.Println("sent pause request, got", resp.StatusCode)
	}

	resume := func() {
		req, err := http.NewRequest("POST", "http://"+hostIP+":9696/resume", nil)
		if err != nil {
			panic(err)
		}

		req.Header.Add("Token", string(token))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}

		log.Println("sent resume request, got", resp.StatusCode)
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
