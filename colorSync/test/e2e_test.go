package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const (
	userServiceURL = "http://localhost:8001"
	roomServiceURL = "http://localhost:8002"
	gameServiceURL = "http://localhost:8003"
)

type User struct {
	Username string
	Password string
	UserID   string
	Token    string
}

type GameMessage struct {
	Type      string `json:"type"`
	Word      string `json:"word,omitempty"`
	Color     string `json:"color,omitempty"`
	YourColor string `json:"your_color,omitempty"`
	Winner    string `json:"winner,omitempty"`
	Message   string `json:"message,omitempty"`
	Round     int    `json:"round,omitempty"`
}

var testLogger = log.New(os.Stdout, "", log.LstdFlags)

func TestE2EGameFlow(t *testing.T) {
	// Check if services are running
	t.Log("Checking if backend services are running...")
	if !checkServicesRunning(t) {
		t.Fatal("❌ Backend services are not running. Please start them first.")
	}

	t.Log("✅ All backend services are running")

	// Test 1: User Registration and Login
	t.Run("RegisterAndLogin", func(t *testing.T) {
		t.Log("Testing user registration and login...")
		peter := registerAndLogin(t, "peter_test", "password123")
		t.Logf("Peter registered: UserID=%s, Token len=%d", peter.UserID, len(peter.Token))

		pam := registerAndLogin(t, "pam_test", "password456")
		t.Logf("Pam registered: UserID=%s, Token len=%d", pam.UserID, len(pam.Token))

		if peter.UserID == "" {
			t.Fatal("❌ Peter UserID is empty")
		}
		if peter.Token == "" {
			t.Fatal("❌ Peter token is empty")
		}
		if pam.UserID == "" {
			t.Fatal("❌ Pam UserID is empty")
		}
		if pam.Token == "" {
			t.Fatal("❌ Pam token is empty")
		}
		t.Log("✅ Users registered and logged in successfully")
	})

	// Test 2: Full Game Flow
	t.Run("FullGameFlow", func(t *testing.T) {
		testLogger.Println("========== FULL GAME FLOW TEST ==========")
		peter := registerAndLogin(t, fmt.Sprintf("peter_%d", time.Now().Unix()), "password123")
		pam := registerAndLogin(t, fmt.Sprintf("pam_%d", time.Now().Unix()), "password456")

		testLogger.Printf("Peter: UserID=%s", peter.UserID)
		testLogger.Printf("Pam: UserID=%s", pam.UserID)

		var wg sync.WaitGroup
		wg.Add(2)

		peterResults := make(chan string, 1)
		pamResults := make(chan string, 1)
		errors := make(chan error, 2)

		// Peter's goroutine
		go func() {
			defer wg.Done()

			// Join matchmaking
			testLogger.Println("🔵 Peter: Joining matchmaking...")
			roomID, err := joinMatchmaking(peter)
			if err != nil {
				errors <- fmt.Errorf("peter join: %w", err)
				return
			}
			testLogger.Printf("✅ Peter joined room: %s", roomID)

			// Wait for room full
			testLogger.Println("🔵 Peter: Waiting for opponent...")
			if err := waitForRoomFull(roomID, 30*time.Second); err != nil {
				errors <- fmt.Errorf("peter wait room: %w", err)
				return
			}
			testLogger.Println("✅ Peter: Opponent found!")

			// Wait for game ready
			if err := waitForGameReady(roomID, 15*time.Second); err != nil {
				errors <- fmt.Errorf("peter wait game: %w", err)
				return
			}
			testLogger.Println("✅ Peter: Game starting...")

			// Play game (with random guesses to test mechanics)
			winner, err := playGameWithRandomAnswers(peter, roomID)
			if err != nil {
				errors <- fmt.Errorf("peter gameplay: %w", err)
				return
			}
			peterResults <- winner

			// Leave room
			if err := leaveRoom(peter, roomID); err != nil {
				errors <- fmt.Errorf("peter leave: %w", err)
				return
			}
			testLogger.Println("✅ Peter: Left room")
		}()

		// Pam's goroutine
		go func() {
			defer wg.Done()
			time.Sleep(2 * time.Second) // Let Peter create room

			testLogger.Println("🔴 Pam: Joining matchmaking...")
			roomID, err := joinMatchmaking(pam)
			if err != nil {
				errors <- fmt.Errorf("pam join: %w", err)
				return
			}
			testLogger.Printf("✅ Pam joined room: %s", roomID)

			// Wait for game ready (room already full)
			if err := waitForGameReady(roomID, 15*time.Second); err != nil {
				errors <- fmt.Errorf("pam wait game: %w", err)
				return
			}
			testLogger.Println("✅ Pam: Game starting...")

			// Play game
			winner, err := playGameWithRandomAnswers(pam, roomID)
			if err != nil {
				errors <- fmt.Errorf("pam gameplay: %w", err)
				return
			}
			pamResults <- winner

			// Leave room
			if err := leaveRoom(pam, roomID); err != nil {
				errors <- fmt.Errorf("pam leave: %w", err)
				return
			}
			testLogger.Println("✅ Pam: Left room")
		}()

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			if err != nil {
				t.Fatal(err)
			}
		}

		// Verify both got same winner
		peterWinner := <-peterResults
		pamWinner := <-pamResults
		if peterWinner != pamWinner {
			t.Fatalf("❌ Winners don't match: peter saw %s, pam saw %s", peterWinner, pamWinner)
		}

		testLogger.Printf("🏆 Game completed! Winner: %s", peterWinner)
		testLogger.Println("========== ✅ FULL GAME FLOW PASSED ==========")
	})

	// Test 3: Rejoin After Game
	t.Run("RejoinAfterGame", func(t *testing.T) {
		testLogger.Println("========== REJOIN TEST ==========")
		peter := registerAndLogin(t, fmt.Sprintf("peter_rejoin_%d", time.Now().Unix()), "password123")

		// First join
		testLogger.Println("Peter: First join...")
		roomID1, err := joinMatchmaking(peter)
		if err != nil {
			t.Fatalf("❌ First join failed: %v", err)
		}
		testLogger.Printf("✅ Peter joined room: %s", roomID1)

		// Leave room
		testLogger.Println("Peter: Leaving first room...")
		if err := leaveRoom(peter, roomID1); err != nil {
			t.Fatalf("❌ Leave failed: %v", err)
		}
		testLogger.Println("✅ Peter left first room")

		// Rejoin immediately
		testLogger.Println("Peter: Rejoining...")
		roomID2, err := joinMatchmaking(peter)
		if err != nil {
			t.Fatalf("❌ Rejoin failed (should succeed): %v", err)
		}
		testLogger.Printf("✅ Peter rejoined room: %s", roomID2)

		// Leave again
		testLogger.Println("Peter: Leaving second room...")
		if err := leaveRoom(peter, roomID2); err != nil {
			t.Fatalf("❌ Second leave failed: %v", err)
		}

		testLogger.Println("========== ✅ REJOIN TEST PASSED ==========")
		testLogger.Printf("Room 1: %s, Room 2: %s", roomID1, roomID2)
	})
}

