package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	go func() {
		// prints all the time in the background, even when no requests are being
		// processed - naughty!.
		for range time.Tick(500 * time.Millisecond) {
			fmt.Printf("Ticking at %s.\n", time.Now().Format("3h 04m 05.000s"))
		}
	}()

	// simple handler, just sleeps 2 seconds - backgrond tasks allowed to run
	// while this is going.
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, this request will take two seconds..")
		w.(http.Flusher).Flush()
		time.Sleep(2 * time.Second)
		fmt.Fprintln(w, ".. world.")
	}))
}
