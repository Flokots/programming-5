// Type definitions matching backend contracts

export type GameFlowState =
  | 'LOGIN'
  | 'JOINING'
  | 'WAITING_ROOM'
  | 'CONNECTING'
  | 'PLAYING'
  | 'GAME_OVER'
  | 'ERROR';

export type ColorOption = 'red' | 'blue' | 'green' | 'yellow';

// Auth Types
export interface AuthResponse {
  user_id: string;
  token: string;
  username: string;
}

// Game Types
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

export interface PlayerStats {
  wins: number;
  total_latency: number;
  avg_latency: number;
}

// WebSocket Message Types
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

export interface GameOverPayload {
  reason: string;
  winner: string;
  stats: { [userId: string]: PlayerStats };
}

export interface ClickMessage {
  type: 'CLICK';
  payload: {
    answer: ColorOption;
  };
}