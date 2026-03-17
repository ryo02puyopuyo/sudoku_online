import React from 'react';

export default function ResultModal({ result, onNewGame }) {
  if (!result) return null;

  const getWinnerText = () => {
    if (result.winnerTeam === 0) {
      return "引き分け";
    }
    return `Team ${result.winnerTeam} の勝利!`;
  };

  return (
    <div className="modal-overlay">
      <div className="modal-content">
        <h2>ゲーム終了！</h2>
        <h3 className={`winner-text team-${result.winnerTeam}`}>{getWinnerText()}</h3>
        <div className="final-scores">
          <div className="final-score-team team-1">
            Team 1: <span>{result.finalScores.team1}</span>
          </div>
          <div className="final-score-team team-2">
            Team 2: <span>{result.finalScores.team2}</span>
          </div>
        </div>
        <button onClick={onNewGame} className="new-game-button modal-button">
          新しい問題で遊ぶ
        </button>
      </div>
    </div>
  );
}