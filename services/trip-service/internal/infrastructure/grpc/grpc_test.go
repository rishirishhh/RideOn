package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "ride-sharing/shared/proto/trip"
	"testing"
)

func TestPreviewRejectsInvalidInput(t *testing.T) {
	handler := &gRPCHandler{}
	for _, req := range []*pb.PreviewTripRequest{
		nil, {},
		{UserID: "rider"},
		{UserID: " ", StartLocation: &pb.Coordinate{}, EndLocation: &pb.Coordinate{}},
		{UserID: "rider", StartLocation: &pb.Coordinate{Latitude: 91}, EndLocation: &pb.Coordinate{}},
	} {
		_, err := handler.PreviewTrip(context.Background(), req)
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("got %v, want InvalidArgument", err)
		}
	}
}
