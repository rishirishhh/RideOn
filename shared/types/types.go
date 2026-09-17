package types

import "math"

type Route struct {
	Distance float64     `json:"distance"`
	Duration float64     `json:"duration"`
	Geometry []*Geometry `json:"geometry"`
}

type Geometry struct {
	Coordinates []*Coordinate `json:"coordinates"`
}

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Valid reports whether the coordinate is finite and within geographic bounds.
func (c *Coordinate) Valid() bool {
	return c != nil && !math.IsNaN(c.Latitude) && !math.IsNaN(c.Longitude) &&
		c.Latitude >= -90 && c.Latitude <= 90 && c.Longitude >= -180 && c.Longitude <= 180
}
