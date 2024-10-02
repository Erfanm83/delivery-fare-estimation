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
)

type DeliveryPoint struct {
	ID        string
	Latitude  float64
	Longitude float64
	Timestamp int64
}

// FareRecord struct to hold each fare entry
type FareRecord struct {
	ID   string
	Fare float64
}

const (
	radiusEarthKm = 6371.0 //radius of earth in kilometers
)

var mu sync.Mutex
var headerWritten = false // Global flag to ensure header is written only oncevar Header bool = false

func main() {
	chunks, err := readDataChunks("input_dataset/delivery_large_data.csv")
	if err != nil {
		log.Fatal(err)
	}

	outputFile, err := os.Create("output_dataset/filtered_data.csv") // Create a new CSV to store the filtered data
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"id_delivery", "lat", "lng", "timestamp"})

	var wg sync.WaitGroup
	for chunk := range chunks {
		wg.Add(1)

		go func(chunk []DeliveryPoint) {
			defer wg.Done() // Ensure the WaitGroup knows this goroutine is done
			if len(chunk) == 0 {
				return
			}

			deliveryID := chunk[0].ID // Assuming all points in the chunk have the same ID
			// Filter out invalid points
			filteredChunk := filterInvalidPoints(chunk)

			// Use a mutex to lock the file writing operation
			mu.Lock() // Lock before writing
			for _, point := range filteredChunk {
				err := writer.Write([]string{
					point.ID,
					fmt.Sprintf("%f", point.Latitude),
					fmt.Sprintf("%f", point.Longitude),
					fmt.Sprintf("%d", point.Timestamp),
				})
				if err != nil {
					log.Fatal("Error writing filtered point to CSV:", err)
				}
			}
			writer.Flush() // Ensure that data is flushed to the file
			mu.Unlock()    // Unlock after writing

			// Fare calculation and writing to fare result file
			fare := calculateFare(filteredChunk)

			outputFareResults("output_dataset/fares.csv", deliveryID, fare)
		}(chunk)
	}

	wg.Wait() // Wait for all goroutines to finish before program exit
	// Sort the fare file
	sortFaresFile("output_dataset/fares.csv")
	fmt.Printf("Program Finished Susseccfully\n\t The Fares saved at output_dataset/fares.csv")
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

	// Add the standard 'flag' amount at the start of each delivery
	var totalFare float64 = 1.30

	for i := 1; i < len(points); i++ {
		// Calculate distance between consecutive points
		distance := haversine(points[i-1].Latitude, points[i-1].Longitude, points[i].Latitude, points[i].Longitude)
		timeDiff := float64(points[i].Timestamp-points[i-1].Timestamp) / 3600.0 // Time difference in hours
		speed := (distance / timeDiff)                                          // Speed in km/h

		// Extract the hour from the timestamp (assuming timestamps are UNIX-based in seconds)
		hour := (points[i-1].Timestamp / 3600) % 24 // Hour of the day (0 to 23)

		if speed > 10 {
			// If the vehicle is moving
			if hour >= 5 && hour < 24 {
				// Daytime rate: from 5:00 AM to Midnight
				totalFare += distance * 0.74
			} else {
				// Nighttime rate: from Midnight to 5:00 AM
				totalFare += distance * 1.30
			}
		} else {
			// If the vehicle is idle (speed <= 10 km/h), charge based on time
			totalFare += timeDiff * 11.90 // Idle rate
		}
	}

	// Ensure the minimum delivery fare is 3.47
	if totalFare < 3.47 {
		totalFare = 3.47
	}

	return totalFare
}

func outputFareResults(filePath, deliveryID string, fare float64) {
	mu.Lock() // Lock the file writing operation

	// Open the file with append mode and create if not exists
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Check and write header if not already done
	if !headerWritten {
		header := []string{"id_delivery", "fare_estimate"}
		if err := writer.Write(header); err != nil {
			log.Fatal("Error writing header:", err)
		}
		headerWritten = true // Set the flag to true after writing the header
	}

	// Write the fare data
	record := []string{
		deliveryID,
		fmt.Sprintf("%.2f", fare),
	}
	if err := writer.Write(record); err != nil {
		log.Fatal("Error writing record:", err)
	}

	mu.Unlock() // Unlock after writing
}

// Function to sort the fare records by delivery ID numerically
func sortFaresFile(filePath string) {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read the CSV content
	reader := csv.NewReader(file)
	var records []FareRecord

	// Skip header
	_, err = reader.Read()
	if err != nil {
		log.Fatal("Error reading CSV header:", err)
	}

	// Read the CSV rows into a slice
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// Convert the delivery ID to an integer
		id, err := strconv.Atoi(line[0])
		if err != nil {
			log.Fatal("Error parsing delivery ID:", err)
		}

		// Parse the fare value as float
		fare, err := strconv.ParseFloat(line[1], 64)
		if err != nil {
			log.Fatal("Error parsing fare:", err)
		}

		// Append to the slice
		records = append(records, FareRecord{
			ID:   strconv.Itoa(id), // Store ID as string for writing it back
			Fare: fare,
		})
	}

	// Sort the records numerically by ID
	sort.Slice(records, func(i, j int) bool {
		id1, _ := strconv.Atoi(records[i].ID)
		id2, _ := strconv.Atoi(records[j].ID)
		return id1 < id2
	})

	// Open the file again for writing the sorted data
	outputFile, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write the header
	writer.Write([]string{"id_delivery", "fare_estimate"})

	// Write the sorted records back to the CSV
	for _, record := range records {
		writer.Write([]string{record.ID, fmt.Sprintf("%.2f", record.Fare)})
	}

	fmt.Println("Fares sorted by id_delivery numerically!")
}
