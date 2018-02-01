package server

import (
	"net/http"
	"fmt"
	"time"
	"log"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func logging(h http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		end := time.Now()
		log.Printf("%s: %d", r.URL.String(), end.Sub(start).Nanoseconds())
	}
	return http.HandlerFunc(f)
}

func Serve() {
	http.Handle("/", logging(http.HandlerFunc(handler)))
	http.ListenAndServe(":8080", nil)
}