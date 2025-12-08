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
		token:          "",
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

// STEP 1: Check if ROOM is full (has 2 players)
func (a *APIClient) checkRoomFull(roomID string) (bool, error) {
	url := fmt.Sprintf("%s/room/%s/ready", a.roomServiceURL, roomID)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var result struct {
		Ready   bool     `json:"ready"`
		Players []string `json:"players"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Ready, nil
}

// STEP 2: Check if GAME is ready (exists)
func (a *APIClient) checkGameReady(roomID string) (bool, error) {
	url := fmt.Sprintf("http://localhost:8003/game/status?room_id=%s", roomID)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		// Read body for logging/debug
		bodyBytes, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status checking game: %d %s", resp.StatusCode, string(bodyBytes))
	}

	// We only care that the game exists (200 OK). Status may be "waiting_for_players" until sockets connect.
	return true, nil
}

// Leave active room (uses JWT for user identity)
func (a *APIClient) leaveRoom(roomID string) error {
	url := fmt.Sprintf("%s/rooms/%s/leave", a.roomServiceURL, roomID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return fmt.Errorf("failed to create leave request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("leave room failed: %s", string(body))
	}
	return nil
}
