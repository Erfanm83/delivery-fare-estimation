package main

import (
	"fmt"
	"math"
)

func main() {

	fmt.Printf("Project Initialized \n")
	fmt.Printf("haversine distance : %f\n", haversine(51.5007,
		0.1246,
		40.6892,
		74.0445))

}

// haversine function to calculate distance
func haversine(lat1, lon1, lat2, lon2 float64) float64 {

	// distance between latitude and longitudes
	deltaLat := (lat2 - lat1) * math.Pi / 180.0
	deltaLon := (lon2 - lon1) * math.Pi / 180.0

	// Convert to radians
	latRad1 := lat1 * math.Pi / 180.0
	latRad2 := lat2 * math.Pi / 180.0

	// formul
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(latRad1)*math.Cos(latRad2)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	radiusEarthKm := 6371.0

	return radiusEarthKm * c
}
