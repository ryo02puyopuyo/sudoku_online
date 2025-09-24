import React from 'react';

export default function UserList({ users, myID, onChangeTeam }) {
  return (
    <div className="user-list-container">
      <h3>参加中のメンバー ({users.length}人)</h3>
      <ul>
        {users.map(user => (
          <li key={user.id} className={user.id === myID ? 'me' : ''}>
            <span className={`team-indicator team-${user.team}`}></span>
            {user.id}
            {user.id === myID && (
              <div className="team-buttons">
                <button
                  onClick={() => onChangeTeam(1)}
                  className={user.team === 1 ? 'active' : ''}
                >1</button>
                <button
                  onClick={() => onChangeTeam(2)}
                  className={user.team === 2 ? 'active' : ''}
                >2</button>
              </div>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}