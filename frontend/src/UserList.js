// src/UserList.js

import React from 'react';

// usersという名前の配列をプロパティとして受け取る
export default function UserList({ users }) {
  return (
    <div className="user-list-container">
      <h3>参加中のメンバー ({users.length}人)</h3>
      <ul>
        {/* 配列の各要素をループしてリスト項目として表示 */}
        {users.map(user => (
          <li key={user}>{user}</li>
        ))}
      </ul>
    </div>
  );
}