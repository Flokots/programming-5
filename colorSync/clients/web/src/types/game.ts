// Web Socket message types (from backend)
export type MessageType =
    | "GAME_START"
    | "ROUND_START"
    | "ROUND_RESULT"
    | "GAME_OVER"
    | "ERROR"
    | "WRONG_ANSWER"
    | "CLICK"; // Client sends click message

export interface WSMessage {
    type: MessageType;
    payload: unknown;
}

// Specify Message Payloads
export interface GameStartPayload {
    max_rounds: number;
}

export interface RoundStartPayload {
    round: number;
    word: string;
    color: string;
}

export interface RoundResultPayload {
    round: number;
    winner: string | null; // user-id, "timeout" or null
    latency_ms?: number; // optional - only if there's a winner
}

export interface PlayerStats {
    wins: number;
    total_latency: number;
    avg_latency: number;
}

export interface GameOverPayload {
    winner: string; // user-id or "draw"
    reason: "game_completed" | "opponent_disconnected ";
    stats: Record<string, PlayerStats>; // user-id mapped to stats
}

export interface ErrorPayload {
    message: string;
}

export interface ClickPayload {
    answer: string; // "red" | "blue" | "green" | "yellow"
}

// Game State (for React components)
export type GameStatus =
    | "connecting"
    | "waiting"
    | "playing"
    | "game_over"
    | "error";

export interface GameState {
    status: GameStatus;

    // current round info
    currentRound: number;
    maxRounds: number;
    word: string;
    color: string;

    // Scores
    myScore: number;
    opponentScore: number;

    // Game over data
    winner: string | null;
    myStats: PlayerStats | null;

    // Error handling
    errorMessage: string | null;

    // User info
    userId: string;
    roomId: string;
}

// Color Types (for Stroop Test)
export type StroopColor = "red" | "blue" | "green" | "yellow";

export const COLOR_MAP: Record<StroopColor, string> = {
    "red": "#FF0000",
    "blue": "#0000FF",
    "green": "#00FF00",
    "yellow": "#FFD700",
};

// API Response Types (from user-service and room-service)
export interface LoginResponse {
    user_id: string;
}

export interface RegisterResponse {
    user_id: string;
}

export interface JoinRoomResponse {
    room_id: string;
    player1_id: string;
    player2_id: string | null;
}

export interface GameReadyResponse {
    ready: boolean;
    player1_id: string;
    player2_id: string;
}