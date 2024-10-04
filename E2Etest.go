package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEndToEndFlow(t *testing.T) {
	// 1. Mock an input CSV file (simulate raw delivery data)
	inputData := `id_delivery,lat,lng,timestamp
1,35.0,51.0,1609459200
1,35.1,51.1,1609459260
2,36.0,52.0,1609459320
2,36.1,52.1,1609459380`

	// Create a temporary input file
	inputFile, err := os.CreateTemp("", "test_input.csv")
	assert.NoError(t, err)
	defer os.Remove(inputFile.Name())

	_, err = inputFile.Write([]byte(inputData))
	assert.NoError(t, err)
	inputFile.Close()

	// 2. Create temporary output files for filtered data and fares
	filteredFile, err := os.CreateTemp("", "test_filtered_data.csv")
	assert.NoError(t, err)
	defer os.Remove(filteredFile.Name())

	faresFile, err := os.CreateTemp("", "test_fares.csv")
	assert.NoError(t, err)
	defer os.Remove(faresFile.Name())

	// 3. Run the full processing flow (similar to main function)
	chunks, err := readDataChunks(inputFile.Name())
	assert.NoError(t, err)

	outputFile, err := os.Create(filteredFile.Name())
	assert.NoError(t, err)
	defer outputFile.Close()
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write header for the filtered data
	writer.Write([]string{"id_delivery", "lat", "lng", "timestamp"})

	// Store fares in a map as in main.go
	fares := make(map[int]string)
	startTime := time.Now() // Capture start time for the elapsed time calculation

	for chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}

		// Filter invalid points
		filteredChunk := filterInvalidPoints(chunk)

		for _, point := range filteredChunk {
			err := writer.Write([]string{
				point.ID,
				fmt.Sprintf("%f", point.Latitude),
				fmt.Sprintf("%f", point.Longitude),
				fmt.Sprintf("%d", point.Timestamp),
			})
			assert.NoError(t, err)
		}
		writer.Flush()

		// Calculate fare and write to result file
		fare := calculateFare(filteredChunk)
		deliveryID := filteredChunk[0].ID
		elapsed := time.Since(startTime)
		fmt.Printf("Calculating delivery %s, total time elapsed: %v Please wait...\n", deliveryID, elapsed)

		id, _ := strconv.Atoi(deliveryID)
		fares[id] = fmt.Sprintf("%s,%.2f", deliveryID, fare)
	}

	// 4. Write the fare results to the fares file
	outputFareResultsFromMap(fares, faresFile.Name())

	// 5. Verify the results
	// Checks that the fare-estimatea were correctly written
	fareData, err := os.ReadFile(faresFile.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(fareData), "1,") // Check fare for id_delivery 1
	assert.Contains(t, string(fareData), "2,") // Check fare for id_delivery 2

	// Optionally, check the content of filtered data
	filteredData, err := os.ReadFile(filteredFile.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(filteredData), "1,35.0,51.0")
}

// Helper function to write fare results from a map (used in main)
func outputFareResultsFromMap(fares map[int]string, filePath string) {
	outputFile, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	writer.Write([]string{"id_delivery", "fare_estimate"})

	var deliveryIDs []int
	for id := range fares {
		deliveryIDs = append(deliveryIDs, id)
	}
	sort.Ints(deliveryIDs)

	for _, id := range deliveryIDs {
		writer.Write([]string{strconv.Itoa(id), fares[id][len(strconv.Itoa(id))+1:]}) // strip id from fare
	}
}
