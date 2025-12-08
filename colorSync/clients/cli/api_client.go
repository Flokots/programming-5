package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient handles HTTP requests to backend services
type APIClient struct {
	userServiceURL string
	roomServiceURL string
	httpClient     *http.Client
	token          string
}

// newAPIClient creates a new APIClient
func newAPIClient() *APIClient {
	return &APIClient{
		userServiceURL: "http://localhost:8001",
		roomServiceURL: "http://localhost:8002",
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		token:		  	"",
	}
}

// Login
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"` 
	Message  string `json:"message"`
}

func (a *APIClient) login(username, password string) (string, error) {
	req := loginRequest{
		Username: username, 
		Password: password,
	}
	body, _ := json.Marshal(req)

	resp, err := a.httpClient.Post(
		a.userServiceURL+"/login",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for 401 
    if resp.StatusCode == http.StatusUnauthorized {
        return "", fmt.Errorf("invalid username or password")
    }

	// Other errors
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed: %s", string(bodyBytes))
	}

	// Success
	var result loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Store token
    a.token = result.Token

	return result.ID, nil
}

// REGISTER
type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"` 
	Message  string `json:"message"`
}

func (a *APIClient) register(username, password string) (string, error) {
	req := registerRequest{
		Username: username, 
		Password: password,
	}
	body, _ := json.Marshal(req)

	resp, err := a.httpClient.Post(
		a.userServiceURL+"/register",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	// Other errors
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("registration failed: %s", string(bodyBytes))
	}

	// Success
	var result registerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	a.token = result.Token

	return result.ID, nil
}

// JOIN ROOM
type joinRoomRequest struct {
	UserID string `json:"user_id"`
}

type joinRoomResponse struct {
	RoomID  string `json:"room_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Add authorization header to request
func (a *APIClient) joinRoom(userID string) (string, error) {
	req := joinRoomRequest{UserID: userID}
	body, _ := json.Marshal(req)

	// Create request with Authorization header
    httpReq, err := http.NewRequest(
        "POST",
        a.roomServiceURL+"/join",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

	// Set headers
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+a.token) // add JWT token

	resp, err := a.httpClient.Do(httpReq)
    if err != nil {
        return "", fmt.Errorf("connection failed: %w", err)
    }
	defer resp.Body.Close()

	// Handle auth errors
	if resp.StatusCode == http.StatusUnauthorized {
        return "", fmt.Errorf("authentication failed - token may be expired")
    }
	if resp.StatusCode == http.StatusForbidden {
        return "", fmt.Errorf("access denied - user ID mismatch")
    }

	// other errors
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("join room failed: %s", string(bodyBytes))
	}

	// Success
	var result joinRoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.RoomID, nil
}

// Check if game is ready for a room
func (a *APIClient) checkGameReady(roomID string) (bool, error) {
	resp, err := a.httpClient.Get(
		fmt.Sprintf("http://localhost:8003/game/status?room_id=%s", roomID),
	)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// If game exists and is waiting for players, it's ready
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	// Game doesn't exist yet
	return false, nil
}
