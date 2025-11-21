package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// GameClient handles WebSocket connection and game logic
type GameClient struct {
	roomID   string
	userID   string
	username string
	conn     *websocket.Conn
	ui       *UI

	gameActive bool // Track if game is active
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// newGameClient creates a new game client
func newGameClient(roomID, userID, username string, ui *UI) *GameClient {
	return &GameClient{
		roomID:     roomID,
		userID:     userID,
		username:   username,
		ui:         ui,
		gameActive: false,
	}
}

// connect establishes WebSocket connection
func (g *GameClient) connect() error {
	url := fmt.Sprintf("ws://localhost:8003/game/ws?room_id=%s&user_id=%s",
		g.roomID, g.userID)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to game: %w", err)
	}

	g.conn = conn
	log.Printf("Connected to game via WebSocket")
	return nil
}

// close closes the WebSocket connection
func (g *GameClient) close() {
	if g.conn != nil {
		g.conn.Close()
	}
}

// playGame runs the main game loop
func (g *GameClient) playGame() error {
	messageChan := make(chan WSMessage)
	errorChan := make(chan error)
	done := make(chan struct{}) // Signal to stop goroutine

	// Goroutine to receive messages
	go func() {
		defer close(messageChan) // Clean shutdown
		for {
			var msg WSMessage
			err := g.conn.ReadJSON(&msg)
			if err != nil {
				select {
				case errorChan <- err:
				case <-done: // Don't block if main loop exited
					return
				}
				return
			}

			select {
			case messageChan <- msg:
			case <-done: // Don't block if main loop exited
				return
			}
		}
	}()

	// Main game loop
	for {
		select {
		case msg, ok := <-messageChan:
			if !ok {
				return nil // Channel closed, exit loop
			}
			gameOver := g.handleMessage(msg)
			if gameOver {
				close(done)                        // Signal goroutine to stop
				time.Sleep(100 * time.Millisecond) // Give goroutie time to exit
				return nil
			}

		case err := <-errorChan:
			close(done) // Signal goroutine to stop
			// Only report error if game is still active
			if g.gameActive {
				return fmt.Errorf("connection error: %w", err)
			}
			return nil // Game ended, ignore connection errors
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (g *GameClient) handleMessage(msg WSMessage) bool {
	switch msg.Type {
	case "GAME_START":
		g.handleGameStart(msg)

	case "ROUND_START":
		g.handleRoundStart(msg)

	case "ROUND_RESULT":
		g.handleRoundResult(msg)

	case "GAME_OVER":
		g.handleGameOver(msg)
		g.conn.Close() // Close connection immediately!
		return true    // Game finished

	case "WRONG_ANSWER":
		g.ui.showError("❌ Wrong! Blocked for this round.")

	case "ERROR":
		if errMsg, ok := msg.Payload["message"].(string); ok {
			g.ui.showError(errMsg)
		}

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}

	return false
}

// handleGameStart processes GAME_START message
func (g *GameClient) handleGameStart(msg WSMessage) {
	maxRounds := int(msg.Payload["max_rounds"].(float64))

	g.ui.showGameStart(maxRounds)
	g.gameActive = true // Game is now active
}

// handleRoundStart processes ROUND_START message and gets player input
func (g *GameClient) handleRoundStart(msg WSMessage) {
	round := int(msg.Payload["round"].(float64))
	word := msg.Payload["word"].(string)
	color := msg.Payload["color"].(string)

	// Display the Stroop test
	g.ui.showRound(round, word, color)

	// Get player input in a goroutine (non-blocking)
	go g.handlePlayerInput()
}

// handlePlayerInput waits for player to click a color
func (g *GameClient) handlePlayerInput() {
	reader := bufio.NewReader(os.Stdin)

	// Read input
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	// Check if game is still active
	if !g.gameActive { //Ignore input if game ended
		return
	}

	// Map shortcuts to full color names
	colorMap := map[string]string{
		"r":      "red",
		"b":      "blue",
		"g":      "green",
		"y":      "yellow",
		"red":    "red",
		"blue":   "blue",
		"green":  "green",
		"yellow": "yellow",
	}

	answer, valid := colorMap[input]
	if !valid {
		g.ui.showError("Invalid input! Use: r/b/g/y or red/blue/green/yellow")
		return
	}

	// Send click to server
	g.sendClick(answer)
}

// sendClick sends a CLICK message to the server
func (g *GameClient) sendClick(answer string) {
	msg := WSMessage{
		Type: "CLICK",
		Payload: map[string]interface{}{
			"answer": answer,
		},
	}

	if err := g.conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send click: %v", err)
	}
}

// handleRoundResult - displays ROUND_RESULT
func (g *GameClient) handleRoundResult(msg WSMessage) {
	round := int(msg.Payload["round"].(float64))

	// Safely handle winner (might be nil, "timeout", or userID)
	var winner string
	if winnerVal, ok := msg.Payload["winner"]; ok && winnerVal != nil {
		winner = winnerVal.(string)
	}

	// Safely handle latency (might not exist for timeout)
	var latency int64
	if latencyFloat, ok := msg.Payload["latency_ms"].(float64); ok {
		latency = int64(latencyFloat)
	}

	// Display result - pass winner string directly
	g.ui.showRoundResult(round, winner, g.userID, latency)
}

// handleGameOver processes GAME_OVER message
func (g *GameClient) handleGameOver(msg WSMessage) {
	g.gameActive = false //  Deactivate game (ignore pending inputs)

	// Safely get reason
	reason := ""
	if r, ok := msg.Payload["reason"].(string); ok {
		reason = r
	}

	// Safely get winner
	winner := ""
	if w, ok := msg.Payload["winner"].(string); ok {
		winner = w
	}

	// Handle disconnection case
	if reason == "opponent_disconnected" {
		if winner == g.userID {
			g.ui.showInfo("🎉 Opponent disconnected - You win by default!")
		} else {
			g.ui.showInfo("You disconnected from the game")
		}
		time.Sleep(3 * time.Second)
		return
	}

	// Normal game end - safely get stats from backend
	stats, ok := msg.Payload["stats"].(map[string]interface{})
	if !ok {
		g.ui.showError("Error: Invalid stats data")
		return
	}

	myStatsData, ok := stats[g.userID].(map[string]interface{})
	if !ok {
		g.ui.showError("Error: Could not find your stats")
		return
	}

	// Safely extract stats with defaults
	wins := 0
	if w, ok := myStatsData["wins"].(float64); ok {
		wins = int(w)
	}

	totalLatency := int64(0)
	if tl, ok := myStatsData["total_latency"].(float64); ok {
		totalLatency = int64(tl)
	}

	avgLatency := int64(0)
	if al, ok := myStatsData["avg_latency"].(float64); ok {
		avgLatency = int64(al)
	}

	// Get opponent's wins (to calculate losses)
	opponentWins := 0
	for uid, statsData := range stats {
		if uid != g.userID {
			if opData, ok := statsData.(map[string]interface{}); ok {
				if w, ok := opData["wins"].(float64); ok {
					opponentWins = int(w)
				}
			}
			break
		}
	}

	// Display game over screen
	g.ui.showGameOver(winner, g.userID, wins, opponentWins, totalLatency, avgLatency)

	// Play again prompt
	fmt.Println()
	fmt.Print("Play again? [y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		fmt.Println("\n🔄 Restarting... Run the command again:")
		fmt.Printf("   go run . --username %s\n", g.username)
	} else {
		fmt.Println("\n👋 Thanks for playing! Goodbye!")
	}

	time.Sleep(2 * time.Second)
}
