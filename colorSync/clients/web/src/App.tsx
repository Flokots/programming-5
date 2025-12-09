import { useState, useEffect, useCallback, useRef } from 'react';
import { GameWebSocket } from './utils/websocket';
import { APIClient } from './utils/api';

import type {
  GameFlowState,
  GameState,
  ColorOption,
  GameStartPayload,
  RoundStartPayload,
  RoundResultPayload,
  GameOverPayload,
} from './types';
import './App.css';

function App() {
  // ============================================
  // STATE MANAGEMENT
  // ============================================

  const [flowState, setFlowState] = useState<GameFlowState>('LOGIN');
  const [error, setError] = useState<string | null>(null);

  // Auth state
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [userId, setUserId] = useState('');
  const [displayUsername, setDisplayUsername] = useState('');

  // Room state
  const [roomId, setRoomId] = useState('');

  // Game state
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

  // Refs for cleanup
  const apiRef = useRef(new APIClient());
  const wsRef = useRef(new GameWebSocket());
  const pollIntervalRef = useRef<number | null>(null);

  // ============================================
  // PHASE 1: AUTHENTICATION
  // ============================================

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    try {
      // Try login first
      try {
        const data = await apiRef.current.login(username, password);
        setUserId(data.user_id);
        setDisplayUsername(data.username);
        setFlowState('JOINING');
      } catch {
        // If login fails, try registration
        console.log('Login failed, attempting registration...');
        const data = await apiRef.current.register(username, password);
        setUserId(data.user_id);
        setDisplayUsername(data.username);
        setFlowState('JOINING');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Authentication failed');
    }
  };

  // ============================================
  // PHASE 2: AUTO-ADVANCE - JOINING → WAITING_ROOM
  // ============================================

  useEffect(() => {
    if (flowState === 'JOINING' && userId) {
      const joinMatchmaking = async () => {
        try {
          console.log('🎮 Attempting to join matchmaking...');
          console.log('   User ID:', userId);
          console.log('   Is Authenticated:', apiRef.current.isAuthenticated());
          
          const data = await apiRef.current.joinMatchmaking(userId);
          
          console.log('✅ Matchmaking join successful!');
          console.log('   Room ID:', data.room_id);
          
          setRoomId(data.room_id);
          setFlowState('WAITING_ROOM');
        } catch (err) {
          console.error('❌ Matchmaking join failed:', err);
          setError(err instanceof Error ? err.message : 'Failed to join matchmaking');
          setFlowState('ERROR');
        }
      };

      joinMatchmaking();
    }
  }, [flowState, userId]);

  // ============================================
  // PHASE 3: AUTO-ADVANCE - WAITING_ROOM → CONNECTING
  // ============================================

  useEffect(() => {
    if (flowState === 'WAITING_ROOM' && roomId) {
      let attempts = 0;
      const maxAttempts = 60;

      const checkRoomReady = async () => {
        try {
          const ready = await apiRef.current.checkRoomReady(roomId);
          
          if (ready) {
            if (pollIntervalRef.current) {
              clearInterval(pollIntervalRef.current);
              pollIntervalRef.current = null;
            }
            console.log('✅ Opponent found! Room is full.');
            
            // Skip WAITING_GAME and go straight to CONNECTING
            // The room service already notified the game service
            console.log('🔌 Room ready, connecting to game WebSocket...');
            setFlowState('CONNECTING');
            return;
          }

          attempts++;
          if (attempts >= maxAttempts) {
            if (pollIntervalRef.current) {
              clearInterval(pollIntervalRef.current);
              pollIntervalRef.current = null;
            }
            setError('Matchmaking timeout - no opponent found');
            setFlowState('ERROR');
          }
        } catch (err) {
          console.error('Error checking room status:', err);
        }
      };

      checkRoomReady();
      pollIntervalRef.current = window.setInterval(checkRoomReady, 1000);

      return () => {
        if (pollIntervalRef.current) {
          clearInterval(pollIntervalRef.current);
          pollIntervalRef.current = null;
        }
      };
    }
  }, [flowState, roomId]);

  // ============================================
  // PHASE 5: WEBSOCKET CONNECTION
  // ============================================

  useEffect(() => {
    if (flowState === 'CONNECTING' && roomId && userId) {
      const connectWebSocket = async () => {
        try {
          await wsRef.current.connect(roomId, userId);
          setFlowState('PLAYING');
        } catch (err) {
          setError(err instanceof Error ? err.message : 'Failed to connect to game');
          setFlowState('ERROR');
        }
      };

      connectWebSocket();
    }
  }, [flowState, roomId, userId]);

  // ============================================
  // PHASE 6: WEBSOCKET MESSAGE HANDLERS
  // ============================================

  useEffect(() => {
    if (flowState === 'PLAYING') {
      const ws = wsRef.current;

      ws.on<GameStartPayload>('GAME_START', (payload) => {
        console.log('🎮 Game starting!');
        setGameState((prev: GameState) => ({
          ...prev,
          maxRounds: payload.max_rounds,
          currentRound: 0,
        }));
      });

      ws.on<RoundStartPayload>('ROUND_START', (payload) => {
        console.log(`🎨 Round ${payload.round}: Word='${payload.word}' Color='${payload.color}'`);
        setGameState((prev: GameState) => ({
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
        console.log(`📊 Round ${payload.round} result: Winner=${payload.winner}`);
        
        setGameState((prev: GameState) => {
          const isYourWin = payload.winner === userId;
          return {
            ...prev,
            yourScore: isYourWin ? prev.yourScore + 1 : prev.yourScore,
            opponentScore: !isYourWin && payload.winner !== 'timeout' 
              ? prev.opponentScore + 1 
              : prev.opponentScore,
            yourLatency: isYourWin ? prev.yourLatency + payload.latency_ms : prev.yourLatency,
            opponentLatency: !isYourWin && payload.winner !== 'timeout'
              ? prev.opponentLatency + payload.latency_ms
              : prev.opponentLatency,
            roundResults: [
              ...prev.roundResults,
              {
                round: payload.round,
                winner: payload.winner,
                latency: payload.latency_ms,
              },
            ],
          };
        });
      });

      ws.on<GameOverPayload>('GAME_OVER', (payload) => {
        console.log('🏁 Game over! Winner:', payload.winner);
        
        setGameState((prev: GameState) => ({
          ...prev,
          winner: payload.winner,
          yourStats: payload.stats[userId] || null,
          opponentStats: Object.entries(payload.stats).find(
            ([id]) => id !== userId
          )?.[1] || null,
        }));
        
        setFlowState('GAME_OVER');
      });

      ws.on('WRONG_ANSWER', () => {
        console.log('❌ Wrong answer!');
        setShowWrongAnswer(true);
        setTimeout(() => setShowWrongAnswer(false), 1000);
      });

      ws.on<{ message?: string }>('ERROR', (payload) => {
        console.error('Game error:', payload);
        setError(payload.message || 'Game error occurred');
      });

      return () => {
        ws.off('GAME_START');
        ws.off('ROUND_START');
        ws.off('ROUND_RESULT');
        ws.off('GAME_OVER');
        ws.off('WRONG_ANSWER');
        ws.off('ERROR');
      };
    }
  }, [flowState, userId]);

  // ============================================
  // GAME ACTIONS
  // ============================================

  const handleColorClick = useCallback((color: ColorOption) => {
    if (answered || flowState !== 'PLAYING') return;

    setAnswered(true);
    wsRef.current.sendClick(color);
  }, [answered, flowState]);

  // Keyboard shortcuts
  useEffect(() => {
    if (flowState !== 'PLAYING' || answered) return;

    const handleKeyPress = (e: KeyboardEvent) => {
      const keyMap: { [key: string]: ColorOption } = {
        'r': 'red',
        'b': 'blue',
        'g': 'green',
        'y': 'yellow',
      };

      const color = keyMap[e.key.toLowerCase()];
      if (color) {
        handleColorClick(color);
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [flowState, answered, handleColorClick]);

  // ============================================
  // CLEANUP ON UNMOUNT
  // ============================================

  useEffect(() => {
    const ws = wsRef.current;
    const api = apiRef.current;
    const currentRoomId = roomId;

    return () => {
      ws.disconnect();
      
      if (currentRoomId) {
        api.leaveRoom(currentRoomId);
      }
      
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current);
      }
    };
  }, [roomId]);

  // ============================================
  // RENDER FUNCTIONS
  // ============================================

  const renderLogin = () => (
    <div className="screen login-screen">
      <div className="header">
        <h1>🎨 COLOR SYNC GAME 🎨</h1>
        <p className="subtitle">Real-time Stroop Test Game</p>
      </div>
      
      <form onSubmit={handleLogin} className="login-form">
        <div className="form-group">
          <label htmlFor="username">Username:</label>
          <input
            id="username"
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Enter username"
            required
            autoFocus
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="password">Password:</label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Enter password"
            required
          />
        </div>
        
        {error && <div className="error-message">{error}</div>}
        
        <button type="submit" className="btn-primary">
          Login / Register
        </button>
        
        <p className="hint">
          New user? We'll automatically create your account!
        </p>
      </form>
    </div>
  );

  const renderWaiting = () => (
    <div className="screen waiting-screen">
      <div className="header">
        <h1>🎨 COLOR SYNC GAME 🎨</h1>
        <p className="username">Playing as: <strong>{displayUsername}</strong></p>
      </div>
      
      <div className="waiting-content">
        <div className="spinner"></div>
        
        {flowState === 'JOINING' && <p className="status">Joining matchmaking...</p>}
        {flowState === 'WAITING_ROOM' && (
          <>
            <p className="status">Waiting for opponent...</p>
            <p className="hint">Room ID: {roomId}</p>
          </>
        )}
        {flowState === 'WAITING_GAME' && <p className="status">Preparing game...</p>}
        {flowState === 'CONNECTING' && <p className="status">Connecting to game server...</p>}
      </div>
    </div>
  );

  const renderGame = () => (
    <div className="screen game-screen">
      <div className="game-header">
        <div className="player-info">
          <span className="username">{displayUsername}</span>
          <span className="score">Score: {gameState.yourScore}</span>
        </div>
        <div className="round-info">
          Round {gameState.currentRound} / {gameState.maxRounds}
        </div>
        <div className="opponent-info">
          <span className="score">Score: {gameState.opponentScore}</span>
          <span className="username">Opponent</span>
        </div>
      </div>

      {gameState.currentRound === 0 ? (
        <div className="game-instructions">
          <h2>🎮 GAME STARTING!</h2>
          <p>Match the <strong>COLOR</strong> of the text, not the word!</p>
          <p>You will play {gameState.maxRounds} rounds</p>
          <div className="controls-hint">
            <p><strong>Controls:</strong></p>
            <p>Click buttons or use keyboard: <kbd>R</kbd> <kbd>B</kbd> <kbd>G</kbd> <kbd>Y</kbd></p>
          </div>
        </div>
      ) : (
        <>
          <div className="word-display">
            <div
              className="word"
              style={{ color: gameState.color }}
            >
              {gameState.word}
            </div>
            <p className="prompt">What COLOR is this text?</p>
          </div>

          <div className="color-buttons">
            <button
              className="color-btn red"
              onClick={() => handleColorClick('red')}
              disabled={answered}
            >
              <span className="key">R</span>
              RED
            </button>
            <button
              className="color-btn blue"
              onClick={() => handleColorClick('blue')}
              disabled={answered}
            >
              <span className="key">B</span>
              BLUE
            </button>
            <button
              className="color-btn green"
              onClick={() => handleColorClick('green')}
              disabled={answered}
            >
              <span className="key">G</span>
              GREEN
            </button>
            <button
              className="color-btn yellow"
              onClick={() => handleColorClick('yellow')}
              disabled={answered}
            >
              <span className="key">Y</span>
              YELLOW
            </button>
          </div>

          {showWrongAnswer && (
            <div className="feedback wrong">
              ❌ Wrong answer! Try again!
            </div>
          )}

          {answered && !showWrongAnswer && (
            <div className="feedback correct">
              ✅ Answer submitted!
            </div>
          )}
        </>
      )}
    </div>
  );

  const renderGameOver = () => {
    const isWinner = gameState.winner === userId;
    const isDraw = gameState.winner === 'draw';

    return (
      <div className="screen game-over-screen">
        <div className="header">
          <h1>🏁 GAME OVER</h1>
        </div>

        <div className="result-banner">
          {isWinner && <div className="banner winner">🎉 YOU WON! 🎉</div>}
          {!isWinner && !isDraw && <div className="banner loser">😞 YOU LOST 😞</div>}
          {isDraw && <div className="banner draw">🤝 IT'S A DRAW! 🤝</div>}
        </div>

        <div className="stats-container">
          <div className="stats-panel your-stats">
            <h3>Your Stats</h3>
            <div className="stat-row">
              <span>Rounds Won:</span>
              <span className="value">{gameState.yourScore}</span>
            </div>
            <div className="stat-row">
              <span>Rounds Lost:</span>
              <span className="value">{gameState.maxRounds - gameState.yourScore}</span>
            </div>
            <div className="stat-row">
              <span>Total Latency:</span>
              <span className="value">{gameState.yourLatency}ms</span>
            </div>
            {gameState.yourScore > 0 && (
              <div className="stat-row">
                <span>Avg Latency:</span>
                <span className="value">
                  {Math.round(gameState.yourLatency / gameState.yourScore)}ms
                </span>
              </div>
            )}
          </div>

          <div className="stats-panel opponent-stats">
            <h3>Opponent Stats</h3>
            <div className="stat-row">
              <span>Rounds Won:</span>
              <span className="value">{gameState.opponentScore}</span>
            </div>
            <div className="stat-row">
              <span>Rounds Lost:</span>
              <span className="value">{gameState.maxRounds - gameState.opponentScore}</span>
            </div>
            <div className="stat-row">
              <span>Total Latency:</span>
              <span className="value">{gameState.opponentLatency}ms</span>
            </div>
            {gameState.opponentScore > 0 && (
              <div className="stat-row">
                <span>Avg Latency:</span>
                <span className="value">
                  {Math.round(gameState.opponentLatency / gameState.opponentScore)}ms
                </span>
              </div>
            )}
          </div>
        </div>

        <div className="round-history">
          <h3>Round History</h3>
          {gameState.roundResults.map((result: { round: number; winner: string; latency: number }) => (
            <div key={result.round} className="history-row">
              <span>Round {result.round}:</span>
              <span className={result.winner === userId ? 'win' : 'loss'}>
                {result.winner === userId ? '✅ You won' : '❌ Opponent won'}
              </span>
              <span className="latency">{result.latency}ms</span>
            </div>
          ))}
        </div>

        <button
          className="btn-primary"
          onClick={() => window.location.reload()}
        >
          🎮 Play Again
        </button>
      </div>
    );
  };

  const renderError = () => (
    <div className="screen error-screen">
      <div className="header">
        <h1>❌ Error</h1>
      </div>
      <div className="error-content">
        <p>{error || 'An unexpected error occurred'}</p>
        <button
          className="btn-primary"
          onClick={() => window.location.reload()}
        >
          Try Again
        </button>
      </div>
    </div>
  );

  return (
    <div className="app">
      {flowState === 'LOGIN' && renderLogin()}
      {['JOINING', 'WAITING_ROOM', 'WAITING_GAME', 'CONNECTING'].includes(flowState) && renderWaiting()}
      {flowState === 'PLAYING' && renderGame()}
      {flowState === 'GAME_OVER' && renderGameOver()}
      {flowState === 'ERROR' && renderError()}
    </div>
  );
}

export default App;
