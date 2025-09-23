import React from 'react';

export default function ScoreBoard({ scores }) {
  return (
    <div className="scoreboard-container">
      <div className="team team-1">
        Team 1: <span>{scores.team1}</span>
      </div>
      <div className="team team-2">
        Team 2: <span>{scores.team2}</span>
      </div>
    </div>
  );
}