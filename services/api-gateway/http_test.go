package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/http/httptest"
	pb "ride-sharing/shared/proto/trip"
	"strings"
	"testing"
)

type fakeTripClient struct {
	pb.TripServiceClient
	calls int
	err   error
}

func (f *fakeTripClient) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest, opts ...grpc.CallOption) (*pb.PreviewTripResponse, error) {
	f.calls++
	return &pb.PreviewTripResponse{}, f.err
}

func TestPreviewValidation(t *testing.T) {
	for _, body := range []string{
		`{}`, `null`, `{"userID":"rider"}`, `{invalid`,
		`{"userID":"rider","pickUp":{"latitude":91},"destination":{}}`,
		`{"userID":"rider","pickUp":{},"destination":{}} {}`,
		strings.Repeat(" ", (1<<20)+1),
	} {
		client := &fakeTripClient{}
		w := httptest.NewRecorder()
		NewTripHandler(client).handleTripPreview(w, httptest.NewRequest("POST", "/trip/preview", strings.NewReader(body)))
		if w.Code != http.StatusBadRequest || client.calls != 0 {
			t.Fatalf("status=%d calls=%d", w.Code, client.calls)
		}
	}
}

func TestPreviewMapsRPCFailures(t *testing.T) {
	for _, tc := range []struct {
		code codes.Code
		want int
	}{
		{codes.InvalidArgument, 400}, {codes.Unavailable, 503}, {codes.DeadlineExceeded, 504}, {codes.Internal, 500},
	} {
		client := &fakeTripClient{err: status.Error(tc.code, "private upstream details")}
		w := httptest.NewRecorder()
		NewTripHandler(client).handleTripPreview(w, httptest.NewRequest("POST", "/trip/preview", strings.NewReader(`{"userID":"rider","pickUp":{},"destination":{}}`)))
		if w.Code != tc.want || client.calls != 1 {
			t.Fatalf("status=%d calls=%d", w.Code, client.calls)
		}
		if strings.Contains(w.Body.String(), "private") {
			t.Fatal("leaked upstream details")
		}
	}
}

func TestCORSPreflight(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /trip/preview", func(http.ResponseWriter, *http.Request) { t.Fatal("preflight reached handler") })
	w := httptest.NewRecorder()
	enableCORS(mux.ServeHTTP)(w, httptest.NewRequest("OPTIONS", "/trip/preview", nil))
	if w.Code != 200 || w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Fatalf("preflight failed: %v", w.Result())
	}
}
