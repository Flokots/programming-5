import {useState, useEffect } from 'react';
import { registerUser, loginUser, joinMatchmaking, isRoomReady } from './utils/api';
import { useGameWebSocket } from './hooks/useGameWebSocket';
import GameBoard from './components/GameBoard';
import StroopTest from './components/StroopTest';
import ScoreBoard from './components/ScoreBoard';
import GameOver from './components/GameOver';


type AppState = 'login' | 'waiting' | 'countdown' | 'playing' | 'gameover';

function App() {
  // Auth state
  const [username, setUsername] = useState('');
  const [userId, setUserId] = useState<string | null>(null);
  const [roomId, setRoomId] = useState<string | null>(null);

  // Game state
  const [appState, setAppState] = useState<AppState>('login');
  const [countdown, setCountdown] = useState(3);
  const [playerCount, setPlayerCount] = useState(1);

  // WebSocket hook - only connect when playing 
  const { gameState, sendMove } = useGameWebSocket(
    appState === 'playing' || appState === 'gameover' ? roomId : null,
    appState === 'playing' || appState === 'gameover' ? userId : null
  );

  // Handle login/register
  const handleLogin = async () => {
    if (!username.trim()) return;

    try {
      console.log('Logging in as:', username);

      // Try login first, register if not found
      let user;
      try {
        user = await loginUser(username);
      } catch (err) {
        console.log('User not found, registering new user');
        user = await registerUser(username);
      }

      setUserId(user.id);
      console.log('User ID:', user.id);

      // Join matchmaking
      console.log('Joining matchmaking...');
      const room = await joinMatchmaking(user.user_d);
      }
    }
  }
}