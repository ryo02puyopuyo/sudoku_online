import React from 'react';

export default function TeamSelector({ myPlayer, onChangeTeam }) {
  if (!myPlayer) return null;

  return (
    <div className="team-selector-container">
      <h3>チームを選択</h3>
      <div className="team-selector-buttons">
        <button
          onClick={() => onChangeTeam(1)}
          className={`team-button team-1 ${myPlayer.team === 1 ? 'active' : ''}`}
        >
          Team 1
        </button>
        <button
          onClick={() => onChangeTeam(2)}
          className={`team-button team-2 ${myPlayer.team === 2 ? 'active' : ''}`}
        >
          Team 2
        </button>
      </div>
    </div>
  );
}