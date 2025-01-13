package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbUri := os.Getenv("NEO4J_URI")
	dbUser := os.Getenv("NEO4J_USERNAME")
	dbPassword := os.Getenv("NEO4J_PASSWORD")

	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.BasicAuth(dbUser, dbPassword, ""))
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}
	defer driver.Close(ctx)

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		log.Fatalf("Failed to verify connectivity to Neo4j: %v", err)
	}
	fmt.Println("Connected to Neo4j successfully!")

	// Get user input
	fmt.Println("\nHow can I help you? (Type 'Done' to exit):")
	var userInput string
	fmt.Scanln(&userInput)

	if userInput == "Done" {
		fmt.Println("Thank you! Goodbye!")
		return
	}

	// Step 1: Extract structured information using OpenAI's API
	chatResponse, err := chat(userInput)
	if err != nil {
		log.Fatalf("Error in chat function: %v", err)
	}

	// Parse the JSON response from OpenAI
	var extractedData map[string]interface{}
	err = json.Unmarshal([]byte(chatResponse), &extractedData)
	if err != nil {
		log.Fatalf("Error parsing extracted data: %v", err)
	}

	fmt.Printf("Extracted data: %+v\n", extractedData)

	// Step 2: Query the database using the extracted data
	graphData, err := queryGraph(ctx, driver, extractedData)
	if err != nil {
		log.Fatalf("Error querying graph: %v", err)
	}

	fmt.Println("\nGraph Query Results (Structured Data):")
	fmt.Println(graphData)

	// Step 3: Send extracted data + graph results to OpenAI to generate a response
	finalPrompt := fmt.Sprintf(`Based on the following structured data and query results, generate a user-friendly response:
Structured Data: %s
Query Results: %s`, chatResponse, graphData)

	finalResponse, err := chat(finalPrompt) // Reuse the `chat` function to send the final request to OpenAI
	if err != nil {
		log.Fatalf("Error generating final user-friendly response: %v", err)
	}

	// Step 4: Print the final response from OpenAI
	fmt.Println("\nFinal Response:")
	fmt.Println(finalResponse)
}
