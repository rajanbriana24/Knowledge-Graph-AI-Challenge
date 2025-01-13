package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Zone represents a geographical zone with its attributes and utilities.
type Zone struct {
	Name                  string
	FamilySize            int
	MaritalStatus         string
	NumChildren           int
	AgeGroup              string
	NearbyParks           int
	NearbySchools         int
	NearbyHospitals       int
	LandType              string
	Landscape             string
	PublicTransportAccess bool
	Utilities             []string
	Buildings             []string
	ShoppingCenters       int
	FitnessCenters        int
	ChildCareServices     int
	AvgHousingCost        int
	CrimeRate             int
	RentalAvailability    int
	AvgSizePerHome        int
	AirQualityIndex       int
	GreenCover            int
	NoisePollutionLevel   int
}

// initializeDatabase reads the CSV file and populates the Neo4j database.
func initializeDatabase(ctx context.Context, driver neo4j.DriverWithContext, csvFilePath string) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// Read zones from CSV
	zones, err := readZonesFromCSV(csvFilePath)
	if err != nil {
		return fmt.Errorf("failed to read zones from CSV: %v", err)
	}

	// Create zones and relationships
	for _, zone := range zones {
		if err := createZone(ctx, session, zone); err != nil {
			return fmt.Errorf("failed to create zone %s: %v", zone.Name, err)
		}
	}

	// Define relationships between zones (including new zones: Gilbert, Chandler, and Sun City)
	zoneNeighbors := map[string][]string{
		"Downtown Phoenix": {"Tempe", "Scottsdale", "Chandler"},
		"Tempe":            {"Downtown Phoenix", "Mesa", "Gilbert"},
		"Scottsdale":       {"Downtown Phoenix", "Mesa"},
		"Mesa":             {"Tempe", "Scottsdale", "Gilbert"},
		"Gilbert":          {"Tempe", "Mesa", "Chandler"},
		"Chandler":         {"Gilbert", "Downtown Phoenix"},
		"Sun City":         {"Scottsdale", "Mesa"}, // Neighboring senior-friendly zones
	}

	// Create neighbor relationships
	for zone, neighbors := range zoneNeighbors {
		for _, neighbor := range neighbors {
			if err := createNeighborRelationship(ctx, session, zone, neighbor); err != nil {
				return fmt.Errorf("failed to create relationship between %s and %s: %v", zone, neighbor, err)
			}
		}
	}

	fmt.Println("Database successfully initialized with data from CSV and relationships.")
	return nil
}

func createNeighborRelationship(ctx context.Context, session neo4j.SessionWithContext, zone1, zone2 string) error {
	fmt.Printf("Creating NEIGHBORS relationship: %s <-> %s\n", zone1, zone2)

	query := `
		MATCH (z1:Zone {name: $zone1}), (z2:Zone {name: $zone2})
		MERGE (z1)-[:NEIGHBORS]->(z2)
		MERGE (z2)-[:NEIGHBORS]->(z1)
	`
	params := map[string]any{
		"zone1": zone1,
		"zone2": zone2,
	}

	return executeQuery(ctx, session, query, params)
}

// func executeQuery(ctx context.Context, session neo4j.SessionWithContext, query string, params map[string]any) error {
// 	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
// 		_, err := tx.Run(ctx, query, params)
// 		return nil, err
// 	})
// 	return err
// }


// readZonesFromCSV reads data from the CSV file and returns a slice of Zone structs.
func readZonesFromCSV(filePath string) ([]Zone, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open CSV file: %v", err)
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("failed to read CSV file: %v", err)
    }

    header := records[0]
    data := records[1:]

    var zones []Zone
    for _, row := range data {
        zone, err := parseZoneFromRow(header, row)
        if err != nil {
            return nil, fmt.Errorf("failed to parse zone: %v", err)
        }

        // Generate mock buildings
        for i := 0; i < 3; i++ {
            zone.Buildings = append(zone.Buildings, fmt.Sprintf("%s Building %c", zone.Name, 'A'+i))
        }

        fmt.Printf("Parsed Zone: %s with Buildings: %v and Utilities: %v\n", zone.Name, zone.Buildings, zone.Utilities)

        zones = append(zones, zone)
    }
    return zones, nil
}


