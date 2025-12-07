package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

    "golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"

	"github.com/Flokots/programming-5/colorSync/shared/auth"
)

// User represents a registered user
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`   // Hashed password
	CreatedAt time.Time `json:"created_at"`
}

// In-memory storage
var (
	users       = make(map[string]*User) // userID -> User
	usersByName = make(map[string]*User) // username -> User
	mu          sync.RWMutex             // Mutex for thread-safe access
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
	fmt.Printf("Endpoints:\n")
    fmt.Printf("   POST /register - Create new user (username + password)\n")
    fmt.Printf("   POST /login    - Authenticate user (returns JWT token)\n")
    fmt.Printf("   GET  /users/:id - Get user info\n")
    fmt.Printf("   GET  /health   - Health check\n")
    fmt.Printf("\n")
	log.Fatal(http.ListenAndServe(port, handler))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	userCount := len(users)
	mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "healthy",
		"service":    "user-service",
		"users_count": userCount, // show user count
	})
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Token     string    `json:"token"` 
	CreatedAt time.Time `json:"created_at"`
	Message  string `json:"message"`
}

// registerHandler creates a new user with hashed password and returns a JWT token
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

	// 4. Validate username length
	if len(req.Username) < 3 {
        http.Error(w, "Username must be at least 3 characters", http.StatusBadRequest)
        return
    }

	// 5. Validate password
	if req.Password == "" {
        http.Error(w, "Password required", http.StatusBadRequest)
        return
    }

	if len(req.Password) < 6 {
        http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
        return
    }

	// 6. Check if username already exists (thread-safe read)
	mu.RLock()
	_, exists := usersByName[req.Username]
	mu.RUnlock()

	if exists {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	// 7. Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(req.Password),
        bcrypt.DefaultCost, // Cost factor 10
    )
    if err != nil {
        log.Printf("Failed to hash password: %v", err)
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }

	// 8. Create new user
	user := &User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	// 9. Store user in memory (thread-safe write)
	mu.Lock()
	users[user.ID] = user
	usersByName[user.Username] = user
	mu.Unlock()

	// 10. Generate JWT token
	token, err := auth.GenerateUserToken(user.ID, user.Username)
    if err != nil {
        log.Printf("Failed to generate token: %v", err)
        http.Error(w, "User created but failed to generate token", http.StatusInternalServerError)
        return
    }

	log.Printf("User registered: %s (ID: %s)", user.Username, user.ID)
    log.Printf("JWT token generated for: %s", user.Username)

	// 11. Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Token:     token,
        CreatedAt: user.CreatedAt,
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
	Password string `json:"password"`
}

type LoginResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Token     string    `json:"token"`
    CreatedAt time.Time `json:"created_at"`
	Message  string `json:"message"`
}

// loginHandler authenticates user and returns JWT token
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
	if req.Username == "" || req.Password == "" {
        http.Error(w, "Username and password required", http.StatusBadRequest)
        return
    }

	// 4. Check if user exists (thread-safe read)
	mu.RLock()
	user, exists := usersByName[req.Username]
	mu.RUnlock()

	if !exists {
        // Use generic error to prevent username enumeration
        log.Printf("Login attempt for non-existent user: %s", req.Username)
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

	// 5. Verify password using bcrypt
	err := bcrypt.CompareHashAndPassword(
        []byte(user.Password), // Hashed password from storage
        []byte(req.Password),  // Plain text password from request
    )
    if err != nil {
        // Wrong password - use same generic error
        log.Printf("Failed login attempt for user: %s (wrong password)", req.Username)
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

	// 6. Generate JWT token
    token, err := auth.GenerateUserToken(user.ID, user.Username)
    if err != nil {
        log.Printf("Failed to generate token: %v", err)
        http.Error(w, "Login successful but failed to generate token", http.StatusInternalServerError)
        return
    }

	log.Printf("User logged in: %s (ID: %s)", user.Username, user.ID)
    log.Printf("JWT token generated for: %s", user.Username)

	// 6. Return user info (successful login) with JWT token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		ID:       user.ID,
		Username: user.Username,
		Token:     token,
		CreatedAt: user.CreatedAt,
		Message:  "Login successful",
	})
}
