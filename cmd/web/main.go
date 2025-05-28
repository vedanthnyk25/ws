package main

import (
	"log"
	"net/http"
	"os"
	"ws/internal/handlers"
)

func main() {
	mux := routes()

	log.Println("Listening to the channel")
	go handlers.ListenToWsChannel()

	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default for local dev
	}

	log.Println("Starting web server on port", port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
