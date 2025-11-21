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