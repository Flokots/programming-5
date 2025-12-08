package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/Flokots/programming-5/colorSync/shared/auth"
	"github.com/Flokots/programming-5/colorSync/shared/middleware"
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

// Service token for Zero Trust communication
var gameServiceToken string

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from React DEV server
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Generate service token for Game Service communication (Zero Trust)
	var err error
	gameServiceToken, err = auth.GenerateServiceToken("room-service")
	if err != nil {
		log.Fatalf("Failed to generate service token: %v", err)
	}
	log.Printf("Service token generated for Game Service communication")

	mux := http.NewServeMux()

	// Protect /join with JWT authentication
	mux.HandleFunc("/join", middleware.RequireAuth(joinRoomHandler))

	// Register routes
	mux.HandleFunc("/rooms/", getRoomHandler) // trailing slash for /rooms/{id}
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/room/", roomReadyHandler) // Note the trailing slash!

	port := ":8002"
	fmt.Printf("Room Service starting on port %s\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("POST /join         - Join matchmaking (requires JWT)\n")
	fmt.Printf("GET  /rooms/:id    - Get room info (public)\n")
	fmt.Printf("GET  /room/:id/ready - Check room status (public)\n")
	fmt.Printf("GET  /health       - Health check (public)\n")
	fmt.Printf("\n")

	handler := corsMiddleware(mux) // Wrap with CORS middleware
	log.Fatal(http.ListenAndServe(port, handler))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

type JoinRequest struct {
	UserID string `json:"user_id"`
}

type JoinResponse struct {
	RoomID  string   `json:"room_id"`
	Players []string `json:"players"`
	Status  string   `json:"status"`
	Message string   `json:"message"`
}

// Use JWT authentication from middleware
func joinRoomHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Get user claims from JWT token (validated by middleware)
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		http.Error(w, "Unauthorized - no user claims", http.StatusUnauthorized)
		return
	}

	// 3. Parse request body
	var req JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 4. Validate UserID
	if req.UserID == "" {
		log.Printf("ERROR: UserID is empty in join request")
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// 5. Verify user_id matches JWT token (security check)
	if req.UserID != claims.UserID {
		log.Printf("User %s (%s) attempted to join as %s",
			claims.Username, claims.UserID, req.UserID)
		http.Error(w, "User ID mismatch - cannot join as another user", http.StatusForbidden)
		return
	}

	log.Printf("User %s (%s) joining matchmaking", claims.Username, req.UserID)

	// 6. Verify user exists by calling User Service
	if !verifyUser(req.UserID) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 7. Find or create room (thread-safe)
	mu.Lock()
	defer mu.Unlock() // Unlock when function exits

	// Check if user is already in the waiting room
    if waitingRoomID != nil {
        waitingRoom := rooms[*waitingRoomID]
        for _, playerID := range waitingRoom.Players {
            if playerID == req.UserID {
                log.Printf("User %s already in waiting room %s", req.UserID, *waitingRoomID)
                http.Error(w, "You are already in matchmaking queue", http.StatusConflict)
                return
            }
        }
    }

	// Check if user is already in any room
    for _, room := range rooms {
        for _, playerID := range room.Players {
            if playerID == req.UserID {
                log.Printf("⚠️ User %s already in room %s", req.UserID, room.ID)
                http.Error(w, "You are already in an active room", http.StatusConflict)
                return
            }
        }
    }

	var room *Room

	// Check if there's a waiting room
	if waitingRoomID != nil {
		// Join existing room
		room = rooms[*waitingRoomID]

		// Double-check players are different
        if len(room.Players) > 0 && room.Players[0] == req.UserID {
            log.Printf("ERROR: Same user attempting to join twice: %s", req.UserID)
            http.Error(w, "Cannot match with yourself", http.StatusConflict)
            return
        }
		room.Players = append(room.Players, req.UserID)
		room.Status = "full"
		waitingRoomID = nil // No longer waiting
		log.Printf("User %s joined room %s (ROOM FULL - 2/2 players)", req.UserID, room.ID)

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

		log.Printf("User %s created room %s and is waiting for opponent(1/2 players)", req.UserID, room.ID)
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JoinResponse{
		RoomID:  room.ID,
		Players: room.Players,
		Status:  room.Status,
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

// notifyGameService notifies Game Service to start the game,
// sends service token for zero trust auth
func notifyGameService(roomID string, players []string) {
	url := fmt.Sprintf("%s/game/start", gameServiceURL)

	payload := map[string]interface{}{
		"room_id": roomID,
		"players": players,
	}

	jsonData, _ := json.Marshal(payload)

	// Create request with service token
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request to Game Service: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	//Add service token for Zero Trust
	req.Header.Set("X-Service-Token", gameServiceToken)

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error calling Game Service: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Game Service notified for room %s", roomID)
	} else {
		log.Printf("Game Service returned status %d", resp.StatusCode)
	}
}

type RoomResponse struct {
	ID      string   `json:"id"`
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

// Make sure this function exists:
func roomReadyHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract room ID from URL path
	// URL format: /room/{roomID}/ready
	path := r.URL.Path
	if !strings.HasPrefix(path, "/room/") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Remove "/room/" prefix
	remainder := strings.TrimPrefix(path, "/room/")

	// Split by "/" to get roomID and "ready"
	parts := strings.Split(remainder, "/")
	if len(parts) != 2 || parts[1] != "ready" {
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}

	roomID := parts[0]

	// Look up room
	mu.RLock()
	room, exists := rooms[roomID]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Check if room is ready (has 2 players)
	ready := len(room.Players) == 2

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ready":   ready,
		"players": room.Players,
	})

	log.Printf("Room %s ready status: %v", roomID, ready)
}
