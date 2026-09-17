package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
	"sync"
)

type inmemRepository struct {
	mu        sync.Mutex
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInmemRepository() *inmemRepository {
	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inmemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, fmt.Errorf("trip is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *trip
	if trip.RideFare != nil {
		fare := *trip.RideFare
		stored.RideFare = &fare
	}
	r.trips[trip.ID.Hex()] = &stored
	return trip, nil
}

func (r *inmemRepository) SaveRideFare(ctx context.Context, f *domain.RideFareModel) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if f == nil {
		return fmt.Errorf("ride fare is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *f
	r.rideFares[f.ID.Hex()] = &stored
	return nil
}
