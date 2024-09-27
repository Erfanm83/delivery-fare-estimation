package main

import (
	"encoding/csv"
	"fmt"
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

	fmt.Printf("Project Initialized \n")
	// haversine test
	fmt.Printf("haversine distance : %f\n", haversine(51.5007,
		0.1246,
		40.6892,
		74.0445))

	// read and filter data test
	points, err := readAndFilterData("sample_data.csv") // Now passes 10 to only test first 10 rows
	if err != nil {
		fmt.Printf("Error reading data: %v\n", err)
		return
	}
	for _, point := range points {
		fmt.Printf("ID: %s, Lat: %f, Lng: %f, Timestamp: %d\n", point.ID, point.Latitude, point.Longitude, point.Timestamp)
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
func readAndFilterData(filename string) ([]DeliveryPoint, error) {

	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	reader := csv.NewReader(file)
	rawCSVData, err := reader.ReadAll()

	if err != nil {
		return nil, err
	}

	var points []DeliveryPoint
	for i, line := range rawCSVData {
		if i == 0 { // Skip header
			continue
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
