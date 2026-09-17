package service

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
	tripTypes "ride-sharing/services/trip-service/pkg/types"

	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repo   domain.TripRepository
	router RouteProvider
}

// RouteProvider isolates routing I/O from trip business logic.
type RouteProvider interface {
	GetRoute(context.Context, *types.Coordinate, *types.Coordinate) (*tripTypes.OsrmApiResponse, error)
}

func NewService(repo domain.TripRepository, router RouteProvider) *Service {
	return &Service{repo: repo, router: router}
}

func (s *Service) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	if fare == nil {
		return nil, fmt.Errorf("ride fare is required")
	}
	t := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
	}

	return s.repo.CreateTrip(ctx, t)
}

func (s *Service) GetRoute(
	ctx context.Context,
	pickup, destination *types.Coordinate,
) (*tripTypes.OsrmApiResponse, error) {

	if !pickup.Valid() || !destination.Valid() {
		return nil, fmt.Errorf("invalid pickup or destination")
	}
	route, err := s.router.GetRoute(ctx, pickup, destination)
	if err != nil {
		return nil, err
	}
	if err := route.Validate(); err != nil {
		return nil, err
	}
	return route, nil
}

func (s *Service) EstimatePackagesPriceWithRoute(route *tripTypes.OsrmApiResponse) ([]*domain.RideFareModel, error) {
	if err := route.Validate(); err != nil {
		return nil, err
	}
	baseFares := getBaseFares()

	estimatedFares := make([]*domain.RideFareModel, len(baseFares))

	for i, f := range baseFares {
		estimatedFares[i] = estimateFareRoute(f, route)
	}
	return estimatedFares, nil
}

func (s *Service) GenerateTripFares(ctx context.Context, rideFares []*domain.RideFareModel, userID string) ([]*domain.RideFareModel, error) {
	fares := make([]*domain.RideFareModel, len(rideFares))

	for i, f := range rideFares {
		id := primitive.NewObjectID()

		fare := &domain.RideFareModel{
			UserID:            userID,
			ID:                id,
			TotalPriceInCents: f.TotalPriceInCents,
			PackageSlug:       f.PackageSlug,
		}

		if err := s.repo.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("failed to save trip fare: %w", err)
		}

		fares[i] = fare
	}

	return fares, nil
}

func estimateFareRoute(f *domain.RideFareModel, route *tripTypes.OsrmApiResponse) *domain.RideFareModel {
	pricingCfg := tripTypes.DefaultPricingConfig()
	carPackagePrice := f.TotalPriceInCents

	distanceKm := route.Routes[0].Distance
	durationInMinutes := route.Routes[0].Duration

	distanceFare := distanceKm * pricingCfg.PricePerUnitOfDistance
	timeFare := durationInMinutes * pricingCfg.PricingPerMinute

	totalPrice := carPackagePrice + distanceFare + timeFare

	return &domain.RideFareModel{
		TotalPriceInCents: totalPrice,
		PackageSlug:       f.PackageSlug,
	}
}

func getBaseFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{
			PackageSlug:       "suv",
			TotalPriceInCents: 200,
		},
		{
			PackageSlug:       "sedan",
			TotalPriceInCents: 350,
		},
		{
			PackageSlug:       "van",
			TotalPriceInCents: 400,
		},
		{
			PackageSlug:       "luxury",
			TotalPriceInCents: 1000,
		},
	}
}
