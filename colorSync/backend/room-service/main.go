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

type JoinRequest struct {
	UserID string `json:"user_id"`
}

type JoinResponse struct {
	RoomID  string `json:"room_id"`
	Players []string `json:"players"`
	Status string `json:"status"`
	Message string `json:"message"`
}

func joinRoomHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Parse request body
	var req JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Validate UserID 
	if req.UserID == "" {
		http.Error(w, "UserID required", http.StatusBadRequest)
		return
	}

	// 4. Verify user exists by calling User Service
	if !verifyUser(req.UserID) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Find or create room (thread-safe)
	mu.Lock()
	defer mu.Unlock() // Unlock when function exits

	var room *Room

	// Check if there's a waiting room 
	if waitingRoomID != nil {
		// Join existing room
		room = rooms[*waitingRoomID]
		room.Players = append(room.Players, req.UserID)
		room.Status = "full"
		waitingRoomID = nil // No longer waiting
		log.Printf("User %s joined room %s (now full)", req.UserID, room.ID)

		// Notify Game Service to start the game
		go notifyGameService(room.ID, room.Players) // Run in background

	} else {
		// Create new room 
		room = &Room{
			ID:      uuid.New().String(),
			Players: []string{req.UserID},
			Status:  "waiting",
		}
		rooms[room.ID] = room
		waitingRoomID = &room.ID

		log.Printf("User %s created and joined room %s (waiting for players)", req.UserID, room.ID)
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JoinResponse{
		RoomID:  room.ID,
		Players: room.Players,
		Status: room.Status,
		Message: fmt.Sprintf("Joined room %s", room.ID),
	})
}

// verifyUser calls User Service to check if user exists
func verifyUser(userID string) bool {
	url := fmt.Sprintf("%s/users/%s", userServiceURL, userID)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error calling User Service: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("User %s not found in User Service", userID)
		return false
	}
	log.Printf("Verified user %s exists", userID)
	return true
}

// notifyGameService notifies Game Service to start the game
func notifyGameService(roomID string, players []string) {
	url := fmt.Sprintf("%s/game/start", gameServiceURL)

	payload := map[string]interface{}{
		"room_id": roomID,
		"players": players,
	}

	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error calling Game Service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Game Service notified for room %s", roomID)
	} else {
		log.Printf("Game Service returned status %d", resp.StatusCode)
	}
}

type RoomResponse struct {
	ID 	string   `json:"id"`
	Players []string `json:"players"`
	Status  string   `json:"status"`
}

func getRoomHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Extract roomID from URL path
	path := r.URL.Path
	const roomsPrefix = "/rooms/"
	if len(path) <= len(roomsPrefix) {
		http.Error(w, "Room ID required", http.StatusBadRequest)
		return
	}
	roomID := path[len(roomsPrefix):]

	// 3. Look up room (thread-safe read)
	mu.RLock()
	room, exists := rooms[roomID]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// 4. Return room info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RoomResponse{
		ID:      room.ID,
		Players: room.Players,
		Status:  room.Status,
	})
}