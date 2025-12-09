// Type definitions matching backend contracts

// ============================================
// AUTH TYPES (User Service)
// ============================================
export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
}

export interface AuthResponse {
  user_id: string;
  token: string;
  username: string;
}

// ============================================
// ROOM TYPES (Room Service)
// ============================================
export interface JoinRequest {
  user_id: string;
}

export interface JoinResponse {
  room_id: string;
  status: string;
}

export interface RoomReadyResponse {
  ready: boolean;
  player_count: number;
}

// ============================================
// GAME TYPES (Game Rules Service)
// ============================================
export interface GameStatusResponse {
  ready: boolean;
}

// ============================================
// WEBSOCKET MESSAGE TYPES
// ============================================
export type WSMessageType =
  | 'GAME_START'
  | 'ROUND_START'
  | 'ROUND_RESULT'
  | 'GAME_OVER'
  | 'WRONG_ANSWER'
  | 'ERROR';

export interface WSMessage {
  type: WSMessageType;
  payload: unknown;
}

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
  winner: string;
  latency_ms: number;
}

export interface PlayerStats {
  wins: number;
  total_latency: number;
}

export interface GameOverPayload {
  reason: string;
  winner: string;
  stats: { [userId: string]: PlayerStats };
}

export interface ClickMessage {
  type: 'CLICK';
  payload: {
    answer: 'red' | 'blue' | 'green' | 'yellow';
  };
}

// ============================================
// UI STATE TYPES
// ============================================
export type GameFlowState =
  | 'LOGIN'
  | 'JOINING'
  | 'WAITING_ROOM'
  | 'CONNECTING'
  | 'PLAYING'
  | 'GAME_OVER'
  | 'ERROR';

export type ColorOption = 'red' | 'blue' | 'green' | 'yellow';

export interface GameState {
  currentRound: number;
  maxRounds: number;
  word: string;
  color: string;
  roundStartTime: number | null;
  yourScore: number;
  opponentScore: number;
  yourLatency: number;
  opponentLatency: number;
  roundResults: Array<{
    round: number;
    winner: string;
    latency: number;
  }>;
  winner: string | null;
  yourStats: PlayerStats | null;
  opponentStats: PlayerStats | null;
}