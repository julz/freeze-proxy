package gate

import (
	"fmt"
	"net/http"
)

// New creates a http handler that will call a pause function when request
// count hits zero, and a resume function when request count increases to 1.
func New(delegate http.Handler, pause, resume func()) http.HandlerFunc {
	type req struct {
		w http.ResponseWriter
		r *http.Request

		done chan struct{}
	}

	reqCh := make(chan req)
	doneCh := make(chan struct{})
	go func() {
		inFlight := 0

		// this loop is entirely synchronous, so there's no cleverness needed in
		// ensuring open and close dont run at the same time etc. Only the
		// delegated ServeHTTP is done in a goroutine.
		for {
			select {
			case <-doneCh:
				inFlight--
				fmt.Println("inFlight after request",inFlight)
				if inFlight == 0 {
					pause()
				}

			case r := <-reqCh:
				inFlight++
				fmt.Println("inFlight before request",inFlight)
				if inFlight == 1 {
					resume()
				}

				go func(r req) {

					fmt.Println("proxy request",inFlight)

					delegate.ServeHTTP(r.w, r.r)
					close(r.done) // return from ServeHTTP.
					doneCh <- struct{}{}
				}(r)
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		done := make(chan struct{})
		fmt.Println("new request... ")
		reqCh <- req{w, r, done}
		// block till we're processed
		<-done
	}
}
