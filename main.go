package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type DeliveryPoint struct {
	ID        string
	Latitude  float64
	Longitude float64
	Timestamp int64
}

const (
	radiusEarthKm = 6371.0 // radius of the Earth in kilometers
)

var mu sync.Mutex

func main() {
	// Capture the start time of the program
	programStartTime := time.Now()

	chunks, err := readDataChunks("input_dataset/large_data.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Use a map to store fares by delivery ID
	fares := make(map[int]string)

	var wg sync.WaitGroup
	for chunk := range chunks {
		wg.Add(1)

		go func(chunk []DeliveryPoint) {
			defer wg.Done()

			if len(chunk) == 0 {
				return
			}

			// Filter invalid points
			filteredChunk := filterInvalidPoints(chunk)
			if len(filteredChunk) == 0 {
				return
			}

			// Calculate fare
			fare := calculateFare(filteredChunk)

			// Convert deliveryID to int for sorting later
			deliveryID, err := strconv.Atoi(filteredChunk[0].ID)
			if err != nil {
				log.Fatal("Error converting delivery ID:", err)
			}

			// Print total time elapsed since the program started
			totalElapsed := time.Since(programStartTime)
			fmt.Printf("Calculating id_delivery = %d, total time elapsed: %v Please wait...\n", deliveryID, totalElapsed)

			// Lock to safely write to the fares map
			mu.Lock()
			fares[deliveryID] = fmt.Sprintf("%d,%.2f", deliveryID, fare)
			mu.Unlock()

		}(chunk)
	}

	wg.Wait()

	// Now write the fares map to a CSV file in sorted order
	outputFile, err := os.Create("output_dataset/fares.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"id_delivery", "fare_estimate"})

	// Sort the delivery IDs for ordered output
	var deliveryIDs []int
	for id := range fares {
		deliveryIDs = append(deliveryIDs, id)
	}
	sort.Ints(deliveryIDs)

	// Write the fares to the CSV file in the correct order
	for _, id := range deliveryIDs {
		writer.Write([]string{strconv.Itoa(id), fares[id][len(strconv.Itoa(id))+1:]}) // strip id from fare
	}

	fmt.Println("Fares have been written successfully in output_dataset/fares successfully :)")
}

// haversine calculates the distance between two latitude/longitude points.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	// Distance between latitude and longitudes
	deltaLat := (lat2 - lat1) * math.Pi / 180.0
	deltaLon := (lon2 - lon1) * math.Pi / 180.0

	// Convert to radians
	latRad1 := lat1 * math.Pi / 180.0
	latRad2 := lat2 * math.Pi / 180.0

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(latRad1)*math.Cos(latRad2)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return radiusEarthKm * c
}

// readDataChunks reads rows from a CSV file and returns delivery points in chunks.
func readDataChunks(filePath string) (chan []DeliveryPoint, error) {
	ch := make(chan []DeliveryPoint)
	go func() {
		defer close(ch)
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		var currentID string
		var points []DeliveryPoint

		for {
			line, err := reader.Read()
			if err == io.EOF {
				if len(points) > 0 {
					ch <- points // Send the last batch
				}
				break
			}
			if err != nil {
				log.Fatal(err)
				return
			}

			if currentID != "" && line[0] != currentID {
				ch <- points
				points = nil // Start a new batch
			}

			if line[0] != "id_delivery" {
				currentID = line[0]
				lat, _ := strconv.ParseFloat(line[1], 64)
				lng, _ := strconv.ParseFloat(line[2], 64)
				timestamp, _ := strconv.ParseInt(line[3], 10, 64)
				points = append(points, DeliveryPoint{
					ID:        line[0],
					Latitude:  lat,
					Longitude: lng,
					Timestamp: timestamp,
				})
			}
		}
	}()
	return ch, nil
}

// filterInvalidPoints filters out points based on speed calculations.
func filterInvalidPoints(points []DeliveryPoint) []DeliveryPoint {
	var validPoints []DeliveryPoint
	if len(points) == 0 {
		return validPoints
	}
	validPoints = append(validPoints, points[0]) // Add the first point by default

	for i := 1; i < len(points); i++ {
		p1 := points[i-1]
		p2 := points[i]
		distance := haversine(p1.Latitude, p1.Longitude, p2.Latitude, p2.Longitude)
		timeDiff := math.Abs(float64(p2.Timestamp - p1.Timestamp))
		speed := (distance / timeDiff) * 3600 // Speed in km/h

		if speed <= 100 {
			validPoints = append(validPoints, p2)
		}
	}

	return validPoints
}

// calculateFare calculates the fare based on delivery points.
func calculateFare(points []DeliveryPoint) float64 {
	if len(points) < 2 {
		return 0 // No fare if there's less than two points
	}

	// Base fare
	var totalFare float64 = 1.30

	for i := 1; i < len(points); i++ {
		// Calculate distance between consecutive points
		distance := haversine(points[i-1].Latitude, points[i-1].Longitude, points[i].Latitude, points[i].Longitude)
		timeDiff := float64(points[i].Timestamp-points[i-1].Timestamp) / 3600.0 // Time difference in hours
		speed := (distance / timeDiff)                                          // Speed in km/h

		// Determine fare based on time of day and speed
		hour := (points[i-1].Timestamp / 3600) % 24 // Hour of the day (0 to 23)

		if speed > 10 {
			// If moving
			if hour >= 5 && hour < 24 {
				totalFare += distance * 0.74
			} else {
				totalFare += distance * 1.30
			}
		} else {
			// If idle
			totalFare += timeDiff * 11.90 // Idle rate
		}
	}

	// Ensure the minimum delivery fare is 3.47
	if totalFare < 3.47 {
		totalFare = 3.47
	}

	return totalFare
}
