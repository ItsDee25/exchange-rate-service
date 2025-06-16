package main

import (
	"github.com/ItsDee25/exchange-rate-service/cmd/server/bootstrap"
)

func main() { // Initialize the server

	// set env variables
	bootstrap.InitServer()
}
