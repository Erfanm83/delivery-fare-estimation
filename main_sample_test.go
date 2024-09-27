package main

import (
	"bytes"
	"encoding/csv"
	"io"
	"strconv"
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
	// Mock CSV data
	// csvData is defined, which shows the contents of a CSV file.
	csvData := `id,lat,lng,timestamp
				1,34.0522,-118.2437,1597680651
				2,36.7783,-119.4179,1597680652`

	reader := csv.NewReader(bytes.NewBufferString(csvData))
	_, err := reader.Read() // Read header to skip

	// handle error
	if err != nil {
		return
	}

	var points []DeliveryPoint
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		lat, err := strconv.ParseFloat(line[1], 64)
		assert.NoError(t, err)
		lng, err := strconv.ParseFloat(line[2], 64)
		assert.NoError(t, err)
		timestamp, err := strconv.ParseInt(line[3], 10, 64)
		assert.NoError(t, err)

		points = append(points, DeliveryPoint{
			ID:        line[0],
			Latitude:  lat,
			Longitude: lng,
			Timestamp: timestamp,
		})
	}

	// Ensure data is read correctly
	assert.Len(t, points, 2, "Two points should be read from the CSV data")
	assert.Equal(t, "1", points[0].ID, "First point ID should match")
}
