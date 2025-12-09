import { useState, useEffect, useCallback, useRef } from 'react';
import { StatusBar } from 'expo-status-bar';
import { StyleSheet, View } from 'react-native';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import { APIClient } from './src/utils/api';
import { GameWebSocket } from './src/utils/websocket';
import LoginScreen from './src/screens/LoginScreen';
import WaitingScreen from './src/screens/WaitingScreen';
import GameScreen from './src/screens/GameScreen';
import GameOverScreen from './src/screens/GameOverScreen';

import type {
  GameFlowState,
  GameState,
  GameStartPayload,
  RoundStartPayload,
  RoundResultPayload,
  GameOverPayload,
} from './src/types';

export default function App() {
  const [flowState, setFlowState] = useState<GameFlowState>('LOGIN');
  const [error, setError] = useState<string | null>(null);

  const [userId, setUserId] = useState('');
  const [displayUsername, setDisplayUsername] = useState('');
  const [roomId, setRoomId] = useState('');

  const [gameState, setGameState] = useState<GameState>({
    currentRound: 0,
    maxRounds: 5,
    word: '',
    color: '',
    roundStartTime: null,
    yourScore: 0,
    opponentScore: 0,
    yourLatency: 0,
    opponentLatency: 0,
    roundResults: [],
    winner: null,
    yourStats: null,
    opponentStats: null,
  });

  const [answered, setAnswered] = useState(false);
  const [showWrongAnswer, setShowWrongAnswer] = useState(false);

  const apiRef = useRef(new APIClient());
  const wsRef = useRef(new GameWebSocket());
  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null);

  // Authentication
  const handleLogin = useCallback(async (username: string, password: string) => {
    setError(null);
    try {
      try {
        const data = await apiRef.current.login(username, password);
        setUserId(data.user_id);
        setDisplayUsername(data.username);
        setFlowState('JOINING');
      } catch {
        const data = await apiRef.current.register(username, password);
        setUserId(data.user_id);
        setDisplayUsername(data.username);
        setFlowState('JOINING');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Authentication failed');
    }
  }, []);

  // Auto-join matchmaking
  useEffect(() => {
    if (flowState === 'JOINING' && userId) {
      const joinMatchmaking = async () => {
        try {
          const data = await apiRef.current.joinMatchmaking(userId);
          setRoomId(data.room_id);
          setFlowState('WAITING_ROOM');
        } catch (err) {
          setError(err instanceof Error ? err.message : 'Failed to join matchmaking');
          setFlowState('ERROR');
        }
      };
      joinMatchmaking();
    }
  }, [flowState, userId]);

  // Poll for opponent
  useEffect(() => {
    if (flowState === 'WAITING_ROOM' && roomId) {
      let attempts = 0;
      const checkRoomReady = async () => {
        const ready = await apiRef.current.checkRoomReady(roomId);
        if (ready) {
          if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
          setFlowState('CONNECTING');
        } else if (++attempts >= 60) {
          if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
          setError('Matchmaking timeout');
          setFlowState('ERROR');
        }
      };

      checkRoomReady();
      pollIntervalRef.current = setInterval(checkRoomReady, 1000);

      return () => {
        if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
      };
    }
  }, [flowState, roomId]);

  // Connect WebSocket
  useEffect(() => {
    if (flowState === 'CONNECTING' && roomId && userId) {
      wsRef.current.connect(roomId, userId)
        .then(() => setFlowState('PLAYING'))
        .catch(() => setFlowState('ERROR'));
    }
  }, [flowState, roomId, userId]);

  // WebSocket handlers
  useEffect(() => {
    if (flowState === 'PLAYING') {
      const ws = wsRef.current;

      ws.on<GameStartPayload>('GAME_START', (payload) => {
        setGameState(prev => ({ ...prev, maxRounds: payload.max_rounds, currentRound: 0 }));
      });

      ws.on<RoundStartPayload>('ROUND_START', (payload) => {
        setGameState(prev => ({
          ...prev,
          currentRound: payload.round,
          word: payload.word,
          color: payload.color,
          roundStartTime: Date.now(),
        }));
        setAnswered(false);
        setShowWrongAnswer(false);
      });

      ws.on<RoundResultPayload>('ROUND_RESULT', (payload) => {
        setGameState(prev => {
          const isYourWin = payload.winner === userId;
          return {
            ...prev,
            yourScore: isYourWin ? prev.yourScore + 1 : prev.yourScore,
            opponentScore: !isYourWin && payload.winner !== 'timeout' ? prev.opponentScore + 1 : prev.opponentScore,
            yourLatency: isYourWin ? prev.yourLatency + payload.latency_ms : prev.yourLatency,
            opponentLatency: !isYourWin && payload.winner !== 'timeout' ? prev.opponentLatency + payload.latency_ms : prev.opponentLatency,
            roundResults: [...prev.roundResults, { round: payload.round, winner: payload.winner, latency: payload.latency_ms }],
          };
        });
      });

      ws.on<GameOverPayload>('GAME_OVER', async (payload) => {
        setGameState(prev => ({
          ...prev,
          winner: payload.winner,
          yourStats: payload.stats[userId] || null,
          opponentStats: Object.entries(payload.stats).find(([id]) => id !== userId)?.[1] || null,
        }));
        setFlowState('GAME_OVER');
        await apiRef.current.leaveRoom(roomId);
      });

      ws.on('WRONG_ANSWER', () => {
        setShowWrongAnswer(true);
        setTimeout(() => setShowWrongAnswer(false), 1000);
      });

      return () => {
        ws.off('GAME_START');
        ws.off('ROUND_START');
        ws.off('ROUND_RESULT');
        ws.off('GAME_OVER');
        ws.off('WRONG_ANSWER');
      };
    }
  }, [flowState, userId, roomId]);

  const handleColorClick = useCallback((color: 'red' | 'blue' | 'green' | 'yellow') => {
    if (!answered && flowState === 'PLAYING') {
      setAnswered(true);
      wsRef.current.sendClick(color);
    }
  }, [answered, flowState]);

  const handlePlayAgain = useCallback(() => {
    setFlowState('JOINING');
    setRoomId('');
    setGameState({
      currentRound: 0,
      maxRounds: 5,
      word: '',
      color: '',
      roundStartTime: null,
      yourScore: 0,
      opponentScore: 0,
      yourLatency: 0,
      opponentLatency: 0,
      roundResults: [],
      winner: null,
      yourStats: null,
      opponentStats: null,
    });
    wsRef.current.disconnect();
  }, []);

  return (
    <SafeAreaProvider>
      <View style={styles.container}>
        <StatusBar style="auto" />
        {flowState === 'LOGIN' && <LoginScreen onLogin={handleLogin} error={error} />}
        {['JOINING', 'WAITING_ROOM', 'CONNECTING'].includes(flowState) && (
          <WaitingScreen flowState={flowState} username={displayUsername} roomId={roomId} />
        )}
        {flowState === 'PLAYING' && (
          <GameScreen
            gameState={gameState}
            username={displayUsername}
            answered={answered}
            showWrongAnswer={showWrongAnswer}
            onColorClick={handleColorClick}
          />
        )}
        {flowState === 'GAME_OVER' && (
          <GameOverScreen
            gameState={gameState}
            userId={userId}
            onPlayAgain={handlePlayAgain}
          />
        )}
      </View>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1a1a2e',
  },
});