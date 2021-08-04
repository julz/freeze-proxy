package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/julz/freeze-proxy/pkg/gate"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func main() {
	hostIP := os.Getenv("HOST_IP")

	log.Println("Connect to freeze daemon on:", hostIP)

	var tokenCfg Token
	refreshToken(&tokenCfg)
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			refreshToken(&tokenCfg)
		}
	}()

	pause := func() {
		req, err := http.NewRequest("POST", "http://"+hostIP+":9696/freeze", nil)
		if err != nil {
			panic(err)
		}

		req.Header.Add("Token", tokenCfg.Get())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}

		log.Println("sent pause request, got", resp.StatusCode)
	}

	resume := func() {
		req, err := http.NewRequest("POST", "http://"+hostIP+":9696/thaw", nil)
		if err != nil {
			panic(err)
		}

		req.Header.Add("Token", tokenCfg.Get())
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

	http.ListenAndServe(":9999", gate.New(proxy, pause, resume))
}

type Token struct {
	sync.RWMutex
	token string
}

func (t *Token) Set(token string) {
	t.Lock()
	defer t.Unlock()

	t.token = token
}
func (t *Token) Get() string {
	t.RLock()
	defer t.RUnlock()

	return t.token
}

func refreshToken(tokenCfg *Token) {
	token, err := ioutil.ReadFile("/var/run/projected/token")
	if err != nil {
		log.Fatal("could not read token", err)
	}
	tokenCfg.Set(string(token))
	log.Println("refresh token...")
}
