package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"ride-sharing/services/trip-service/internal/domain"
	"sync"
	"testing"
)

func TestConcurrentWrites(t *testing.T) {
	repo := NewInmemRepository()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fare := &domain.RideFareModel{ID: primitive.NewObjectID(), UserID: "rider"}
			trip := &domain.TripModel{ID: primitive.NewObjectID(), RideFare: fare}
			if err := repo.SaveRideFare(context.Background(), fare); err != nil {
				t.Error(err)
			}
			if _, err := repo.CreateTrip(context.Background(), trip); err != nil {
				t.Error(err)
			}
			// Caller mutations must not change stored data.
			fare.UserID = "changed"
			trip.Status = "changed"
		}()
	}
	wg.Wait()
	if len(repo.trips) != 100 || len(repo.rideFares) != 100 {
		t.Fatal("lost writes")
	}
	for _, trip := range repo.trips {
		if trip.Status == "changed" || trip.RideFare.UserID != "rider" {
			t.Fatal("repository retained caller-owned state")
		}
	}
	for _, fare := range repo.rideFares {
		if fare.UserID != "rider" {
			t.Fatal("repository retained caller-owned fare")
		}
	}
}

func TestCanceledWrite(t *testing.T) {
	repo := NewInmemRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := repo.SaveRideFare(ctx, &domain.RideFareModel{}); err != context.Canceled {
		t.Fatalf("error = %v", err)
	}
	if len(repo.rideFares) != 0 {
		t.Fatal("canceled operation persisted a fare")
	}
}
