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

// NewClient creates and initializes a new Client instance
func NewClient(username string) *Client {
	return &Client{
		username:  username,
		apiClient: NewAPIClient(), // Initialize the API client
		ui:        NewUI(),        // Initialize the UI renderer
	}
}
