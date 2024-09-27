package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHaversine(t *testing.T) {
	// Testing with known geographical coordinates
	result := haversine(51.5007, 0.1246, 40.6892, 74.0445) // London to New York
	expected := 5575.0                                     // expected distance in kilometers

	// The distance should be accurate within a small error
	assert.InEpsilon(t, expected, result, 0.01, "Haversine function should calculate correct distance")
}

func TestReadDataChunks(t *testing.T) {
	//assuming "3" exists at line 58
	startRow := 58
	deliveryID := "3"

	points, err := readDataChunks(deliveryID, startRow)
	assert.NoError(t, err, "Failed to read data chunks from file")
	assert.NotEmpty(t, points, "No points were read; expected at least one matching point")

	if len(points) > 0 {
		assert.Equal(t, deliveryID, points[0].ID, "The first point should have the correct delivery ID")
	}
}

func TestFilterInvalidPoints(t *testing.T) {
	points := []DeliveryPoint{
		{ID: "1", Latitude: 35.0, Longitude: 40.0, Timestamp: 1000},
		{ID: "1", Latitude: 35.0, Longitude: 40.0, Timestamp: 1300}, // Within valid speed limits
		{ID: "1", Latitude: 36.0, Longitude: 41.0, Timestamp: 2000}, // Exceeds valid speed limits
	}

	filteredPoints := filterInvalidPoints(points)
	assert.Len(t, filteredPoints, 2, "Expected only valid points to be returned")
	assert.Equal(t, int64(1300), filteredPoints[1].Timestamp, "Filtered points should include the point with valid speed")
}