func checkServicesRunning(t *testing.T) bool {
	services := map[string]string{
		"User": userServiceURL + "/health",
		"Room": roomServiceURL + "/health",
		"Game": gameServiceURL + "/health",
	}

	client := &http.Client{Timeout: 2 * time.Second}
	for name, url := range services {
		resp, err := client.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Logf("❌ %s Service not running", name)
			return false
		}
		resp.Body.Close()
	}
	return true
}

func registerAndLogin(t *testing.T, username, password string) *User {
	user := &User{Username: username, Password: password}

	// Try login
	loginData, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	resp, err := http.Post(userServiceURL+"/login", "application/json", bytes.NewBuffer(loginData))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result struct {
			UserID string `json:"user_id"`
			Token  string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		user.UserID = result.UserID
		user.Token = result.Token
		return user
	}

	// Register if not exists
	registerData, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	resp, err = http.Post(userServiceURL+"/register", "application/json", bytes.NewBuffer(registerData))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		ID     string `json:"id"`
		UserID string `json:"user_id"`
		Token  string `json:"token"`
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	if result.UserID == "" {
		user.UserID = result.ID
	} else {
		user.UserID = result.UserID
	}
	user.Token = result.Token

	return user
}

func joinMatchmaking(user *User) (string, error) {
	data, _ := json.Marshal(map[string]string{"user_id": user.UserID})
	req, _ := http.NewRequest("POST", roomServiceURL+"/join", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("join failed: %s", body)
	}

	var result struct {
		RoomID string `json:"room_id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.RoomID, nil
}

func waitForRoomFull(roomID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, _ := http.Get(fmt.Sprintf("%s/room/%s/ready", roomServiceURL, roomID))
		if resp != nil && resp.StatusCode == http.StatusOK {
			var result struct {
				Ready bool `json:"ready"`
			}
			json.NewDecoder(resp.Body).Decode(&result)
			resp.Body.Close()
			if result.Ready {
				return nil
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout")
}

func waitForGameReady(roomID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, _ := http.Get(fmt.Sprintf("%s/game/status?room_id=%s", gameServiceURL, roomID))
		if resp != nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout")
}

// Play game with random answers to test game mechanics (not trying to win)
func playGameWithRandomAnswers(user *User, roomID string) (string, error) {
	wsURL := fmt.Sprintf("ws://localhost:8003/game/ws?room_id=%s&user_id=%s", roomID, user.UserID)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	testLogger.Printf("✅ %s: Connected to game", user.Username)

	// Read game start
	var startMsg GameMessage
	if err := conn.ReadJSON(&startMsg); err != nil {
		return "", err
	}
	testLogger.Printf("🎮 %s: Role = %s", user.Username, startMsg.YourColor)

	colors := []string{"red", "blue", "yellow", "green"}
	roundsWon := 0

	// Play 5 rounds with random answers
	for round := 1; round <= 5; round++ {
		var msg GameMessage
		if err := conn.ReadJSON(&msg); err != nil {
			return "", err
		}

		if msg.Type == "round_start" {
			testLogger.Printf("📝 %s: Round %d - Word='%s' Color='%s'",
				user.Username, round, msg.Word, msg.Color)

			// Pick random answer (testing mechanics, not correctness)
			randomAnswer := colors[rand.Intn(len(colors))]

			clickMsg := map[string]string{
				"type":  "click",
				"color": randomAnswer,
			}
			conn.WriteJSON(clickMsg)

			// Read result
			var result GameMessage
			conn.ReadJSON(&result)

			if result.Type == "round_win" {
				roundsWon++
				testLogger.Printf("   ✅ %s: Won round %d!", user.Username, round)
			} else {
				testLogger.Printf("   ❌ %s: Lost round %d", user.Username, round)
			}
		}
	}

	// Read game over
	var gameOver GameMessage
	conn.ReadJSON(&gameOver)

	testLogger.Printf("🏁 %s: Game over - Winner: %s (won %d/5 rounds)",
		user.Username, gameOver.Winner, roundsWon)

	return gameOver.Winner, nil
}

func leaveRoom(user *User, roomID string) error {
	req, _ := http.NewRequest("POST",
		fmt.Sprintf("%s/rooms/%s/leave", roomServiceURL, roomID),
		bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Authorization", "Bearer "+user.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("leave failed: %s", body)
	}
	return nil
}
