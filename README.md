# 🎨 ColorSync - Distributed Stroop Effect Game

A real-time multiplayer Stroop effect reflex game demonstrating microservices architecture, WebSocket communication, and cross-platform client development.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENTS                                 │
│  ┌──────────┐      ┌──────────┐      ┌──────────────┐           │
│  │   CLI    │      │   Web    │      │   Mobile     │           │
│  │  (Go)    │      │ (React + │      │ (React       │           │
│  │          │      │   Vite)  │      │  Native)     │           │
│  │          │      │          │      │              │           │
│  └────┬─────┘      └────┬─────┘      └──────┬───────┘           │
│       │                 │                    │                  │
└───────┼─────────────────┼────────────────────┼──────────────────┘
        │                 │                    │
        │    HTTP/REST    │                    │
        │    + JWT Auth   │                    │
        │                 │                    │
┌───────▼─────────────────▼────────────────────▼──────────────────┐
│                    BACKEND SERVICES                             │
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  User Service   │  │  Room Service   │  │ GameRulesService│  │
│  │    :8001        │  │     :8002       │  │     :8003       │  │
│  │                 │  │                 │  │                 │  │ 
│  │ • Registration  │  │ • Matchmaking   │  │ • Game Logic    │  │
│  │ • Login         │◀─┤ • Room Mgmt     │◀─┤ • WebSocket     │. │
│  │ • JWT Auth      │  │ • Player Verify │  │ • Round Mgmt    │. │
│  │ • User Data     │  │                 │  │ • Score Track   │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘. │
│         │                      │                     │          │
│         │  Service-to-Service  │  Zero Trust Auth    │          │
│         │  HTTP + JWT Tokens   │  (Service Tokens)   │          │
│         └──────────────────────┴─────────────────────┘          │
└─────────────────────────────────────────────────────────────────┘

Communication Protocols:
─────────────────────────
• Client ↔ User Service:    HTTP REST + JWT
• Client ↔ Room Service:    HTTP REST + JWT
• Client ↔ Game Service:    WebSocket (Real-time)
• Service ↔ Service:        HTTP + Service Tokens (Zero Trust)
```

## Technology Stack

### Backend Services (Go 1.21+)

| Service | Technology | Port | Purpose |
|---------|-----------|------|---------|
| **User Service** | Go + `net/http` | 8001 | Authentication & user management |
| **Room Service** | Go + `net/http` | 8002 | Matchmaking & room coordination |
| **Game Rules Service** | Go + `gorilla/websocket` | 8003 | Game logic & real-time communication |
| **Shared Library** | Go modules | - | JWT auth middleware (Zero Trust) |

**Key Libraries:**
- `github.com/golang-jwt/jwt/v5` - JWT token generation/validation
- `github.com/gorilla/websocket` - WebSocket protocol
- `github.com/google/uuid` - Unique ID generation

### Clients

| Client | Technology | Purpose |
|--------|-----------|---------|
| **CLI** | Go + `fmt`/`bufio` | Simple terminal-based client |
| **Web** | React 18 + TypeScript + Vite | Browser-based SPA |
| **Mobile** | React Native + Expo | iOS/Android native app |

**Shared Client Stack:**
- TypeScript for type safety
- WebSocket API for real-time game
- REST API for authentication/matchmaking

---

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Node.js 18+ and npm (for web/mobile clients)
- Expo Go app (for mobile testing)

### 1. Start Backend Services

**Terminal 1 - User Service:**
```bash
cd backend/user-service
go run main.go
# Listens on http://localhost:8001
```

**Terminal 2 - Room Service:**
```bash
cd backend/room-service
go run main.go
# Listens on http://localhost:8002
```

**Terminal 3 - Game Rules Service:**
```bash
cd backend/game-rules-service
go run main.go
# Listens on http://localhost:8003
```

### 2. Start Clients

**CLI Client:**
```bash
cd clients/cli
go run main.go
```

**Web Client:**
```bash
cd clients/web
npm install
npm run dev
# Open http://localhost:5173
```

**Mobile Client:**
```bash
cd clients/mobile
npm install
npx expo start --web
# Scan QR code with Expo Go app
```

---

## Game Rules

**ColorSync** is a two-player Stroop effect test:

1. **Matchmaking:** Players join a queue and are automatically paired
2. **Objective:** Identify the **COLOR** of the text (not the word itself)
3. **Rounds:** 5 rounds per game
4. **Scoring:** 
   - Fastest correct answer wins the round
   - Wrong answers lock you out for that round
   - Timeout after 5 seconds = no winner
5. **Winner:** Player with most round wins

**Example Round:**
```
Word displayed: "BLUE"  (in yellow color)
Correct answer: Yellow  ✅
Wrong answer: Blue      ❌ (locked out for this round)
```

---

##  API Documentation

### Service-to-Service Communication

#### **User Service API** (Port 8001)

**Public Endpoints:**
```http
POST /register
Content-Type: application/json

