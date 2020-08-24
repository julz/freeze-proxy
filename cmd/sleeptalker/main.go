package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	go func() {
		for range time.Tick(500 * time.Millisecond) {
			fmt.Printf("Ticking at %s.\n", time.Now().Format("3h 04m 05.000s"))
		}
	}()

	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, this request will take two seconds..")
		w.(http.Flusher).Flush()
		time.Sleep(2 * time.Second)
		fmt.Fprintln(w, ".. world.")
	}))
}
