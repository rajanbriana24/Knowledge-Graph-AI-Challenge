package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	apiURL = "https://api.openai.com/v1/chat/completions" 
	apiKey = "YOUR_API_KEY"                               // Replace with your API key.
)

// RequestBody represents the payload for the LLM API request.
type RequestBody struct {
	Model    string    json:"model"
	Messages []Message json:"messages"
	Stream   bool      json:"stream"
}

// Message represents a single message in the conversation.
type Message struct {
	Role    string json:"role"
	Content string json:"content"
}

// ResponseChoice contains the choices in the API response.
type ResponseChoice struct {
	Message Message json:"message"
}

// ResponseBody represents the LLM API response.
type ResponseBody struct {
	Choices []ResponseChoice json:"choices"
}

func chat(userText) {

	// Instruction for the LLM to return structured JSON output.
	prompt := fmt.Sprintf(`Extract and structure the following information from the given text. Return the result as a JSON object with null for missing fields:
{
  "age": null,
  "family_size": null,
  "profession": null,
  "income_level": null,
  "preferences": null,
  "moving_timeline": null,
  "budget": null,
  "amenities_priority": null,
  "location_name": null,
  "city": null,
  "region": null,
  "walkability_score": null,
  "transit_score": null,
  "property_type": null,
  "price": null,
  "address": null
}
Input text: "%s"`, userText)

	// Create the request body.
	requestBody := RequestBody{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant that extracts structured JSON information from user text."},
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	// Serialize the request body to JSON.
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("Error encoding request body: %v\n", err)
		return
	}

	// Create the HTTP request.
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v\n", err)
		return
	}

	// Add headers.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the HTTP request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending HTTP request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read the response.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	// Parse the response JSON.
	var response ResponseBody
	if err := json.Unmarshal(respBody, &response); err != nil {
		fmt.Printf("Error decoding response JSON: %v\n", err)
		return
	}

	// Display the extracted JSON output.
	if len(response.Choices) > 0 {
		fmt.Println("Extracted JSON Response:")
		fmt.Println(response.Choices[0].Message.Content)
	} else {
		fmt.Println("No response from the model.")
	}
}