package main

import (
	"encoding/csv"
	"math"
	"os"
	"strconv"
)

type DeliveryPoint struct {
	ID        string
	Latitude  float64
	Longitude float64
	Timestamp int64
}

const (
	radiusEarthKm = 6371.0 //radius of earth in kilometers
)

func main() {
	// start reading from the second row and look for delivery ID "2"
	points, err := readDataChunks("2", 26) // Adjust the row index based on whether data includes a header
	if err != nil {
		panic(err)
	}

	for _, point := range points {
		println("ID:", point.ID, "Lat:", point.Latitude, "Lng:", point.Longitude, "Timestamp:", point.Timestamp)
	}

}

// haversine func to calculate distance
func haversine(lat1, lon1, lat2, lon2 float64) float64 {

	// distance between latitude and longitudes
	deltaLat := (lat2 - lat1) * math.Pi / 180.0
	deltaLon := (lon2 - lon1) * math.Pi / 180.0

	// Convert to  radians
	latRad1 := lat1 * math.Pi / 180.0
	latRad2 := lat2 * math.Pi / 180.0

	// formul
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(latRad1)*math.Cos(latRad2)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return radiusEarthKm * c
}

// read and filter the data
// func readAndFilterData(filename string) ([]DeliveryPoint, error) {

// 	file, err := os.Open(filename)

// 	if err != nil {
// 		return nil, err
// 	}

// 	defer file.Close()

// 	reader := csv.NewReader(file)
// 	rawCSVData, err := reader.ReadAll()

// 	if err != nil {
// 		return nil, err
// 	}

// 	var points []DeliveryPoint
// 	for i, line := range rawCSVData {
// 		if i == 0 { // Skip header
// 			continue
// 		}
// 		lat, _ := strconv.ParseFloat(line[1], 64)
// 		lng, _ := strconv.ParseFloat(line[2], 64)
// 		timestamp, _ := strconv.ParseInt(line[3], 10, 64)
// 		points = append(points, DeliveryPoint{
// 			ID:        line[0],
// 			Latitude:  lat,
// 			Longitude: lng,
// 			Timestamp: timestamp,
// 		})
// 	}
// 	return points, nil

// }

// readDataChunks reads rows from a CSV file starting from 'startRow' for a specific delivery ID
// until a different delivery ID is encountered.
func readDataChunks(deliveryID string, startRow int) ([]DeliveryPoint, error) {
	file, err := os.Open("sample_data.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var points []DeliveryPoint

	// Move the reader to the startRow, skipping the initial rows
	for i := 0; i < startRow-1; i++ {
		if _, err = reader.Read(); err != nil {
			return nil, err // Handle EOF or other read errors
		}
	}

	// Start reading from startRow, and collect points until the ID changes
	for {
		line, err := reader.Read()
		if err != nil {
			break // Stop on EOF or other read errors
		}
		if line[0] != deliveryID {
			break // Stop if the delivery ID is different
		}

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

	return points, nil
}

// filterInvalidPoints filters out points based on speed calculations between consecutive points.
// func filterInvalidPoints(points []DeliveryPoint) []DeliveryPoint {
// 	var validPoints []DeliveryPoint
// 	if len(points) == 0 {
// 		return validPoints
// 	}
// 	validPoints = append(validPoints, points[0]) // Add the first point by default

// 	for i := 1; i < len(points); i++ {
// 		p1 := points[i-1]
// 		p2 := points[i]
// 		distance := haversine(p1.Latitude, p1.Longitude, p2.Latitude, p2.Longitude)
// 		timeDiff := math.Abs(float64(p2.Timestamp - p1.Timestamp))
// 		speed := (distance / timeDiff) * 3600 // speed in km/h

// 		if speed <= 100 {
// 			validPoints = append(validPoints, p2)
// 		}
// 	}

// 	return validPoints
// }
