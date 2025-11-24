import './GameBoard.css';
import StroopTest from './StroopTest';
import ScoreBoard from './ScoreBoard';
import type { GameState, StroopColor } from '../types/game';
import { useState } from 'react';

interface GameBoardProps {
  gameState: GameState | null;
  sendAnswer: (color: StroopColor) => void;
  isConnected: boolean;
  username: string;
}

export default function GameBoard({ gameState, sendAnswer, isConnected, username }: GameBoardProps) {
  const [isWaitingForNextRound, setIsWaitingForNextRound] = useState(false);

  if (!gameState) {
    return (
      <div className="game-board">
        <div className="loading">Connecting to game...</div>
      </div>
    );
  }

  const handleColorSelect = (color: StroopColor) => {
    // Disable buttons immediately after click
    setIsWaitingForNextRound(true);
    
    // Send answer
    sendAnswer(color);

    // Re-enable after 1 second (or when next round starts)
    setTimeout(() => {
      setIsWaitingForNextRound(false);
    }, 1000);
  };

  // Disable buttons if:
  // - Not connected
  // - Waiting for next round
  // - Game is over
  const buttonsDisabled = 
    !isConnected || 
    isWaitingForNextRound || 
    gameState.status === 'game_over' ||
    gameState.status === 'waiting';

  return (
    <div className="game-board">
      {/* Connection status */}
      <div className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
        <div className="status-dot"></div>
        {isConnected ? 'Connected' : 'Disconnected'}
      </div>

      {/* Round info */}
      <div className="round-info">
        <h2>Round {gameState.currentRound} of {gameState.maxRounds}</h2>
      </div>

      {/* Score board */}
      <ScoreBoard
        scores={{
          [username]: gameState.myScore,
          'Opponent': gameState.opponentScore
        }}
        currentPlayer={username}
      />

      {/* Stroop word display */}
      {gameState.word && gameState.color && (
        <div className="stroop-display">
          <p className="instruction">What COLOR is this word?</p>
          <div 
            className="stroop-word" 
            style={{ color: gameState.color }}
          >
            {gameState.word}
          </div>
        </div>
      )}

      {/* Color buttons */}
      <StroopTest
        onColorSelect={handleColorSelect}
        disabled={buttonsDisabled}
      />

      {/* Error message */}
      {gameState.errorMessage && (
        <div className="error-message">
          ⚠️ {gameState.errorMessage}
        </div>
      )}
    </div>
  );
}
