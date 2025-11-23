import { useState } from 'react';
import { loginOrRegister, joinRoom, waitForGameReady } from './utils/api';
import { useGameWebSocket } from './hooks/useGameWebSocket';
import GameBoard from './components/GameBoard';
import GameOver from './components/GameOver';
import './App.css';

type AppStatus = 'login' | 'matchmaking' | 'waiting' | 'playing';

function App() {
  const [username, setUsername] = useState('');
  const [userId, setUserId] = useState('');
  const [roomId, setRoomId] = useState('');
  const [status, setStatus] = useState<AppStatus>('login');
  const [error, setError] = useState('');

  // ALWAYS call the hook (pass empty strings when not ready)
  const { gameState, sendAnswer, isConnected } = useGameWebSocket(
    roomId || '',  // Empty string if not set
    userId || ''   // Empty string if not set
  );

  // Join Game Flow
  const handleJoinGame = async () => {
    if (!username.trim()) {
      setError('Please enter a username');
      return;
    }

    try {
      setError('');
      setStatus('matchmaking');

      // Step 1: Login or register user
      console.log('Logging in as:', username);
      const uid = await loginOrRegister(username);
      setUserId(uid);
      console.log('User ID:', uid);

      // Step 2: Join matchmaking queue
      console.log('Joining matchmaking queue...');
      const rid = await joinRoom(uid);
      setRoomId(rid);
      console.log('Room ID:', rid);

      // Step 3: Wait for opponent to join
      setStatus('waiting');
      console.log('Waiting for opponent...');
      await waitForGameReady(rid);

      // Step 4: Start game
      setStatus('playing');
      console.log('Game ready! Starting...');
    } catch (err) {
      console.error('Error joining game:', err);
      setError(err instanceof Error ? err.message : 'Failed to join game');
      setStatus('login');
    }
  };

  const handlePlayAgain = () => {
    setUsername('');
    setUserId('');
    setRoomId('');
    setStatus('login');
    setError('');
  };

  // Render: Login Screen
  if (status === 'login') {
    return (
      <div className="app-container login-screen">
        <div className="login-card">
          <h1 className="game-title">🎨 Color Sync</h1>
          <p className="game-subtitle">Test your Stroop effect skills!</p>
          
          <div className="login-form">
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your username"
              className="username-input"
              onKeyPress={(e) => e.key === 'Enter' && handleJoinGame()}
              autoFocus
            />
            <button
              onClick={handleJoinGame}
              className="join-button"
            >
              Join Game
            </button>
          </div>

          {error && (
            <p className="error-message">{error}</p>
          )}

          <div className="game-instructions">
            <h3>How to Play:</h3>
            <ul>
              <li>Read the <strong>COLOR</strong> of the word, not the word itself</li>
              <li>Click the button matching the color</li>
              <li>Fastest correct answer wins the round</li>
              <li>First to 3 wins or best of 5 rounds wins!</li>
            </ul>
          </div>
        </div>
      </div>
    );
  }

  // Render: Matchmaking
  if (status === 'matchmaking') {
    return (
      <div className="app-container loading-screen">
        <div className="loading-card">
          <div className="spinner"></div>
          <h2>🔄 Joining Matchmaking...</h2>
          <p>Setting up your game room...</p>
        </div>
      </div>
    );
  }

  // Render: Waiting for Opponent
  if (status === 'waiting') {
    return (
      <div className="app-container loading-screen">
        <div className="loading-card">
          <div className="spinner"></div>
          <h2>⏳ Finding Opponent...</h2>
          <p className="player-info">Welcome, <strong>{username}</strong>!</p>
          <p className="room-info">Room: {roomId.slice(0, 8)}...</p>
          <p>Waiting for another player to join...</p>
          <div className="waiting-hint">
            💡 Open another browser tab to play against yourself!
          </div>
        </div>
      </div>
    );
  }

  // Render: Playing Game
  if (status === 'playing' && gameState) {
    // Show game over screen
    if (gameState.status === 'game_over') {
      return (
        <GameOver
          gameState={gameState}
          username={username}
          onPlayAgain={handlePlayAgain}
        />
      );
    }

    // Show game board
    return (
      <GameBoard
        gameState={gameState}
        sendAnswer={sendAnswer}
        isConnected={isConnected}
        username={username}
      />
    );
  }

  // Fallback Loading State
  return (
    <div className="app-container loading-screen">
      <div className="loading-card">
        <div className="spinner"></div>
        <h2>Loading...</h2>
      </div>
    </div>
  );
}

export default App;
