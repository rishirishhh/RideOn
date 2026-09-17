package types

import (
	"fmt"
	"math"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
)

type OsrmApiResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"routes"`
}

// Validate checks the selected route before pricing or converting its geometry.
func (o *OsrmApiResponse) Validate() error {
	if o == nil || len(o.Routes) == 0 {
		return fmt.Errorf("routing response contains no routes")
	}
	if o.Code != "" && o.Code != "Ok" {
		return fmt.Errorf("routing response code: %s", o.Code)
	}
	route := o.Routes[0]
	if math.IsNaN(route.Distance) || math.IsInf(route.Distance, 0) || route.Distance < 0 ||
		math.IsNaN(route.Duration) || math.IsInf(route.Duration, 0) || route.Duration < 0 {
		return fmt.Errorf("routing response contains invalid distance or duration")
	}
	for _, coord := range route.Geometry.Coordinates {
		if len(coord) < 2 || !(&types.Coordinate{Longitude: coord[0], Latitude: coord[1]}).Valid() {
			return fmt.Errorf("routing response contains invalid geometry")
		}
	}
	return nil
}

func (o *OsrmApiResponse) ToProto() (*pb.Route, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}
	route := o.Routes[0]
	geometry := route.Geometry.Coordinates
	coordinates := make([]*pb.Coordinate, len(geometry))
	for i, coord := range geometry {
		coordinates[i] = &pb.Coordinate{
			Latitude:  coord[1],
			Longitude: coord[0],
		}
	}

	return &pb.Route{
		Geometry: []*pb.Geometry{
			{
				Coordinates: coordinates,
			},
		},
		Distance: route.Distance,
		Duration: route.Duration,
	}, nil
}

type PricingConfig struct {
	PricePerUnitOfDistance float64
	PricingPerMinute       float64
}

func DefaultPricingConfig() *PricingConfig {
	return &PricingConfig{
		PricePerUnitOfDistance: 1.5,
		PricingPerMinute:       0.25,
	}
}
