import { useState, useEffect} from 'react';
import './App.css';
import { registerUser, loginUser, joinMatchmaking, isRoomReady } from './utils/api';
import { useGameWebSocket } from './hooks/useGameWebSocket';
import GameBoard from './components/GameBoard';
import GameOver from './components/GameOver';

type AppState = 'login' | 'waiting' | 'countdown' | 'playing' | 'gameover';

function App() {
  // Auth state
  const [username, setUsername] = useState('');
  const [userId, setUserId] = useState('');
  const [roomId, setRoomId] = useState('');

  // Game state
  const [appState, setAppState] = useState<AppState>('login');
  const [countdown, setCountdown] = useState(3);
  const [playerCount, setPlayerCount] = useState(1);

  // WebSocket hook - connect when playing OR gameover
  const shouldConnect = appState === 'playing' || appState === 'gameover';
  const { gameState, sendAnswer, isConnected } = useGameWebSocket(
    shouldConnect ? roomId : '',
    shouldConnect ? userId : ''
  );

  // Handle login/register
  const handleLogin = async () => {
  if (!username.trim()) return;

  try {
    console.log('Logging in as: ', username);

    // Try login first, register if not found
    let user;
      try {
        user = await loginUser(username);
      } catch {
        console.log('👤 User not found, registering...');
        user = await registerUser(username);
      }

      setUserId(user.user_id);
      console.log('User ID: ', user.user_id);

      // Join matchmaking
      console.log('Joining matchmaking...');
      const room = await joinMatchmaking(user.user_id);
      setRoomId(room.room_id);
      console.log('Room ID:', room.room_id);

      // Set Player count
      setPlayerCount(room.players?.length || 1);

      // Move to waiting state
      setAppState('waiting');
  } catch (error) {
    console.error('Login failed:', error);
  }
};
  // Poll for room readiness
  useEffect(() => {
    if (appState !== 'waiting') return;
    if (!roomId) return;

    console.log('⏳ Polling for opponent...');

    const interval = setInterval(async () => {
      try {
        const ready = await isRoomReady(roomId);

        if (ready) {
          console.log('🎉 Room is ready! Starting countdown...');
          clearInterval(interval);
          setCountdown(3);
          setAppState('countdown');
        }
      } catch (error) {
        console.error('Error checking room status:', error);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [appState, roomId]);

  // Handle countdown decrement
  useEffect(() => {
  if (appState !== 'countdown') return;
  if (countdown > 0) {
    const timer = setTimeout(() => {
      setCountdown(countdown - 1);
    }, 1000);
    return () => clearTimeout(timer);
  }
}, [appState, countdown]);

// Transition to playing state when countdown reaches 0
useEffect(() => {
  if (appState === 'countdown' && countdown === 0) {
    console.log('🚀 Starting game!');
    const timer = setTimeout(() => {
      setAppState('playing');
    }, 100); // Small delay to avoid synchronous update
    return () => clearTimeout(timer);
  }
}, [appState, countdown]);

  // Render based on state
  if (appState === 'login') {
    return (
      <div className="app-container">
        <div className="login-screen">
          <h1>🎨 ColorSync</h1>
          <p>A multiplayer Stroop Test game</p>

          <div className="login-form">
            <input
              type="text"
              placeholder="Enter your username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleLogin()}
              autoFocus
            />
            <button onClick={handleLogin}>
              Join Game
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (appState === 'waiting') {
    return (
      <div className="app-container">
        <div className="waiting-screen">
          <h1>⏳ Waiting for Opponent...</h1>
          <div className="player-count">
            <div className="player-indicator active">Player 1 ✓</div>
            <div className={`player-indicator ${playerCount === 2 ? 'active' : ''}`}>
              Player 2 {playerCount === 2 ? '✓' : '...'}
            </div>
          </div>
          <p className="room-info">Room: {roomId?.slice(0, 8)}...</p>
        </div>
      </div>
    );
  }

  if (appState === 'countdown') {
    return (
      <div className="app-container">
        <div className="countdown-screen">
          <h1>🎮 Get Ready!</h1>
          <div className="countdown-number">{countdown}</div>
          <p>Game starting...</p>
        </div>
      </div>
    );
  }

  if (gameState?.status === 'game_over' && userId && username) {
    return (
      <div className="app-container">
        <GameOver
          gameState={gameState}
          username={username}
          onPlayAgain={() => window.location.reload()}
        />
      </div>
    );
  }

  if (appState === 'playing' && gameState && userId && username) {
    return (
      <div className="app-container">
        <GameBoard
          gameState={gameState}
          sendAnswer={sendAnswer}
          isConnected={isConnected}
          username={username}
        />
      </div>
    );
  }

  return (
    <div className="app-container">
      <div className="loading-screen">
        <h1>Loading...</h1>
      </div>
    </div>
  );
}

export default App;