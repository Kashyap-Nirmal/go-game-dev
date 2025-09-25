package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Response structure
type Response struct {
	UserID   int    `json:"user_id"`
	Response string `json:"response"`
	DelayMS  int    `json:"delay_ms"`
}

// GameEngine with a channel
type GameEngine struct {
	mu       sync.Mutex
	winner   *Response
	ResponseChan chan Response
	wg       sync.WaitGroup
}

// NewGameEngine initializes the engine
func NewGameEngine() *GameEngine {
	ge := &GameEngine{
		ResponseChan: make(chan Response, 100),
	}
	ge.wg.Add(1)
	go ge.processResponses()
	return ge
}

// processResponses reads from channel and determines winner
func (g *GameEngine) processResponses() {
	defer g.wg.Done()
	for res := range g.ResponseChan {
		g.mu.Lock()
		if g.winner == nil && res.Response == "yes" {
			g.winner = &res
			fmt.Printf("🏆 Winner is User %d (Delay: %dms)\n", res.UserID, res.DelayMS)
		}
		g.mu.Unlock()
	}
}

// handleResponse accepts forwarded responses from API server
func (g *GameEngine) handleResponse(w http.ResponseWriter, r *http.Request) {
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

	// Send to processing channel
	select {
	case g.ResponseChan <- res:
	default:
		fmt.Println("Warning: GameEngine channel full, dropping response")
	}

	w.Header().Set("Content-Type", "application/json")
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.winner != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"winner":  g.winner.UserID,
			"message": "Game has a winner",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"winner":  nil,
			"message": "Waiting for correct answer...",
		})
	}
}

// homepage
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Game Engine</title>
		</head>
		<body style="font-family:sans-serif;text-align:center;margin-top:50px;">
			<h1>Game Engine Running</h1>
			<p>POST responses to <code>/game/response</code></p>
			<p>The first user who answers <b>yes</b> wins!</p>
		</body>
		</html>
	`)
}

func main() {
	engine := NewGameEngine()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/game/response", engine.handleResponse)

	fmt.Println("Game Engine running on http://localhost:9090")
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println("Error starting Game Engine:", err)
	}
}
