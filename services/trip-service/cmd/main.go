package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/infrastructure/routing"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
	"syscall"
	"time"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9093"

func main() {
	inmemRepo := repository.NewInmemRepository()
	router := routing.NewOSRM(&http.Client{Timeout: 4 * time.Second}, env.GetString("OSRM_URL", "http://router.project-osrm.org"))
	svc := service.NewService(inmemRepo, router)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpcserver.NewServer()

	grpc.NewGRPCHandler(grpcServer, svc)
	log.Printf("Starting gRPC server Trip service on port %s", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	grpcServer.GracefulStop()
}
