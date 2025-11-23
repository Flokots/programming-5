import type { GameState } from '../types/game';
import './GameOver.css';

interface GameOverProps {
  gameState: GameState;
  username: string;
  onPlayAgain: () => void;
}

function GameOver({ gameState, username, onPlayAgain }: GameOverProps) {
  const isWinner = gameState.winner === gameState.userId;
  const isDraw = gameState.winner === 'draw';
  const wasDisconnect = gameState.myStats === null;

  // Determine game result message
  const getResultMessage = () => {
    if (wasDisconnect) {
      return {
        emoji: '🔌',
        title: 'Connection Lost',
        subtitle: 'The game ended unexpectedly'
      };
    }
    if (isDraw) {
      return {
        emoji: '🤝',
        title: "It's a Draw!",
        subtitle: 'You both played equally well'
      };
    }
    if (isWinner) {
      return {
        emoji: '🎉',
        title: 'You Won!',
        subtitle: 'Congratulations, champion!'
      };
    }
    return {
      emoji: '😔',
      title: 'You Lost',
      subtitle: 'Better luck next time!'
    };
  };

  const result = getResultMessage();

  return (
    <div className={`game-over ${isWinner ? 'winner' : 'loser'}`}>
      <div className="game-over-card">
        {/* Result Header */}
        <div className="result-header">
          <div className="result-emoji">{result.emoji}</div>
          <h1 className="result-title">{result.title}</h1>
          <p className="result-subtitle">{result.subtitle}</p>
        </div>

        {/* Final Score */}
        {!wasDisconnect && (
          <div className="final-score">
            <div className="score-row">
              <div className={`player-score ${isWinner ? 'winner-score' : ''}`}>
                <div className="player-label">
                  {username} {isWinner && '👑'}
                </div>
                <div className="player-points">{gameState.myScore}</div>
              </div>

              <div className="score-divider">-</div>

              <div className={`player-score ${!isWinner && !isDraw ? 'winner-score' : ''}`}>
                <div className="player-points">{gameState.opponentScore}</div>
                <div className="player-label">
                  Opponent {!isWinner && !isDraw && '👑'}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Player Statistics */}
        {gameState.myStats && (
          <div className="stats-section">
            <h3 className="stats-title">📊 Your Stats</h3>
            <div className="stats-grid">
              <div className="stat-item">
                <div className="stat-label">Rounds Won</div>
                <div className="stat-value">{gameState.myStats.wins}</div>
              </div>
              <div className="stat-item">
                <div className="stat-label">Rounds Lost</div>
                <div className="stat-value">
                  {gameState.maxRounds - gameState.myStats.wins}
                </div>
              </div>
              <div className="stat-item">
                <div className="stat-label">Total Latency</div>
                <div className="stat-value">{gameState.myStats.total_latency}ms</div>
              </div>
              <div className="stat-item">
                <div className="stat-label">Avg Response</div>
                <div className="stat-value highlight">
                  {Math.round(gameState.myStats.avg_latency)}ms
                </div>
              </div>
            </div>

            {/* Performance Badge */}
            {gameState.myStats.avg_latency < 1000 && (
              <div className="performance-badge fast">
                ⚡ Lightning Fast Reflexes!
              </div>
            )}
            {gameState.myStats.avg_latency >= 1000 && gameState.myStats.avg_latency < 2000 && (
              <div className="performance-badge good">
                👍 Good Response Time!
              </div>
            )}
            {gameState.myStats.avg_latency >= 2000 && (
              <div className="performance-badge slow">
                🐢 Take Your Time Next Round!
              </div>
            )}
          </div>
        )}

        {/* Disconnection Message */}
        {wasDisconnect && (
          <div className="disconnect-message">
            <p>⚠️ The connection was lost or your opponent disconnected.</p>
            <p>Stats are not available for incomplete games.</p>
          </div>
        )}

        {/* Action Buttons */}
        <div className="action-buttons">
          <button onClick={onPlayAgain} className="play-again-button">
            🎮 Play Again
          </button>
          <button 
            onClick={() => window.location.href = '/'} 
            className="home-button"
          >
            🏠 Home
          </button>
        </div>

        {/* Fun Fact */}
        <div className="fun-fact">
          <p>💡 <strong>Did you know?</strong></p>
          <p>The Stroop effect was discovered in 1935 by John Ridley Stroop!</p>
        </div>
      </div>
    </div>
  );
}

export default GameOver;