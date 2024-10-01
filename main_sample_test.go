package main

import (
	"os"
	"strconv"
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
	// Mock a small CSV file for testing
	fileContent := `id_delivery,lat,lng,timestamp
1,35.0,51.0,1609459200
1,35.1,51.1,1609459260
2,36.0,52.0,1609459320`

	// Create temporary CSV file
	f, err := os.CreateTemp("", "test_data.csv")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.Write([]byte(fileContent))
	assert.NoError(t, err)
	f.Close()

	// Test the readDataChunks function
	ch, err := readDataChunks(f.Name())
	assert.NoError(t, err)

	var i int = 1
	for points := range ch {
		assert.Greater(t, len(points), 0, "Expected to read at least one point")
		assert.Equal(t, strconv.Itoa(i), points[0].ID, "Expected the Chunk ID Number to be %d", i)
		i += 1
	}
}

func TestFilterInvalidPoints(t *testing.T) {
	points := []DeliveryPoint{
		{ID: "1", Latitude: 35.0, Longitude: 40.0, Timestamp: 1000},
		{ID: "1", Latitude: 35.1, Longitude: 40.1, Timestamp: 1300}, // Valid speed
		{ID: "1", Latitude: 35.0, Longitude: 40.0, Timestamp: 1400}, // Invalid speed
	}

	filteredPoints := filterInvalidPoints(points)
	assert.Len(t, filteredPoints, 1, "Expected only valid points to be returned")
}

func TestCalculateFare(t *testing.T) {
	points := []DeliveryPoint{
		{ID: "1", Latitude: 35.0, Longitude: 40.0, Timestamp: 1609459200},
		{ID: "1", Latitude: 35.1, Longitude: 40.1, Timestamp: 1609459260}, // Valid point
		{ID: "1", Latitude: 35.2, Longitude: 40.2, Timestamp: 1609459320}, // Valid point
	}

	// Expect fare > 3.47 (the minimum) since we have valid points
	fare := calculateFare(points)
	assert.GreaterOrEqual(t, fare, 3.47, "Fare should not be less than the minimum required fare")
	assert.Greater(t, fare, 1.30, "Fare should be greater than the base flag amount")
}

func TestOutputFareResults(t *testing.T) {
	// Create a temporary file
	file, err := os.CreateTemp("", "test_fares.csv")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Pass the file path to the output function
	outputFareResults(file.Name(), "1", 5.50)

	// Read back the file to ensure it wrote correctly
	data, err := os.ReadFile(file.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(data), "1,5.50", "The output file should contain the fare")
}
