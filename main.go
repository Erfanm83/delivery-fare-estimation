package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
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

// var mu sync.Mutex
// var headerWritten = false // Global flag to ensure header is written only oncevar Header bool = false

func main() {
	chunks, err := readDataChunks("sample_data.csv")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for chunk := range chunks {

		fmt.Printf("chunk : %v\n", chunk)
		fmt.Printf("------------------------------------------------------------------------\n")

		wg.Add(1)
		go func(chunk []DeliveryPoint) {
			defer wg.Done()
			if len(chunk) == 0 {
				return
			}
			deliveryID := chunk[0].ID // Assuming all points in the chunk have the same ID
			filteredChunk := filterInvalidPoints(chunk)
			fare := calculateFare(filteredChunk)
			fmt.Printf("deliveryID : %v,fare : %v\n", deliveryID, fare)

			fmt.Printf("------------------------------------------------------------------------\n")
			// outputFareResults(deliveryID, fare)
		}(chunk)
	}
	wg.Wait()

	// start reading from the second row and look for delivery ID "2"
	// points, err := readDataChunks("3", 58)
	// if err != nil {
	// 	panic(err)
	// }

	// for _, p := range points {
	// 	fmt.Printf("ID: %s, Lat: %f, Lng: %f, Time: %d\n", p.ID, p.Latitude, p.Longitude, p.Timestamp)
	// }
	// fmt.Printf("we have %d of data before filtering\n", len(points))
	// fmt.Println("After filtering : ")

	// filteredPoints := filterInvalidPoints(points)
	// Output or use filteredPoints as needed
	// for _, p := range filteredPoints {
	// 	fmt.Printf("ID: %s, Lat: %f, Lng: %f, Time: %d\n", p.ID, p.Latitude, p.Longitude, p.Timestamp)
	// }
	// fmt.Printf("we have %d of data after filtering\n", len(filteredPoints))

	// finalFares := calculateFares(filteredPoints)

	// file, err := os.Create("output_fares.csv")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// writer := csv.NewWriter(file)
	// defer writer.Flush()

	// writer.Write([]string{"id_delivery", "fare_estimate"}) // Writing header
	// for id, fare := range finalFares {
	// 	writer.Write([]string{id, fmt.Sprintf("%.2f", fare)})
	// }
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

// readDataChunks reads rows from a CSV file starting from 'startRow' for a specific delivery ID
// until a different delivery ID is encountered.
// func readDataChunks(deliveryID string, startRow int) ([]DeliveryPoint, error) {
// 	file, err := os.Open("sample_data.csv")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	reader := csv.NewReader(file)
// 	var points []DeliveryPoint

// 	// Move the reader to the startRow, skipping the initial rows
// 	for i := 0; i < startRow-1; i++ {
// 		if _, err = reader.Read(); err != nil {
// 			return nil, err // Handle EOF or other read errors
// 		}
// 	}

// 	// Start reading from startRow, and collect points until the ID changes
// 	for {
// 		line, err := reader.Read()
// 		if err != nil {
// 			break // Stop on EOF or other read errors
// 		}
// 		if line[0] != deliveryID {
// 			break // Stop if the delivery ID is different
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

//		return points, nil
//	}
func readDataChunks(filePath string) (chan []DeliveryPoint, error) {
	ch := make(chan []DeliveryPoint)
	go func() {
		defer close(ch)
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err) // Handle the error as appropriate for your case
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
					ch <- points // send the last batch
				}
				break
			}
			if err != nil {
				log.Fatal(err) // Handle the error as appropriate for your case
				return
			}

			if currentID != "" && line[0] != currentID {
				ch <- points
				points = nil // start a new batch
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

// filterInvalidPoints filters out points based on speed calculations between consecutive points.
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
		speed := (distance / timeDiff) * 3600 // speed in km/h

		if speed <= 100 {
			validPoints = append(validPoints, p2)
		}
	}

	return validPoints
}

func calculateFare(points []DeliveryPoint) float64 {
	if len(points) < 2 {
		return 0 // No fare if there's less than two points
	}
	var totalFare float64
	for i := 1; i < len(points); i++ {
		distance := haversine(points[i-1].Latitude, points[i-1].Longitude, points[i].Latitude, points[i].Longitude)
		totalFare += distance // $1.00 per km
	}
	return totalFare
}

// func outputFareResults(deliveryID string, fare float64) {
// 	filePath := "fares.csv"

// 	mu.Lock() // Ensure that no other goroutine can enter this section while one is working

// 	// Open the file with append mode and create if not exists
// 	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer file.Close()

// 	writer := csv.NewWriter(file)
// 	defer writer.Flush()

// 	// Check and write header if not already done
// 	if !headerWritten {
// 		header := []string{"id_delivery", "fare"}
// 		if err := writer.Write(header); err != nil {
// 			log.Fatal("Error writing header:", err)
// 		}
// 		headerWritten = true // Set the flag to true after writing the header
// 	}

// 	mu.Unlock() // Release the mutex lock after the header check

// 	// Write the fare data
// 	record := []string{
// 		deliveryID,
// 		fmt.Sprintf("%.2f", fare),
// 	}
// 	if err := writer.Write(record); err != nil {
// 		log.Fatal("Error writing record:", err)
// 	}
// }
