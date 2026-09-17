package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/env"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8081")
)

func main() {
	log.Println("Starting API Gateway")
	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer tripService.Close()

	handler := NewTripHandler(tripService.Client)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /trip/preview", handler.handleTripPreview)
	mux.HandleFunc("POST /trip/start", handler.handleTripStart)
	mux.HandleFunc("/ws/drivers", handleDriveWebsocket)
	mux.HandleFunc("/ws/riders", handleRidersWebsocket)

	server := &http.Server{
		Addr:              httpAddr,
		Handler:           enableCORS(mux.ServeHTTP),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Server listening on %s", httpAddr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Printf("Error starting the server: %v", err)
	case sig := <-shutdown:
		log.Printf("Server is shutting down due to %v signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Could not stop ")
			server.Close()
		}
	}
}
