// src/UserList.js
import React, { useState, useEffect } from 'react';

export default function UserList({ users, myPlayer, onChangeName }) {
  const [isEditingName, setIsEditingName] = useState(false);
  const [nameInput, setNameInput] = useState("");

  useEffect(() => {
    if (myPlayer) {
      setNameInput(myPlayer.name);
    }
  }, [myPlayer]);

  const handleNameSubmit = (e) => {
    e.preventDefault();
    if (nameInput.trim() && nameInput.trim() !== myPlayer.name) {
      onChangeName(nameInput.trim());
    }
    setIsEditingName(false);
  };

  return (
    <div className="user-list-container">
      <h3>参加中のメンバー ({users.length}人)</h3>
      <ul>
        {users.map(user => (
          <li key={user.id} className={user.id === myPlayer?.id ? 'me' : ''}>
            <span className={`team-indicator team-${user.team}`}></span>
            
            {user.id === myPlayer?.id && isEditingName ? (
              <form onSubmit={handleNameSubmit} className="name-edit-form">
                <input
                  type="text"
                  value={nameInput}
                  onChange={(e) => setNameInput(e.target.value)}
                  onBlur={handleNameSubmit}
                  autoFocus
                  maxLength="15"
                />
              </form>
            ) : (
              // ▼▼▼ 名前と編集アイコンをグループ化 ▼▼▼
              <>
                <span 
                  className="player-name"
                  onClick={() => {
                    if (user.id === myPlayer?.id) {
                      setIsEditingName(true);
                    }
                  }}
                >
                  {user.name}
                </span>
                {/* 自分自身のプレイヤーの場合のみ編集アイコンを表示 */}
                {user.id === myPlayer?.id && (
                  <span className="edit-name-icon" onClick={() => setIsEditingName(true)}>
                    ✏️
                  </span>
                )}
              </>
              // --- ▲▲▲ ここまで ---
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}