Request:
{
  "username": "alice",
  "password": "securepass123"
}

Response: 200 OK
{
  "id": "96e698fc-2640-4300-8086-04f6ad26985c",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "alice"
}
```

```http
POST /login
Content-Type: application/json

Request:
{
  "username": "alice",
  "password": "securepass123"
}

Response: 200 OK
{
  "id": "96e698fc-2640-4300-8086-04f6ad26985c",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "alice"
}
```

**Internal Endpoints (Service-to-Service):**
```http
GET /users/{user_id}
Authorization: Bearer <JWT_TOKEN>

Response: 200 OK
{
  "id": "96e698fc-2640-4300-8086-04f6ad26985c",
  "username": "alice"
}
```

---

#### **Room Service API** (Port 8002)

**Protected Endpoints (Require JWT):**
```http
POST /join
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

Request:
{
  "user_id": "96e698fc-2640-4300-8086-04f6ad26985c"
}

Response: 200 OK (New Room)
{
  "room_id": "bc8005f2-3a19-4015-b8e8-f24bab86d7ea",
  "players": ["96e698fc-2640-4300-8086-04f6ad26985c"],
  "status": "waiting",
  "message": "Waiting for opponent"
}

Response: 200 OK (Room Full)
{
  "room_id": "bc8005f2-3a19-4015-b8e8-f24bab86d7ea",
  "players": [
    "96e698fc-2640-4300-8086-04f6ad26985c",
    "2f889035-411a-42d5-aa9d-f1c5c65c00e2"
  ],
  "status": "full",
  "message": "Room is full! Game starting..."
}

Error: 409 Conflict
{
  "error": "You are already in an active room"
}
```

**Public Endpoints:**
```http
GET /room/{room_id}/ready

Response: 200 OK
{
  "ready": true,
  "player_count": 2
}
```

**Service-to-Service (Internal):**
```http
POST /game/start
X-Service-Token: <SERVICE_TOKEN>
Content-Type: application/json

Request (from Room Service to Game Service):
{
  "room_id": "bc8005f2-3a19-4015-b8e8-f24bab86d7ea",
  "players": [
    "96e698fc-2640-4300-8086-04f6ad26985c",
    "2f889035-411a-42d5-aa9d-f1c5c65c00e2"
  ]
}

