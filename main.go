package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
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
	c             = math.Pi / 360.0
	degToRad      = math.Pi / 180.0
)

func main() {
	var currentID string
	var tempData DeliveryPoint
	var totalFare float64

	// Start time of the program
	programStartTime := time.Now()

	// Open input file
	inputFile, err := os.Open("input_dataset/huge_data.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	// Create output file
	outputFile, err := os.Create("output_dataset/fares.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()
	writer.Write([]string{"id_delivery", "fare_estimate"})

	reader := csv.NewReader(inputFile)
	_, err = reader.Read() // Skip header
	if err != nil {
		log.Fatal(err)
	}

	for {
		line, err := reader.Read()
		// unexpected error
		if err != nil {
			if len(currentID) > 0 {
				// Write a very basic fare
				if totalFare < 3.47 {
					totalFare = 3.47
				}
				writer.Write([]string{currentID, fmt.Sprintf("%.2f", totalFare)})
			}
			break
		}

		deliveryID := line[0]
		lat, _ := strconv.ParseFloat(line[1], 64)
		lng, _ := strconv.ParseFloat(line[2], 64)
		timestamp, _ := strconv.ParseInt(line[3], 10, 64)

		currentPoint := DeliveryPoint{
			ID:        deliveryID,
			Latitude:  lat,
			Longitude: lng,
			Timestamp: timestamp,
		}

		if currentID != deliveryID {
			if len(currentID) > 0 {
				// Write the fare for the previous delivery ID
				if totalFare < 3.47 {
					totalFare = 3.47
				}
				writer.Write([]string{currentID, fmt.Sprintf("%.2f", totalFare)})
			}

			// Start a new delivery calculation
			currentID = deliveryID
			tempData = currentPoint
			totalFare = 0.0
			continue
		}

		distance := fastHaversine(tempData.Latitude, tempData.Longitude, currentPoint.Latitude, currentPoint.Longitude)
		timeDiff := math.Abs(float64(currentPoint.Timestamp - tempData.Timestamp))
		speed := distance / timeDiff * 3600.0

		// Calculate fare based on speed and time of day
		if speed <= 100 { // Valid speed threshold
			hour := (tempData.Timestamp / 3600) % 24
			if speed > 10 {
				// Moving
				if hour >= 5 && hour < 24 {
					totalFare += distance * 0.74 // Daytime rate
				} else {
					totalFare += distance * 1.30 // Midnight rate
				}
			} else {
				// Idle
				totalFare += (timeDiff / 3600.0) * 11.90 // Idle rate per hour
			}
			tempData = currentPoint
		}
	}

	totalElapsed := time.Since(programStartTime)
	fmt.Printf("total time elapsed: %v\n", totalElapsed)
	fmt.Println("Fares have been written successfully in output_dataset/fares.csv successfully :)")
}

// calculates the distance between two latitude/longitude points.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	// Distance between latitude and longitudes
	deltaLat := (lat2 - lat1) * c
	deltaLon := (lon2 - lon1) * c

	// Convert to radians
	latRad1 := lat1 * c
	latRad2 := lat2 * c

	// Haversine formula
	a := math.Sin(deltaLat)*math.Sin(deltaLat) +
		math.Cos(latRad1)*math.Cos(latRad2)*math.Sin(deltaLon)*math.Sin(deltaLon)
	d := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return radiusEarthKm * d
}

func fastHaversine(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	lat1, lon1, lat2, lon2 = lat1*degToRad, lon1*degToRad, lat2*degToRad, lon2*degToRad

	// Using the spherical law of cosines for faster approximation
	return radiusEarthKm * math.Acos(math.Sin(lat1)*math.Sin(lat2)+
		math.Cos(lat1)*math.Cos(lat2)*math.Cos(lon2-lon1))
}
