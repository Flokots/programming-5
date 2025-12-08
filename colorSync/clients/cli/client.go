package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// Client represents the CLI game client
type Client struct {
	username  string     // Player's username e.g "arbeiter"
	userID    string     // UUID from user service e.g "25769518-e1de-4c7a-b7f5-c7648195898d"
	roomID    string     // Room ID from room service e.g "6392b3fc-2745-46df-bba5-60390b4ad397"
	apiClient *APIClient // Pointer to HTTP client, handles the HTTP requests
	ui        *UI        // Pointer to UI renderer, handles terminal display
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

	// Prompt for username if not provided
	if strings.TrimSpace(c.username) == "" {
		c.username = promptForUsername()
	}

	// Prompt for password
	fmt.Print("Enter password: ")
	password := promptForPassword()

	// Try login first (with password)
	fmt.Println("Logging in user...")
	userID, err := c.apiClient.login(c.username, password)
	if err != nil {
		// If login fails, try registration
		fmt.Println("User not found, registering...")
		userID, err = c.apiClient.register(c.username, password) // Pass password for registration
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

	// STEP 1: Wait for opponent (room becomes full)
	fmt.Println("Waiting for opponent...")
	if err := c.waitForRoomFull(); err != nil {
		return fmt.Errorf("failed waiting for opponent: %w", err)
	}

	// STEP 2: Room is full; wait briefly for Game Service to be notified/created
	fmt.Println("Opponent found! Preparing game...")
	if err := c.waitForGameReady(); err != nil {
		return fmt.Errorf("failed waiting for game: %w", err)
	}

	// NOW connect to game
	log.Println("Connecting to game...")
	gameClient := newGameClient(c.roomID, c.userID, c.username, c.ui)
	if err := gameClient.connect(); err != nil {
		// 🆕 Best-effort cleanup on connection failure
		_ = c.apiClient.leaveRoom(c.roomID)
		return fmt.Errorf("failed to connect to game: %w", err)
	}

	// Play game (this will block until game ends)
	err = gameClient.playGame()
	gameClient.close()

	// Always leave the room after game ends (best-effort)
	if leaveErr := c.apiClient.leaveRoom(c.roomID); leaveErr != nil {
		log.Printf("Warning: failed to leave room: %v", leaveErr)
	}

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

// Wait until room has 2 players
func (c *Client) waitForRoomFull() error {
	const maxAttempts = 60
	for range maxAttempts {
		full, err := c.apiClient.checkRoomFull(c.roomID)
		if err != nil {
			log.Printf("Error checking room status: %v", err)
		}
		if full {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for opponent")
}

// Wait until game exists (Game Rules Service has created it)
func (c *Client) waitForGameReady() error {
	const maxAttempts = 15
	for range maxAttempts {
		ready, err := c.apiClient.checkGameReady(c.roomID)
		if err != nil {
			log.Printf("Error checking game status: %v", err)
		}
		if ready {
			fmt.Println("Game ready!")
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for game to start")
}
