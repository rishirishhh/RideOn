package routing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	"ride-sharing/shared/types"
	"strings"
)

// OSRM owns the external routing protocol; the service only depends on GetRoute.
type OSRM struct {
	client  *http.Client
	baseURL string
}

func NewOSRM(client *http.Client, baseURL string) *OSRM {
	return &OSRM{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (o *OSRM) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error) {
	if !pickup.Valid() || !destination.Valid() {
		return nil, fmt.Errorf("invalid pickup or destination")
	}
	url := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson", o.baseURL,
		pickup.Longitude, pickup.Latitude, destination.Longitude, destination.Latitude)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create routing request: %w", err)
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call routing provider: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("routing provider returned HTTP %d", resp.StatusCode)
	}
	const maxResponseBytes = 4 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read routing response: %w", err)
	}
	if len(body) > maxResponseBytes {
		return nil, fmt.Errorf("routing response exceeds size limit")
	}
	var route tripTypes.OsrmApiResponse
	if err := json.Unmarshal(body, &route); err != nil {
		return nil, fmt.Errorf("decode routing response: %w", err)
	}
	if err := route.Validate(); err != nil {
		return nil, err
	}
	return &route, nil
}
