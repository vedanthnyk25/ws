package main

import (
	"log"
	"net/http"
	"ws/internal/handlers"
)

func main() {
	mux := routes()

	log.Println("Listening to the channel")
	go handlers.ListenToWsChannel()

	log.Println("Staring web server on port 8080")

	_ = http.ListenAndServe(":8080", mux)
}
