import React from 'react';

export default function UserList({ users }) {
  return (
    <div className="user-list-container">
      <h3>参加中のメンバー ({users.length}人)</h3>
      <ul>
        {users.map(user => (
          <li key={user}>{user}</li>
        ))}
      </ul>
    </div>
  );
}