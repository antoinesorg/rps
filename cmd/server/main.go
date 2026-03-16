package main

import (
	"fmt"
	"net/http"
	"os"

	"rps/internal/game"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dataDir := "var/data"
	svc, err := game.NewService(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init service: %v\n", err)
		os.Exit(1)
	}

	handler := game.NewHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /games", handler.HealthCheck)
	mux.HandleFunc("POST /games", handler.CreateGame)
	mux.HandleFunc("POST /games/{id}/moves", handler.PostMove)

	addr := ":" + port
	fmt.Fprintf(os.Stderr, "listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
