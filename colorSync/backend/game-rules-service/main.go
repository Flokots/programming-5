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
	Connections  map[string]*websocket.Conn `json:"-"` // Don't serialize connections
	Status       string                     `json:"status"`
	CurrentRound int                        `json:"current_round"`
	MaxRounds    int                        `json:"max_rounds"`
	Results      []RoundResult              `json:"results"`

	disconnected map[string]bool `json:"-"` // Track disconnected players playerID -> disconnected

	// Round state (for click handling)
	currentWord    string
	currentColor   string
	roundStartTime time.Time
	roundAnswered  bool
	roundFinished  bool
	roundWinner    string
	roundLatency   int64
	wrongAnswers   map[string]bool // Track who got it wrong

	mu sync.Mutex
}

type RoundResult struct {
	Round   int    `json:"round"`
	Word    string `json:"word"`
	Color   string `json:"color"`
	Winner  string `json:"winner"`
	Latency int64  `json:"latency_ms"`
}

// WebSocket message types
type WSMessage struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

var (
	games    = make(map[string]*Game) // roomID to Game
	gamesMu  sync.RWMutex
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for simplicity; adjust in production
		},
	}
)

// Stroop colors and words
var colors = []string{"red", "blue", "green", "yellow"}
var words = []string{"RED", "BLUE", "GREEN", "YELLOW"}

func main() {
	http.HandleFunc("/game/start", startGameHandler)
	http.HandleFunc("/game/ws", wsHandler)
	http.HandleFunc("/game/status", gameStatusHandler) // ← ADD THIS
	http.HandleFunc("/health", healthHandler)

	port := ":8003"
	fmt.Printf("Game Rules Service running on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// NEW: Check if game exists
func gameStatusHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		http.Error(w, "room_id required", http.StatusBadRequest)
		return
	}

	gamesMu.RLock()
	game, exists := games[roomID]
	gamesMu.RUnlock()

	if !exists {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"room_id": game.RoomID,
		"status":  game.Status,
	})
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
		http.Error(w, "Invalid game start request", http.StatusBadRequest)
		return
	}

	// Create game session
	game := &Game{
		RoomID:       req.RoomID,
		Players:      req.Players,
		Connections:  make(map[string]*websocket.Conn),
		disconnected: make(map[string]bool),
		Status:       "waiting_for_players",
		MaxRounds:    5,
		Results:      []RoundResult{},
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
		Message: "Game created",
		Status:  "waiting_for_players",
	})
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Get room_id and user_id from query params
	roomID := r.URL.Query().Get("room_id")
	userID := r.URL.Query().Get("user_id")

	if roomID == "" || userID == "" {
		http.Error(w, "room_id and user_id required", http.StatusBadRequest)
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
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Register player connection
	game.mu.Lock()

	// Check if game already started or finished
	if game.Status == "in_progress" || game.Status == "finished" {
		log.Printf("Player %s tried to connect but game is %s", userID, game.Status)
		game.mu.Unlock()
		conn.Close()
		return
	}

	// Close old connection if exists
	if oldConn, exists := game.Connections[userID]; exists {
		oldConn.Close()
	}

	game.Connections[userID] = conn
	game.disconnected[userID] = false // Mark as connected
	connCount := len(game.Connections)

	log.Printf("Player %s connected via WebSocket (%d/2)", userID, connCount)

	// Start game only if BOTH players connected and game not started yet
	shouldStart := connCount == 2 && game.Status == "waiting_for_players"

	if shouldStart {
		game.Status = "in_progress"
		game.mu.Unlock()
		log.Printf("Both players ready! Starting game...")
		go runGame(game)
	} else {
		game.mu.Unlock()
	}

	// Listen for messages from this player
	go handlePlayerMessages(game, userID, conn)
}

