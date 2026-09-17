package routing

import (
	"context"
	"io"
	"net/http"
	"ride-sharing/shared/types"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestOSRMResponses(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		wantErr    bool
	}{
		{"valid", `{"code":"Ok","routes":[{"distance":1200,"duration":300,"geometry":{"coordinates":[[77,12],[78,13]]}}]}`, 200, false},
		{"empty routes", `{"code":"NoRoute","routes":[]}`, 200, true},
		{"malformed geometry", `{"routes":[{"geometry":{"coordinates":[[77]]}}]}`, 200, true},
		{"negative distance", `{"routes":[{"distance":-1}]}`, 200, true},
		{"invalid json", `{`, 200, true},
		{"upstream error", `unavailable`, 503, true},
		{"oversized response", strings.Repeat(" ", (4<<20)+1), 200, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if !strings.Contains(r.URL.Path, "77.000000,12.000000;78.000000,13.000000") {
					t.Errorf("wrong coordinate order: %s", r.URL.Path)
				}
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}
			router := NewOSRM(client, "http://routing.test/")
			route, err := router.GetRoute(context.Background(), &types.Coordinate{Latitude: 12, Longitude: 77}, &types.Coordinate{Latitude: 13, Longitude: 78})
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tc.wantErr)
			}
			if err == nil {
				proto, err := route.ToProto()
				if err != nil {
					t.Fatal(err)
				}
				coord := proto.Geometry[0].Coordinates[0]
				if coord.Latitude != 12 || coord.Longitude != 77 {
					t.Fatalf("incorrect geometry: %v", coord)
				}
			}
		})
	}
}

func TestOSRMCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) { return nil, r.Context().Err() })}
	_, err := NewOSRM(client, "http://routing.test").GetRoute(ctx, &types.Coordinate{}, &types.Coordinate{})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}
