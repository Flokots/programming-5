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
	users       = make(map[string]*User) //userID -> User
	usersByName = make(map[string]*User) //username -> User
	mu          sync.RWMutex             //Mutex for thread-safe access
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from React DEV server
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Create a new ServerMux(router)
	mux := http.NewServeMux()


	// Register routes
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/users/", getUserHandler) // trailing slash for /users/{id}
	mux.HandleFunc("/health", healthHandler)

	handler := corsMiddleware(mux) // Wrap with CORS middleware
	
	port := ":8001"
	fmt.Printf("User Service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, handler))
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

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Extract user ID from URL path
	// URL format: /users/{id}
	// Example: /users/ceb3499b-e0ca-4b3f-af07-5dcc287d0ac7
	const usersPrefix = "/users/"
	path := r.URL.Path
	if len(path) <= len(usersPrefix) {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}
	userID := path[len(usersPrefix):]
	// 3. Look up user (thread-safe read)
	mu.RLock()
	user, exists := users[userID]
	mu.RUnlock()

	// 4. Check if user exists
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 5. Return user info
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	})

	log.Printf("Retrieved user: %s (ID: %s)", user.Username, user.ID)
}

type LoginRequest struct {
	Username string `json:"username"`
}

type LoginResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Parse JSON from request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Validate username
	if req.Username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	// 4. Check if user exists (thread-safe read)
	mu.RLock()
	user, exists := usersByName[req.Username]
	mu.RUnlock()

	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 5. Return user info (successful login)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		ID:       user.ID,
		Username: user.Username,
		Message:  "Login successful",
	})

	log.Printf("User logged in: %s (ID: %s)", user.Username, user.ID)
}
