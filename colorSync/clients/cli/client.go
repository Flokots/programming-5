package main

import (
	"fmt"
	"log"
	"os/user"
)

// Client represents the CLI game client
type Client struct {
	username string // Player's username e.g "arbeiter"
	userID   string // UUID from user service e.g "25769518-e1de-4c7a-b7f5-c7648195898d"
	roomID   string // Room ID from room service e.g "6392b3fc-2745-46df-bba5-60390b4ad397"

	apiClient  *APIClient  // Pointer to HTTP client, handles the HTTP requests
	gameClient *GameClient // Pointer to WebSocket client, handles the real-time game connection
	ui         *UI         // Pointer to UI renderer, handles terminal display
}

// newClient creates and initializes a new Client instance
func newClient(username string) *Client {
	return &Client{
		username:  username,
		apiClient: newAPIClient(), // Initialize the API client
		ui:        newUI(),        // Initialize the UI renderer
	}
}

// Run executes the main game flow
func (c *Client) Run() error {
	// Show welcome screen
	c.ui.ShowWelcome()

	// STEP 1: Register user
	c.ui.ShowInfo("Registering user...")

	userID, err := c.apiClient.Register(c.username)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}
	c.userID = userID

	c.ui.ShowSuccess(fmt.Sprintf("Registered as %s", c.username))
	log.Printf("DEBUG: User ID = %s", userID)

	// STEP 2: Join matchmaking queue
	c.ui.ShowInfo("Joining matchmaking queue...")

	roomID, err := c.apiClient.JoinRoom(userID)
	if err != nil {
		return fmt.Errorf("failed to join room: %w", err)
	}
	c.roomID = roomID

	log.Printf("Debug: Room ID = %s", roomID)

	// STEP 3: Connect to game server via WebSocket
	c.ui.ShowInfo("Waiting for opponent...")

	// Create game client
	c.gameClient = newGameClient(c.roomID, c.userID, c.username, c.ui)

	// Connect WebSocket
	if err := c.gameClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to game: %w", err)
	}
	defer c.gameClient.Close() // Always close when done

	// STEP 4: Start game loop
	return c.gameClient.PlayGame()
}