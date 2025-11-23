import type { GameState, StroopColor } from '../types/game';
import ScoreBoard from './ScoreBoard';
import StroopTest from './StroopTest';
import './GameBoard.css';

interface GameBoardProps {
  gameState: GameState;
  sendAnswer: (color: StroopColor) => void;
  isConnected: boolean;
  username: string;
}

function GameBoard({ gameState, sendAnswer, isConnected, username }: GameBoardProps) {
  
  // Waiting for Game to Start
  if (gameState.status === 'waiting') {
    return (
      <div className="game-board waiting-state">
        <div className="waiting-content">
          <div className="pulse-animation">⏳</div>
          <h2>Get Ready!</h2>
          <p>Game starting soon...</p>
          <div className="connection-status">
            {isConnected ? (
              <span className="connected">Connected</span>
            ) : (
              <span className="disconnected">Connecting...</span>
            )}
          </div>
        </div>
      </div>
    );
  }

  // Error State
  if (gameState.status === 'error') {
    return (
      <div className="game-board error-state">
        <div className="error-content">
          <h2>Connection Error</h2>
          <p>{gameState.errorMessage || 'Something went wrong'}</p>
          <button 
            onClick={() => window.location.reload()}
            className="retry-button"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  // Main Game Board (Playing)
  return (
    <div className="game-board playing-state">
      {/* Header with connection status */}
      <div className="game-header">
        <div className="player-name">
          👤 {username}
        </div>
        <div className="connection-indicator">
          {isConnected ? (
            <span className="status-dot connected"></span>
          ) : (
            <span className="status-dot disconnected"></span>
          )}
        </div>
      </div>

      {/* Score Display */}
      <ScoreBoard
        currentRound={gameState.currentRound}
        maxRounds={gameState.maxRounds}
        myScore={gameState.myScore}
        opponentScore={gameState.opponentScore}
      />

      {/* Stroop Test Area */}
      <div className="stroop-container">
        {gameState.word && gameState.color ? (
          <StroopTest
            word={gameState.word}
            color={gameState.color}
          />
        ) : (
          <div className="waiting-for-round">
            <div className="pulse-animation">⏱️</div>
            <p>Next round starting...</p>
          </div>
        )}
      </div>

      {/* Error Message Display */}
      {gameState.errorMessage && (
        <div className="error-banner">
          {gameState.errorMessage}
        </div>
      )}

      {/* Answer Buttons */}
      <div className="answer-buttons">
        <button
          onClick={() => sendAnswer('red')}
          className="answer-button red"
          disabled={!gameState.word || gameState.status !== 'playing'}
        >
          RED
        </button>
        <button
          onClick={() => sendAnswer('blue')}
          className="answer-button blue"
          disabled={!gameState.word || gameState.status !== 'playing'}
        >
          BLUE
        </button>
        <button
          onClick={() => sendAnswer('green')}
          className="answer-button green"
          disabled={!gameState.word || gameState.status !== 'playing'}
        >
          GREEN
        </button>
        <button
          onClick={() => sendAnswer('yellow')}
          className="answer-button yellow"
          disabled={!gameState.word || gameState.status !== 'playing'}
        >
          YELLOW
        </button>
      </div>

      {/* Game Instructions */}
      <div className="game-hint">
        💡 Click the <strong>COLOR</strong> of the word, not what it says!
      </div>
    </div>
  );
}

export default GameBoard;