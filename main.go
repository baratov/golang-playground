package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("OK"))
}
