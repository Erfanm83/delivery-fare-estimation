package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type DeliveryPoint struct {
	ID        string
	Latitude  float64
	Longitude float64
	Timestamp int
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

	programStartTime := time.Now()

	inputFile, err := os.Open("input_dataset/huge_data.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	bufferedReader := bufio.NewReader(bufio.NewReaderSize(inputFile, 1e9))

	outputFile, err := os.Create("output_dataset/fares.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()
	writer.WriteString("id_delivery,fare_estimate\n")

	headerLine, err := bufferedReader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	_ = headerLine

	for {
		line, err := bufferedReader.ReadString('\n')
		if err != nil {
			if len(currentID) > 0 {
				if totalFare < 3.47 {
					totalFare = 3.47
				}
				writer.WriteString(fmt.Sprintf("%s,%.2f\n", currentID, totalFare))
			}
			break
		}

		fields := strings.Split(strings.TrimSpace(line), ",")
		deliveryID := fields[0]
		lat, _ := strconv.ParseFloat(fields[1], 64)
		lng, _ := strconv.ParseFloat(fields[2], 64)
		timestamp, _ := strconv.Atoi(fields[3])

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
				writer.WriteString(fmt.Sprintf("%s,%.2f\n", currentID, totalFare))
			}

			// Start a new delivery calculation
			currentID = deliveryID
			tempData = currentPoint
			totalFare = 0.0
			continue
		}

		distance := haversine(tempData.Latitude, tempData.Longitude, currentPoint.Latitude, currentPoint.Longitude)
		timeDiff := math.Abs(float64(currentPoint.Timestamp - tempData.Timestamp))
		speed := (distance / timeDiff) * 3.60

		if speed <= 100 {
			hour := (tempData.Timestamp / 3600) % 24
			if speed > 10 {
				// Moving
				if hour >= 5 && hour < 24 {
					totalFare += distance * 0.74
				} else {
					totalFare += distance * 1.30
				}
			} else {
				// Idle
				totalFare += (timeDiff / 3600.0) * 11.90
			}
			tempData = currentPoint
		}
	}

	totalElapsed := time.Since(programStartTime)
	fmt.Printf("total time elapsed: %v\n", totalElapsed)
	fmt.Println("Fares have been written successfully in output_dataset/fares.csv successfully :)")
}

// calculates the distance between two latitude/longitude points.
func Fasthaversine(lat1, lon1, lat2, lon2 float64) float64 {
	lat1, lon1, lat2, lon2 = lat1*degToRad, lon1*degToRad, lat2*degToRad, lon2*degToRad

	// spherical law of cosines for faster approximation
	return radiusEarthKm * math.Acos(math.Sin(lat1)*math.Sin(lat2)+
		math.Cos(lat1)*math.Cos(lat2)*math.Cos(lon2-lon1))
}

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
