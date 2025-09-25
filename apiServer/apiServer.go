package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Response defines the JSON payload from mock users
type Response struct {
	UserID   int    `json:"user_id"`
	Response string `json:"response"`
	DelayMS  int    `json:"delay_ms"`
}

// Forwarder holds the channel for forwarding responses
type Forwarder struct {
	ForwardChan chan Response
	GameURL     string
	wg          sync.WaitGroup
}

// NewForwarder initializes the forwarder
func NewForwarder(gameURL string) *Forwarder {
	f := &Forwarder{
		ForwardChan: make(chan Response, 1000), // buffered channel
		GameURL:     gameURL,
	}
	f.wg.Add(1)
	for i := 0; i < 5; i++ {
    go f.startForwarding()
}

	return f
}

// startForwarding reads from the channel and sends to game engine
func (f *Forwarder) startForwarding() {
	defer f.wg.Done()
	for res := range f.ForwardChan {
		data, _ := json.Marshal(res)
		_, err := http.Post(f.GameURL, "application/json", bytes.NewBuffer(data))
		if err != nil {
			fmt.Printf("Error forwarding User %d to Game Engine: %v\n", res.UserID, err)
		} else {
			fmt.Printf("Forwarded User %d response to Game Engine\n", res.UserID)
		}
	}
}

// handleResponse accepts mock user responses
func (f *Forwarder) handleResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	var res Response
	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	f.ForwardChan <- res // blocks if channel is full

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"msg":    fmt.Sprintf("Received response from user %d", res.UserID),
	})
}

func main() {
	gameEngineURL := "http://localhost:9090/game/response"
	forwarder := NewForwarder(gameEngineURL)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "API Server Running. POST to /submit")
	})
	http.HandleFunc("/submit", forwarder.handleResponse)

	fmt.Println("API Server running on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting API server:", err)
	}
}
