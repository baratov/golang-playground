package main

import (
	"github.com/baratov/golang-playground/server"
)

func main() {
	// config values as PORT should be taken from ENV VARs
	server.Serve("8080", false)
}
