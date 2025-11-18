package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Game represents an active game session
type Game struct {
	RoomID       string                     `json:"room_id"`
	Players      []string                   `json:"players"`
	Connections  map[string]*websocket.Conn // player ID to WebSocket connection
	Status       string                     `json:"status"`
	CurrentRound int                        `json:"current_round"`
	MaxRounds    int                        `json:"max_rounds"`
	Results      []RoundResult              `json:"results"`
	mu           sync.Mutex
}

type RoundResult struct {
	Round 	 int               `json:"round"`
	Word     string             `json:"word"`
	Color   string             `json:"color"`
	Winner   string             `json:"winner"`
	Latency int64			  `json:"latency_ms"`
}

// WebSocket message types
type WSMessage struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

var (
	games	  = make(map[string]*Game) // roomID to Game
	gamesMutex sync.RWMutex
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for simplicity; adjust in production
		},
	}
)

// Stroop colors and words
var colors = []string{"red", "blue", "green", "yellow"}
var words  = []string{"RED", "BLUE", "GREEN", "YELLOW"}

func main() {
	http.HandleFunc("/game/start", startGameHandler)
	http.HandleFunc("/game/ws", wsHandler)
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

	// Create game session
	game := &Game{
		RoomID:      req.RoomID,
		Players:     req.Players,
		Connections: make(map[string]*websocket.Conn),
		Status:      "waiting_for_players",
		CurrentRound: 0,
		MaxRounds:   5,
		Results:     []RoundResult{},
	}

	gamesMu.Lock()
	games[req.RoomID] = game
	gamesMu.Unlock()


	log.Printf("Game created for room %s (waiting for WebSocket connections)", req.RoomID)
	log.Printf("Players: %s vs %s", req.Players[0], req.Players[1])

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StartGameResponse{
		RoomID:  req.RoomID,
		Message: "Game created, waiting for players to connect via WebSocket",
		Status:  "waiting_for_players",
	})
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Get room_id and user_id from query params
	roomID := r.URL.Query().Get("room_id")
	userID := r.URL.Query().Get("user_id")

	if roomID == "" || userID == "" {
		http.Error(w, "Missing room_id or user_id", http.StatusBadRequest)
		return
	}

	// Find game
	gamesMu.RLock()
	game, exists := games[roomID]
	gamesMu.RUnlock()

	if !exists {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}

	// Register player connection
	game.mu.Lock()
	game.Connections[userID] = conn
	connCount := len(game.Connections)
	game.mu.Unlock()
	
	log.Printf("Player %s connected to room %s via WebSocket (%d/2)", userID, roomID, connCount)

	// If both players are connected, start the game
	if connCount == 2 {
		log.Printf("Both players connected! Starting game for room %s", roomID)
		go runGame(game)
	}

	// Listen for messages from this player
	go handlePlayerMessages(game, userID, conn)
}

func handlePlayerMessages(game *Game, userID string, conn *websocket.Conn) {
	defer conn.Close()

	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Player %s disconnected: %v", userID, err)
			return
		}

		log.Printf("Received message from player %s: %s", userID, msg.Type)

		// Handle different message types
		switch msg.Type {
		case "CLICK":
			handleClick(game, userID, msg.Payload)
		case "PING":
			// Heartbeat message
			conn.WriteJSON(WSMessage{Type: "PONG", Payload: map[string]interface{}{}})
		default:
			log.Printf("Unknown message type from player %s: %s", userID, msg.Type)
		}
	}
}