// parseZoneFromRow parses a single CSV row into a Zone struct.
func parseZoneFromRow(header, row []string) (Zone, error) {
	zone := Zone{}
	var err error

	zone.Name = row[indexOf(header, "Zone")]
	if zone.FamilySize, err = strconv.Atoi(row[indexOf(header, "FamilySize")]); err != nil {
		return zone, fmt.Errorf("error parsing FamilySize: %v", err)
	}
	zone.MaritalStatus = row[indexOf(header, "MaritalStatus")]
	if zone.NumChildren, err = strconv.Atoi(row[indexOf(header, "NumChildren")]); err != nil {
		return zone, fmt.Errorf("error parsing NumChildren: %v", err)
	}
	zone.AgeGroup = row[indexOf(header, "AgeGroup")]
	if zone.NearbyParks, err = strconv.Atoi(row[indexOf(header, "NearbyParks")]); err != nil {
		return zone, fmt.Errorf("error parsing NearbyParks: %v", err)
	}
	if zone.NearbySchools, err = strconv.Atoi(row[indexOf(header, "NearbySchools")]); err != nil {
		return zone, fmt.Errorf("error parsing NearbySchools: %v", err)
	}
	if zone.NearbyHospitals, err = strconv.Atoi(row[indexOf(header, "NearbyHospitals")]); err != nil {
		return zone, fmt.Errorf("error parsing NearbyHospitals: %v", err)
	}
	zone.LandType = row[indexOf(header, "LandType")]
	zone.Landscape = row[indexOf(header, "Landscape")]
	zone.PublicTransportAccess = row[indexOf(header, "PublicTransportAccess")] == "Yes"
	zone.Utilities = strings.Split(row[indexOf(header, "Utilities")], ",")
	if zone.ShoppingCenters, err = strconv.Atoi(row[indexOf(header, "ShoppingCenters")]); err != nil {
		return zone, fmt.Errorf("error parsing ShoppingCenters: %v", err)
	}

	return zone, nil
}

// createZone creates a zone node and its relationships with utilities in the database.
func createZone(ctx context.Context, session neo4j.SessionWithContext, zone Zone) error {
    fmt.Printf("Creating Zone: %s with Properties: %+v\n", zone.Name, zone)

    // Create Zone node
    query := `
        MERGE (z:Zone {name: $name})
        SET z += $properties
    `
    properties := map[string]any{
        "FamilySize":            zone.FamilySize,
        "MaritalStatus":         zone.MaritalStatus,
        "NumChildren":           zone.NumChildren,
        "AgeGroup":              zone.AgeGroup,
        "NearbyParks":           zone.NearbyParks,
        "NearbySchools":         zone.NearbySchools,
        "NearbyHospitals":       zone.NearbyHospitals,
        "LandType":              zone.LandType,
        "Landscape":             zone.Landscape,
        "PublicTransportAccess": zone.PublicTransportAccess,
        "ShoppingCenters":       zone.ShoppingCenters,
        "FitnessCenters":        zone.FitnessCenters,
        "ChildCareServices":     zone.ChildCareServices,
        "AvgHousingCost":        zone.AvgHousingCost,
        "CrimeRate":             zone.CrimeRate,
        "RentalAvailability":    zone.RentalAvailability,
        "AvgSizePerHome":        zone.AvgSizePerHome,
        "AirQualityIndex":       zone.AirQualityIndex,
        "GreenCover":            zone.GreenCover,
        "NoisePollutionLevel":   zone.NoisePollutionLevel,
    }

    if err := executeQuery(ctx, session, query, map[string]any{"name": zone.Name, "properties": properties}); err != nil {
        return fmt.Errorf("error creating zone %s: %v", zone.Name, err)
    }

    // Create relationships for buildings
    for _, building := range zone.Buildings {
        fmt.Printf("Creating Building: %s in Zone: %s\n", building, zone.Name)
        query := `
            MERGE (b:Building {name: $building})
            MERGE (z:Zone {name: $zone})
            MERGE (b)-[:WITHIN_ZONE]->(z)
        `
        if err := executeQuery(ctx, session, query, map[string]any{"building": building, "zone": zone.Name}); err != nil {
            return fmt.Errorf("error creating building %s for zone %s: %v", building, zone.Name, err)
        }
    }

    // Create relationships for utilities
    for _, utility := range zone.Utilities {
        fmt.Printf("Creating Utility: %s SERVED_BY Zone: %s\n", utility, zone.Name)
        query := `
            MERGE (u:Utility {name: $utility})
            MERGE (z:Zone {name: $zone})
            MERGE (z)-[:SERVED_BY]->(u)
        `
        if err := executeQuery(ctx, session, query, map[string]any{"utility": utility, "zone": zone.Name}); err != nil {
            return fmt.Errorf("error creating utility %s for zone %s: %v", utility, zone.Name, err)
        }
    }

    return nil
}


// indexOf finds the index of a column name in the header.
func indexOf(header []string, columnName string) int {
	for i, name := range header {
		if name == columnName {
			return i
		}
	}
	return -1
}

// executeQuery executes a single query against the database.
func executeQuery(ctx context.Context, session neo4j.SessionWithContext, query string, params map[string]any) error {
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})
	return err
}