func handlePlayerMessages(game *Game, userID string, conn *websocket.Conn) {
	defer func() {
		// Mark player as disconnected
		game.mu.Lock()
		game.disconnected[userID] = true
		game.mu.Unlock()

		conn.Close()
		log.Printf("Player %s disconnected", userID)

		// Check if game should end due to disconnection
		checkDisconnection(game)
	}()

	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Player %s connection error: %v", userID, err)
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

// Check if game should end due to player disconnection
func checkDisconnection(game *Game) {
	game.mu.Lock()
	defer game.mu.Unlock()

	// Only handle if game is in progress
	if game.Status != "in_progress" {
		return
	}

	// Check if any player is disconnected
	for playerID, disconnected := range game.disconnected {
		if disconnected {
			log.Printf("Player %s disconnected during game - ending game", playerID)

			// Find the other player (winner by default)
			var winner string
			for _, pid := range game.Players {
				if pid != playerID {
					winner = pid
					break
				}
			}

			// Mark game as finished
			game.Status = "finished"

			// Notify remaining player
			if conn, exists := game.Connections[winner]; exists {
				conn.WriteJSON(WSMessage{
					Type: "GAME_OVER",
					Payload: map[string]interface{}{
						"reason":  "opponent_disconnected",
						"winner":  winner,
						"results": game.Results,
					},
				})

				// Close after delay
				time.AfterFunc(3*time.Second, func() {
					conn.Close()
				})
			}
			return
		}
	}
}

func runGame(game *Game) {
	// Reset game state for new game
	game.mu.Lock()
	game.Results = []RoundResult{} // Clear previous results
	game.CurrentRound = 0
	game.mu.Unlock()

	// Send game start message
	broadcast(game, WSMessage{
		Type: "GAME_START",
		Payload: map[string]interface{}{
			"room_id":    game.RoomID,
			"max_rounds": game.MaxRounds,
			"players":    game.Players,
		},
	})

	time.Sleep(2 * time.Second) // Give players time to get ready

	// Run rounds
	for round := 1; round <= game.MaxRounds; round++ {
		// Check if anyone disconnected
		game.mu.Lock()
		anyDisconnected := false
		for _, disconnected := range game.disconnected {
			if disconnected {
				anyDisconnected = true
				break
			}
		}
		game.mu.Unlock()

		if anyDisconnected {
			log.Printf("Game ended early due to disconnection")
			return
		}

		game.mu.Lock()
		game.CurrentRound = round
		game.mu.Unlock()

		playRound(game, round)
		time.Sleep(3 * time.Second) // Pause between rounds
	}

	// Calculate final stats
	stats := make(map[string]map[string]interface{})
	for _, playerID := range game.Players {
		wins := 0
		totalLatency := int64(0)

		for _, result := range game.Results {
			if result.Winner == playerID {
				wins++
				totalLatency += result.Latency
			}
		}

		avgLatency := int64(0)
		if wins > 0 {
			avgLatency = totalLatency / int64(wins)
		}

		stats[playerID] = map[string]interface{}{
			"wins":          wins,
			"total_latency": totalLatency,
			"avg_latency":   avgLatency,
		}
	}

	winner := determineWinner(game)

	// Mark game as finished
	game.mu.Lock()
	game.Status = "finished"
	game.mu.Unlock()

	// Game over
	broadcast(game, WSMessage{
		Type: "GAME_OVER",
		Payload: map[string]interface{}{
			"reason":  "game_completed",
			"results": game.Results,
			"winner":  winner,
			"stats":   stats,
		},
	})

	log.Printf("Game finished")

	// Cleanup after delay
	time.Sleep(5 * time.Second)

	game.mu.Lock()
	for _, conn := range game.Connections {
		conn.Close()
	}
	game.Status = "completed"
	log.Printf("Room %s completed and closed", game.RoomID)
	game.mu.Unlock()
}

func playRound(game *Game, roundNum int) {
	game.mu.Lock()

	words := []string{"RED", "BLUE", "GREEN", "YELLOW"}
	colors := []string{"red", "blue", "green", "yellow"}

	// Create a new rand source with current time
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	word := words[r.Intn(len(words))]
	color := colors[r.Intn(len(colors))]

	log.Printf("🎨 Round %d: Word='%s' Color='%s'", roundNum, word, color) // ← DEBUG

	game.currentWord = word
	game.currentColor = color
	game.roundStartTime = time.Now()
	game.roundAnswered = false
	game.roundFinished = false
	game.roundWinner = ""
	game.wrongAnswers = make(map[string]bool)

	game.mu.Unlock()

	log.Printf("Round %d: Word='%s', Color='%s'", roundNum, word, color)

	// Broadcast round start
	broadcast(game, WSMessage{
		Type: "ROUND_START",
		Payload: map[string]interface{}{
			"round": roundNum,
			"word":  word,
			"color": color,
		},
	})

	// Wait for first correct answer (max 5 seconds)
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// Time's up, no one answered correctly
			game.mu.Lock()
			if !game.roundAnswered {
				log.Printf("Round %d timed out - no correct answer", roundNum)
				game.roundWinner = "timeout"
			}
			game.roundFinished = true // LOCK round - no more clicks!
			game.mu.Unlock()
			goto RoundEnd

		case <-ticker.C:
			// Check if round has been answered
			game.mu.Lock()
			answered := game.roundAnswered
			game.mu.Unlock()

			if answered {
				game.mu.Lock()
				game.roundFinished = true // LOCK round - no more clicks!
				game.mu.Unlock()
				goto RoundEnd
			}
		}
	}

