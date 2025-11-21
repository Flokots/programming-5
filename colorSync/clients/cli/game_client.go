package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// GameClient handles the game logic and WebSocket connection
type GameClient struct {
	roomID	 string
	userID	 string
	username string
	conn *websocket.Conn
	ui  *UI

	// Game state
	myScore    int
	opponentScore int
	currentRound int
}

// WSMessage represents a message sent/received via WebSocket
type WSMessage struct {
	Type    string          `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// newGameClient creates a new GameClient
func newGameClient(roomID, userID, username string, ui *UI) *GameClient{
	return &GameClient{
		roomID:   roomID,
		userID:   userID,
		username: username,
		ui:       ui,
	}
}

// connect establishes a WebSocket connection to the game server
func (g *GameClient) connect() error {
	url := fmt.Sprintf("ws://localhost:8003/game/ws?room_id=%s&user_id=%s", g.roomID, g.userID)

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
	// Start listening for messages
	messageChan := make(chan WSMessage)
	errorChan := make(chan error)

	// Goroutine to receive messages
	go func() {
		for {
			var msg WSMessage
			err := g.conn.ReadJSON(&msg)
			if err != nil {
				errorChan <- err
				return
			}
			messageChan <- msg
		}
	}()

	// Main game loop 
	for {
		select {
		case msg := <- messageChan:
			gameOver := g.handleMessage(msg)
			if gameOver {
				return nil
			}

		case err := <- errorChan:
			return fmt.Errorf("connection error: %w", err)
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (g *GameClient) handleMessage(msg WSMessage) bool {
	switch msg.Type {
		case "GAME_START":
			g.handleGameStart(msg)

		case "ROUND_START":
			g.handleRoundResult(msg)

		case "ROUND_RESULT":
			g.handleRoundResult(msg)
			
		case "GAME_OVER":
			g.handleGameOver(msg)
			return true // Game finished

		case "WRONG_ANSWER":
			g.ui.showError("Wrong answer! Blocked for this round.")
		
		case "ERROR":
			if errMsg, ok := msg.Payload["message"].(string); ok {
				g.ui.showError(errMsg)
			}
		
		default:
			log.Printf("Unknown message type: %s", msg.Type)
	}
	return false
}

// handleGameStart processes the GAME_START message
func (g *GameClient) handleGameStart(msg WSMessage) {
	maxRounds := int(msg.Payload["max_rounds"].(float64))
	
	g.ui.showGameStart(maxRounds)
	g.myScore = 0
	g.opponentScore = 0
	g.currentRound = 0
}

// handleRoundStart processes the ROUND_START message and gets player input
func (g *GameClient) handleRoundStart(msg WSMessage) {
	round := int(msg.Payload["round"].(float64))
	word := msg.Payload["word"].(string)
	color := msg.Payload["color"].(string)

	g.currentRound = round

	// Display the stroop test
	g.ui.showRound(round, word, color)

	// Get player input in a goroutine (non-blocking)
	go g.handlePlayerInput()
}

// handlePlayerInput waits for player to click a color (select color)
func (g *GameClient) handlePlayerInput() {
	reader := bufio.NewReader(os.Stdin)

	// Read input 
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	// Map shortcuts to full color names
	colorMap := map[string]string{
		"r": "red",
		"b": "blue",
		"g": "green",
		"y": "yellow",
		"red": "red",
		"blue": "blue",
		"green": "green",
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
		log.Printf("Failed to send CLICK: %v", err)
	}
}

// handleRoundResult processes the ROUND_RESULT message
func (g *GameClient) handleRoundResult(msg WSMessage) {
	round := int(msg.Payload["round"].(float64))
	winner := msg.Payload["winner"].(string)

	var latency int64
	if latencyFloat, ok := msg.Payload("latency_ms").(float64); ok {
		latency = int64(latencyFloat)
	}

	// Update scores
	if winner == g.userID {
		g.myScore++
	} else if winner != "" && winner != "timeout" {
		g.opponentScore++
	}
	
	// Display round result
	iWon := winner == g.userID
	isDraw := winner == "timeout" || winner == ""

	g.ui.showRoundResult(round, iWon, isDraw, latency, g.myScore, g.opponentScore)
}

// handleGameOver processes the GAME_OVER message
func (g *GameClient) handleGameOver(msg WSMessage) {
	reason := msg.Payload["reason"].(string)
	winner := msg.Payload["winner"].(string)

	// Handle disconnection case
	if reason == "opponent_disconnected" {
		if winner == g.userID {
			g.ui.showInfo("Opponent disconnected. You win by default!")
		} else {
			g.ui.showInfo("You disconnected from the game.")
		}
		time.Sleep(3 * time.Second)
		return
	}

	// Normal game end
	stats := msg.Payload["stats"].(map[string]interface{})
	myStats := stats[g.userID].(map[string]interface{})

	wins := int(myStats["wins"].(float64))
	totalLatency := int64(myStats["total_latency"].(float64))
	avgLatency := int64(myStats["avg_latency"].(float64))

	iWon := winner == g.userID
	isDraw := winner == "draw"

	g.ui.showGameOver(iWon, isDraw, wins, g.myScore-wins,  avgLatency, totalLatency)

	// Ask if player wants to play again
	fmt.Println()
	fmt.Println("Play again? [y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		fmt.Println("\nRestarting.. Run the command again to play another game!")
		fmt.Printf("go run . --username %s\n", g.username)
	} else {
		fmt.Println("Thanks for playing! Goodbye.")
	}

	time.Sleep(2 * time.Second)
}