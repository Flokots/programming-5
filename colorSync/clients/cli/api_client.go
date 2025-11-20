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
}

// newAPIClient creates a new APIClient
func newAPIClient() *APIClient {
	return &APIClient{
		userServiceURL: "http://localhost:8001",
		roomServiceURL: "http://localhost:8002",
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Login
type loginRequest struct {
	Username string `json:"username"`
}

type loginResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func (a *APIClient) login(username string) (string, error) {
	req := loginRequest{Username: username}
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

	// 404 = user not found (will register instead)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("user not found")
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

	return result.ID, nil
}

// REGISTER
type registerRequest struct {
	Username string `json:"username"`
}

type registerResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func (a *APIClient) register(username string) (string, error) {
	req := registerRequest{Username: username}
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

	return result.ID, nil
}