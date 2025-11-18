package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Game represents an active game session
type Game struct {
	RoomID  string   `json:"room_id"`
	Players []string `json:"players"`
	Status  string   `json:"status"`
}

func main() {
	http.HandleFunc("/game/start", startGameHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8003"
	fmt.Printf("Game Rules Service running on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

type StartGameRequest struct {
	RoomID  string   `json:"room_id"`
	Players []string `json:"players"`
}

type StartGameResponse struct {
	RoomID  string `json:"room_id"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func startGameHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req StartGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.RoomID == "" || len(req.Players) != 2 {
		http.Error(w, "Invalid game start request: need room_id and 2 players", http.StatusBadRequest)
		return
	}

	// Stub: Log game start. TODO: Implement actual game logic
	log.Printf("Game starting for room %s", req.RoomID)
	log.Printf("Player1: %s", req.Players[0])
	log.Printf("Player2: %s", req.Players[1])
	log.Printf("TODO: Implement WebSocket logic for real-time gameplay")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StartGameResponse{
		RoomID:  req.RoomID,
		Message: "Game starting (stub mode)",
		Status:  "starting",
	})

	log.Printf("Game service acknowledged start for room %s", req.RoomID)
}