RoundEnd:
	// Store result
	game.mu.Lock()
	result := RoundResult{
		Round:   roundNum,
		Word:    game.currentWord,
		Color:   game.currentColor,
		Winner:  game.roundWinner,
		Latency: game.roundLatency,
	}
	game.Results = append(game.Results, result)
	game.mu.Unlock()

	// Broadcast round result
	broadcast(game, WSMessage{
		Type: "ROUND_RESULT",
		Payload: map[string]interface{}{
			"round":      roundNum,
			"winner":     result.Winner,
			"latency_ms": result.Latency,
		},
	})
}

func handleClick(game *Game, userID string, payload map[string]interface{}) {
	game.mu.Lock()
	defer game.mu.Unlock()

	// Check if round is over
	if game.roundFinished {
		log.Printf("Player %s clicked but round already finished", userID)
		return
	}

	// Check if round already answered correctly
	if game.roundAnswered {
		log.Printf("Player %s clicked but round already won by someone else", userID)
		return
	}

	// Check if this player already got it wrong this round
	if game.wrongAnswers[userID] {
		log.Printf("Player %s BLOCKED. Already answered wrong this round", userID)
		return
	}

	// Get player's answer
	answer, ok := payload["answer"].(string)
	if !ok {
		log.Printf("Invalid answer from player %s", userID)
		return
	}

	// Calculate latency
	latency := time.Since(game.roundStartTime).Milliseconds()

	// Check if answer is correct (must match the COLOR, not the word!)
	correctAnswer := game.currentColor

	log.Printf("Player %s clicked '%s' (correct: '%s') - %dms",
		userID, answer, correctAnswer, latency)

	if answer == correctAnswer {
		// Correct answer!
		game.roundAnswered = true
		game.roundWinner = userID
		game.roundLatency = latency
		log.Printf("Player %s correct in %dms", userID, latency)
	} else {
		// WRONG - block this player from trying again
		game.wrongAnswers[userID] = true
		log.Printf("Player %s WRONG (blocked for this round)", userID)

		// Send feedback to client
		if conn, exists := game.Connections[userID]; exists {
			conn.WriteJSON(WSMessage{
				Type: "ROUND_FEEDBACK",
				Payload: map[string]interface{}{
					"message": "Wrong answer! Blocked for this round.",
				},
			})
		}
	}
}

func broadcast(game *Game, msg WSMessage) {
	game.mu.Lock()
	defer game.mu.Unlock()

	for _, conn := range game.Connections {
		conn.WriteJSON(msg)
	}
}

func determineWinner(game *Game) string {
	// Count wins and total latency per player
	type PlayerStats struct {
		Wins         int
		TotalLatency int64
	}

	stats := make(map[string]*PlayerStats)

	// Initialize stats for both players
	for _, playerID := range game.Players {
		stats[playerID] = &PlayerStats{
			Wins:         0,
			TotalLatency: 0,
		}
	}

	// Calculate stats from results
	for _, result := range game.Results {
		if result.Winner != "" && result.Winner != "timeout" {
			stats[result.Winner].Wins++
			stats[result.Winner].TotalLatency += result.Latency
		}
	}

	// Find winner by wins first, then by latency
	var winner string
	maxWins := 0
	lowestLatency := int64(9999999999)

	for playerID, playerStats := range stats {
		// Primary: Most wins
		if playerStats.Wins > maxWins {
			maxWins = playerStats.Wins
			lowestLatency = playerStats.TotalLatency
			winner = playerID
		} else if playerStats.Wins == maxWins && playerStats.Wins > 0 {
			// Tiebreaker: lowest latency
			if playerStats.TotalLatency < lowestLatency {
				lowestLatency = playerStats.TotalLatency
				winner = playerID
			}
		}
	}

	// Log the decision
	log.Printf("Final Scores:")
	for playerID, playerStats := range stats {
		log.Printf("- Player %s: %d wins, %dms total latency", playerID, playerStats.Wins, playerStats.TotalLatency)
	}

	// If no one won any rounds, it's a draw
	if winner == "" {
		log.Printf("Result: DRAW (0-0)")
		return "draw"
	}

	log.Printf("Winner: %s", winner)
	return winner
}
