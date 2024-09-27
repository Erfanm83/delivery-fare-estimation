package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHaversine(t *testing.T) {
	// Testing with known geographical coordinates
	result := haversine(51.5007, 0.1246, 40.6892, 74.0445) // London to New York
	// expected distance in kilometers
	expected := 5575.0

	// The distance should be accurate within a small margin of error
	assert.InEpsilon(t, expected, result, 0.01, "Haversine function should calculate correct distance")
}

func TestReadAndFilterData(t *testing.T) {
	// The path of Tests Files
	path := filepath.Join(".", "sample_data.csv")

	// Call the function with the file
	points, err := readAndFilterData(path)
	assert.NoError(t, err, "Failed to read and filter data from file")

	// Test for expected results
	expectedNumberOfPoints := 97 // The number expected to Read from csv file
	assert.Equal(t, expectedNumberOfPoints, len(points), "Number of parsed points does not match expected")

	// Further assertions can check for specific data integrity and values
	if len(points) > 0 {
		assert.Equal(t, "1", points[0].ID, "First point ID should be 1")
		assert.InDelta(t, 35.706552, points[0].Latitude, 0.0001, "Latitude of first point should match")
		assert.InDelta(t, 51.412262, points[0].Longitude, 0.0001, "Longitude of first point should match")
	}
}
