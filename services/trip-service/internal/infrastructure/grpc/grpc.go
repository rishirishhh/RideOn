package grpc

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer
	service domain.TripService
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}
	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateTrip not implemented")
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickup := req.GetStartLocation()
	destination := req.GetEndLocation()

	if pickup == nil || destination == nil || strings.TrimSpace(req.GetUserID()) == "" {
		return nil, status.Error(codes.InvalidArgument, "userID, pickup and destination are required")
	}
	pickupCoord := &types.Coordinate{
		Latitude:  pickup.Latitude,
		Longitude: pickup.Longitude,
	}
	destinationCoord := &types.Coordinate{
		Latitude:  destination.Latitude,
		Longitude: destination.Longitude,
	}

	if !pickupCoord.Valid() || !destinationCoord.Valid() {
		return nil, status.Error(codes.InvalidArgument, "invalid pickup or destination")
	}
	userID := req.GetUserID()
	t, err := h.service.GetRoute(ctx, pickupCoord, destinationCoord)
	if err != nil {
		return nil, rpcError(ctx, "failed to get route", err)
	}

	route, err := t.ToProto()
	if err != nil {
		return nil, rpcError(ctx, "invalid route", err)
	}
	estimatedFares, err := h.service.EstimatePackagesPriceWithRoute(t)
	if err != nil {
		return nil, rpcError(ctx, "failed to estimate fares", err)
	}
	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, userID)
	if err != nil {
		return nil, rpcError(ctx, "failed to save fares", err)
	}
	return &pb.PreviewTripResponse{
		Route:     route,
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}

func rpcError(ctx context.Context, message string, err error) error {
	log.Printf("%s: %v", message, err)
	if ctx.Err() != nil {
		return status.FromContextError(ctx.Err()).Err()
	}
	return status.Error(codes.Internal, message)
}
