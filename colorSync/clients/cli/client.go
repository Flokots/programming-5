package main

import (
	"fmt"
	"log"
	"time"
)

// Client represents the CLI game client
type Client struct {
	username string // Player's username e.g "arbeiter"
	userID   string // UUID from user service e.g "25769518-e1de-4c7a-b7f5-c7648195898d"
	roomID   string // Room ID from room service e.g "6392b3fc-2745-46df-bba5-60390b4ad397"

	apiClient  *APIClient  // Pointer to HTTP client, handles the HTTP requests
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
	c.ui.showWelcome()

	// Login/Register
	fmt.Println("Registering user...")
	userID, err := c.apiClient.login(c.username)
	if err != nil {
		fmt.Println("User not found, registering...")
		userID, err = c.apiClient.register(c.username)
		if err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}
		fmt.Printf("Registered as %s\n", c.username)
	} else {
		fmt.Printf("Welcome back, %s!\n", c.username)
	}
	c.userID = userID

	// Join room
	fmt.Println("Joining matchmaking queue...")
	roomID, err := c.apiClient.joinRoom(userID)
	if err != nil {
		return fmt.Errorf("failed to join room: %w", err)
	}
	c.roomID = roomID
	log.Printf("Debug: Room ID = %s", roomID)

	// Wait for opponent
	fmt.Println("Waiting for opponent...")
	if err := c.waitForGameReady(); err != nil {
		return fmt.Errorf("failed waiting for game: %w", err)
	}

	// NOW connect to game
	log.Println("Connecting to game...")
	gameClient := newGameClient(c.roomID, c.userID, c.username, c.ui)
	if err := gameClient.connect(); err != nil {
		return fmt.Errorf("failed to connect to game: %w", err)
	}
	
	// Play game (this will block until game ends)
	err = gameClient.playGame()
	gameClient.close()

	if err != nil {
		return fmt.Errorf("game error: %w", err)
	}

	// show exit message
	fmt.Println()
    c.ui.showInfo("💡 To play again, run:")
    fmt.Printf("   go run . --username %s\n", c.username)
    fmt.Println()
    fmt.Println("👋 Thanks for playing!")

	return nil
}

//  Wait for game to be ready
func (c *Client) waitForGameReady() error {
	maxAttempts := 30 // 30 seconds max wait

	for range maxAttempts {
		// Check if game exists
		ready, err := c.apiClient.checkGameReady(c.roomID)
		if err != nil {
			log.Printf("Error checking game status: %v", err)
		}

		if ready {
			fmt.Println("Opponent found! Starting game...")
			return nil
		}

		// Wait 1 second before checking again
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for opponent")
}
