package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// Room represents a game room
type Room struct {
	ID      string   `json:"id"`
	Players []string `json:"players"` // Array of user IDs
	Status  string   `json:"status"`  // e.g., "waiting", or "full"
}

// In-memory storage
var (
	rooms          = make(map[string]*Room)  //roomID -> Room
	waitingRoomID  *string                   // ID of room waiting for players
	mu             sync.RWMutex              //Mutex for thread-safe access
	userServiceURL = "http://localhost:8001" // User service endpoint
	gameServiceURL = "http://localhost:8003" // Game service endpoint
)


func main() {
	// Register routes
	http.HandleFunc("/join", joinRoomHandler)
	http.HandleFunc("/rooms/", getRoomHandler) // trailing slash for /rooms/{id}
	http.HandleFunc("/health", healthHandler)

	port := ":8002"
	fmt.Printf("Room Service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}