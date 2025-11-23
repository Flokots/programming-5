interface ScoreBoardProps {
  currentRound: number;
  maxRounds: number;
  myScore: number;
  opponentScore: number;
}

function ScoreBoard({ currentRound, maxRounds, myScore, opponentScore }: ScoreBoardProps) {
  return (
    <div className="scoreboard">
      {/* Round Counter */}
      <div className="round-display">
        <span className="round-label">Round</span>
        <span className="round-numbers">
          {currentRound} <span className="round-separator">/</span> {maxRounds}
        </span>
      </div>

      {/* Score Display */}
      <div className="score-display">
        <div className="score-section my-score">
          <div className="score-label">YOU</div>
          <div className="score-value">{myScore}</div>
        </div>

        <div className="score-separator">-</div>

        <div className="score-section opponent-score">
          <div className="score-value">{opponentScore}</div>
          <div className="score-label">OPPONENT</div>
        </div>
      </div>

      {/* Progress Bar */}
      <div className="progress-bar">
        <div 
          className="progress-fill my-progress"
          style={{ width: `${(myScore / maxRounds) * 100}%` }}
        ></div>
        <div 
          className="progress-fill opponent-progress"
          style={{ width: `${(opponentScore / maxRounds) * 100}%` }}
        ></div>
      </div>
    </div>
  );
}

export default ScoreBoard;