Response: 200 OK
{
  "room_id": "bc8005f2-3a19-4015-b8e8-f24bab86d7ea",
  "message": "Game created",
  "status": "waiting_for_players"
}
```

---

#### **Game Rules Service API** (Port 8003)

**WebSocket Connection:**
```
ws://localhost:8003/game/ws?room_id={ROOM_ID}&user_id={USER_ID}
```

---

### Client-Server WebSocket Messages

#### **Server → Client Messages**

**1. GAME_START**
```json
{
  "type": "GAME_START",
  "payload": {
    "max_rounds": 5
  }
}
```
*Sent when both players connect and game begins.*

---

**2. ROUND_START**
```json
{
  "type": "ROUND_START",
  "payload": {
    "round": 1,
    "word": "BLUE",
    "color": "yellow"
  }
}
```
*Sent at the start of each round. `word` is the text displayed, `color` is the actual color of the text.*

---

**3. WRONG_ANSWER**
```json
{
  "type": "WRONG_ANSWER",
  "payload": {}
}
```
*Sent immediately when a player clicks the wrong color. Player is locked out for this round.*

---

**4. ROUND_RESULT**
```json
{
  "type": "ROUND_RESULT",
  "payload": {
    "round": 1,
    "winner": "96e698fc-2640-4300-8086-04f6ad26985c",
    "latency_ms": 1234
  }
}
```
*Sent after round ends (5 seconds or correct answer). `winner` can be user_id or `"timeout"`.*

---

**5. GAME_OVER**
```json
{
  "type": "GAME_OVER",
  "payload": {
    "reason": "completed",
    "winner": "96e698fc-2640-4300-8086-04f6ad26985c",
    "stats": {
      "96e698fc-2640-4300-8086-04f6ad26985c": {
        "wins": 3,
        "total_latency": 4567,
        "avg_latency": 1522
      },
      "2f889035-411a-42d5-aa9d-f1c5c65c00e2": {
        "wins": 2,
        "total_latency": 5890,
        "avg_latency": 2945
      }
    }
  }
}
```
*Sent when all rounds are complete. Includes final scores and statistics.*

---

**6. ERROR**
```json
{
  "type": "ERROR",
  "payload": {
    "message": "Connection lost"
  }
}
```
*Sent when an error occurs during the game.*

---

#### **Client → Server Messages**

**1. CLICK**
```json
{
  "type": "CLICK",
  "payload": {
    "answer": "yellow"
  }
}
```
*Sent when player clicks a color button. `answer` must be one of: `"red"`, `"blue"`, `"green"`, `"yellow"`.*

---

## 🏗️ Design Decisions & Rationale

### Why Microservices Architecture?

**Separation of Concerns:**
- Each service has a single, well-defined responsibility
- User management, matchmaking, and game logic are independent
- Services can be developed, tested, and deployed separately

**Scalability:**
- Game Rules Service can scale horizontally to handle more concurrent games
- User/Room services can scale independently based on load
- WebSocket connections isolated from authentication logic

**Fault Isolation:**
- If Game Service crashes, users can still register/login
- Room Service failure doesn't affect ongoing games
- Easier to debug and maintain

### Why Go for Backend?

**Performance:**
- Native concurrency with goroutines (perfect for WebSocket connections)
- Low latency for real-time game requirements
- Efficient memory management

**Simplicity:**
- Standard library includes robust `net/http` package
- Easy deployment (single binary, no runtime dependencies)
- Strong typing prevents runtime errors

**WebSocket Support:**
- Gorilla WebSocket library is production-grade
- Handles thousands of concurrent connections efficiently

### Why JWT for Authentication?

**Stateless:**
- No server-side session storage needed
- Each service can independently verify tokens
- Easy to scale horizontally

**Zero Trust:**
- Service-to-service calls use separate service tokens
- Room Service can't impersonate users
- Cryptographically signed and verifiable

**Cross-Platform:**
- Works identically in CLI, web, and mobile clients
- Standard Authorization header format

### Why WebSocket for Game Communication?

**Real-Time Requirements:**
- Sub-100ms latency critical for fair competition
- Stroop test requires instant feedback
- HTTP polling would introduce 500ms+ delay

**Bidirectional:**
- Server pushes round updates immediately
- Client sends clicks without waiting
- Full-duplex communication

**Connection Efficiency:**
- Single persistent connection vs. multiple HTTP requests
- Reduces overhead and server load
- Better for mobile devices (battery life)

### Why Multiple Client Types?

**Demonstrates Portability:**
- Same backend serves all platforms
- Proves API design is technology-agnostic
- Shows real-world distributed system design

**Different Use Cases:**
- CLI: Quick testing, automation, headless environments
- Web: Accessible anywhere, no installation
- Mobile: Native experience, touch optimized

---
## Testing

### Manual Testing

1. Start all three backend services
2. Open two client instances (any combination of CLI/Web/Mobile)
3. Register/login with different usernames
4. Join matchmaking on both clients
5. Play the game when paired

## Security Features

### Authentication
- Passwords never stored in plain text
- JWT tokens with 24-hour expiry
- Token refresh not implemented (students can add this)

### Authorization
- All Room/Game endpoints require valid JWT
- Service-to-service calls use separate service tokens (Zero Trust)
- WebSocket connections validate user_id matches JWT claims

### Input Validation
- Username/password length checks
- Color answer validation (only 4 valid colors)
- Room ID and User ID format validation

---

## Future Enhancements

- [ ] **Persistent Storage:** PostgreSQL for user data and game history
- [ ] **Leaderboard:** Track all-time wins and rankings
- [ ] **Tournaments:** Multi-round elimination brackets
- [ ] **Replay System:** Save and review past games
- [ ] **Docker Compose:** One-command deployment
- [ ] **Kubernetes:** Production-ready orchestration
- [ ] **Monitoring:** Prometheus + Grafana dashboards
- [ ] **Rate Limiting:** Prevent API abuse
- [ ] **Reconnection Logic:** Handle network interruptions gracefully

---

## 👨Development

### Adding a New Feature

1. **Backend:**
   - Add endpoint to appropriate service
   - Update shared types if needed
   - Add authentication if required

2. **Clients:**
   - Update `api.ts` with new endpoint
   - Update TypeScript types
   - Implement UI changes

3. **Testing:**
   - Add unit tests for new logic
   - Update E2E tests if flow changes

---

## 📄License

MIT License - Educational Project

---

## Acknowledgments

Built as part of Programming 5 course requirements demonstrating:
- Distributed systems design
- Real-time communication protocols
- Microservices architecture
- Cross-platform development
- RESTful and WebSocket APIs

---

**Project by:** Florence Kotohoyoh
**Course:** Programming 5  
**Date:** December 2025