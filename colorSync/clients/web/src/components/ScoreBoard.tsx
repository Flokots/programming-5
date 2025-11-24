import './ScoreBoard.css';

interface ScoreBoardProps {
  scores: Record<string, number>;
  currentPlayer: string;
}

export default function ScoreBoard({ scores, currentPlayer }: ScoreBoardProps) {
  const players = Object.keys(scores);

  return (
    <div className="score-board">
      <h3>Score</h3>
      <div className="scores">
        {players.map((player) => (
          <div
            key={player}
            className={`player-score ${player === currentPlayer ? 'current' : ''}`}
          >
            <div className="player-name">
              {player === currentPlayer ? '👤 ' : '🤖 '}
              {player}
            </div>
            <div className="score">{scores[player]}</div>
          </div>
        ))}
      </div>
    </div>
  );
}