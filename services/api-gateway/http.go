package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ride-sharing/shared/contracts"
	pb "ride-sharing/shared/proto/trip"
)

type tripHandler struct {
	tripService pb.TripServiceClient
}

func NewTripHandler(tripService pb.TripServiceClient) *tripHandler {
	return &tripHandler{tripService: tripService}
}

func (h *tripHandler) handleTripStart(w http.ResponseWriter, r *http.Request) {
	var reqBody startTripRequest
	if err := readJSON(w, r, &reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if strings.TrimSpace(reqBody.UserID) == "" || strings.TrimSpace(reqBody.RideFareID) == "" {
		http.Error(w, "userID and rideFareID are required", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	trip, err := h.tripService.CreateTrip(ctx, reqBody.toProto())
	if err != nil {
		log.Printf("Failed to start a trip: %v", err)
		http.Error(w, "Failed to start trip", httpStatusFromRPC(err))
		return
	}

	response := contracts.APIResponse{Data: trip}
	writeJSON(w, http.StatusCreated, response)
}

func (h *tripHandler) handleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := readJSON(w, r, &reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if strings.TrimSpace(reqBody.UserID) == "" || !reqBody.Pickup.Valid() || !reqBody.Destination.Valid() {
		http.Error(w, "userID, pickup and destination must be valid", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tripPreview, err := h.tripService.PreviewTrip(ctx, reqBody.toProto())
	if err != nil {
		log.Printf("Failed to preview a trip: %v", err)
		http.Error(w, "Failed to preview trip", httpStatusFromRPC(err))
		return
	}

	response := contracts.APIResponse{Data: tripPreview}
	writeJSON(w, http.StatusCreated, response)
}

func httpStatusFromRPC(err error) int {
	switch status.Code(err) {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.Unimplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}
