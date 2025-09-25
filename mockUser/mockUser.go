package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// Response defines the JSON structure to send
type Response struct {
	UserID   int    `json:"user_id"`
	Response string `json:"response"`
	DelayMS  int    `json:"delay_ms"`
}

// simulateUser simulates a user sending JSON yes/no to an API endpoint.
func simulateUser(id int, endpoint string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Random delay between 10–1000ms
	delay := rand.Intn(991) + 10
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// Random yes/no response
	resp := "no"
	if rand.Intn(2) == 0 {
		resp = "yes"
	}

	// Prepare JSON payload
	payload := Response{
		UserID:   id,
		Response: resp,
		DelayMS:  delay,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("User %d: failed to marshal JSON: %v\n", id, err)
		return
	}

	// Send POST request
	respObj, err := http.Post(endpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("User %d: error sending request: %v\n", id, err)
		return
	}
	defer respObj.Body.Close()

	fmt.Printf("User %d sent %s after %dms → Status: %s\n",
		id, resp, delay, respObj.Status)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var n int
	fmt.Print("Enter number of users: ")
	_, err := fmt.Scanln(&n)
	if err != nil || n <= 0 {
		fmt.Println("Invalid input. Please enter a positive integer.")
		return
	}

	endpoint := "http://localhost:8080/submit"

	var wg sync.WaitGroup
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go simulateUser(i, endpoint, &wg)
	}
	wg.Wait()
}

// NOTE: With the assumption that Yes is always the correct answer.