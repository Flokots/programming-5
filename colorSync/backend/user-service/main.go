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
	http.HandleFunc("/login", loginHandler)
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

