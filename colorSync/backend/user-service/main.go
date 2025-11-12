package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// User represents a registered user
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// In-memory storage
var (
	users   = make(map[string]*User) //userID -> User
	usersByName = make(map[string]*User) //username -> User
	mu      sync.RWMutex //Mutex for thread-safe access
)

func main() {
	// Register routes
	http.HandleFunc("/register", registerHandler)
	//http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/users/", getUserHandler)  // trailing slash for /users/{id}
	http.HandleFunc("/health", healthHandler)

	port := ":8001"
	fmt.Printf("User Service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

type RegisterRequest struct {
	Username string `json:"username"`
}

type RegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Parse JSON from request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Validate username 
	if req.Username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	// 4. Check if username already exists (thread-safe read)
	mu.RLock()
	_, exists := usersByName[req.Username]
	mu.RUnlock()

	if exists {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	// 5. Create new user
	user := &User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		CreatedAt: time.Now(),
	}

	// 6. Store user in memory (thread-safe write)
	mu.Lock()
	users[user.ID] = user
	usersByName[user.Username] = user
	mu.Unlock()

	// 7. Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Message:  "User registered successfully",
	})

	log.Printf("Registered new user: %s (ID: %s)", user.Username, user.ID)
}

