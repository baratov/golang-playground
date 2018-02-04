package server

import (
	"net/http"
	"fmt"
	"time"
)

type middleware func(http.Handler) http.Handler

func chain(h http.Handler, mw []middleware) http.Handler {
	for _, middleware := range mw{
		h = middleware(h)
	}
	return h
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func logging() middleware{
	return func(h http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			fmt.Fprintln(w, start)
			defer func() {
				end := time.Now()
				fmt.Fprintln(w, end)
			}()
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(f)
	}
}

func oneMore() middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
			fmt.Fprintln(w, "oneMore before")
			defer fmt.Fprintln(w, "oneMore after")
			h.ServeHTTP(w, r)
		})
	}
}

func Serve() {
	commonChain := []middleware {logging(), oneMore()}
	http.Handle("/", chain(http.HandlerFunc(handler), commonChain))
	http.ListenAndServe(":8080", nil)
}