package service

import (
	"context"
	"encoding/json"
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	"ride-sharing/shared/types"
	"testing"
)

type fakeRouter struct {
	route *tripTypes.OsrmApiResponse
	calls int
}

func (f *fakeRouter) GetRoute(context.Context, *types.Coordinate, *types.Coordinate) (*tripTypes.OsrmApiResponse, error) {
	f.calls++
	return f.route, nil
}

func TestGetRouteValidatesProviderResult(t *testing.T) {
	router := &fakeRouter{}
	svc := NewService(nil, router)
	if _, err := svc.GetRoute(context.Background(), &types.Coordinate{}, &types.Coordinate{}); err == nil {
		t.Fatal("expected error for nil route")
	}
	if router.calls != 1 {
		t.Fatal("routing dependency was not called")
	}
	if _, err := svc.GetRoute(context.Background(), nil, &types.Coordinate{}); err == nil {
		t.Fatal("expected input error")
	}
	if router.calls != 1 {
		t.Fatal("invalid input reached routing provider")
	}
}

func TestEstimateRequiresValidRoute(t *testing.T) {
	svc := NewService(nil, nil)
	for _, route := range []*tripTypes.OsrmApiResponse{nil, {}} {
		if _, err := svc.EstimatePackagesPriceWithRoute(route); err == nil {
			t.Fatal("expected error for missing route")
		}
	}
	var route tripTypes.OsrmApiResponse
	if err := json.Unmarshal([]byte(`{"routes":[{"distance":1000,"duration":60}]}`), &route); err != nil {
		t.Fatal(err)
	}
	fares, err := svc.EstimatePackagesPriceWithRoute(&route)
	if err != nil || len(fares) != 4 {
		t.Fatalf("fares=%v error=%v", fares, err)
	}
}
