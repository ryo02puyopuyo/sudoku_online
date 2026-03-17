import React, { useState } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';

export default function LoginRegisterModal({ isOpen, onClose }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  if (!isOpen) {
    return null;
  }

  const handleRegister = async () => {
    setError('');
    if (!username || !password) {
      setError('ユーザー名とパスワードを入力してください。');
      return;
    }
    try {
      await axios.post('/api/register', { username, password });
      alert('登録が完了しました。続けてログインしてください。');
    } catch (err) {
      setError(err.response?.data || '登録に失敗しました。');
    }
  };

  const handleLogin = async () => {
    setError('');
    if (!username || !password) {
      setError('ユーザー名とパスワードを入力してください。');
      return;
    }
    try {
      const response = await axios.post('/api/login', { username, password });
      const { token } = response.data;
      if (token) {
        localStorage.setItem('auth_token', token);
      }
      onClose();
      navigate('/game');
    } catch (err) {
      setError('ログインに失敗しました。ユーザー名またはパスワードを確認してください。');
    }
  };

  const handleOverlayClick = (e) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  return (
    <div className="modal-overlay" onClick={handleOverlayClick}>
      <div className="modal-content">
        <button onClick={onClose} className="close-button">×</button>
        <h2>ログイン / 新規登録</h2>

        {error && <p className="error-message">{error}</p>}

        <div className="input-group">
          <label htmlFor="modal-username">ユーザー名</label>
          <input type="text" id="modal-username" value={username} onChange={(e) => setUsername(e.target.value)} required />
        </div>
        <div className="input-group">
          <label htmlFor="modal-password">パスワード</label>
          <input type="password" id="modal-password" value={password} onChange={(e) => setPassword(e.target.value)} required />
        </div>

        <div className="modal-actions">
          <button onClick={handleRegister} className="register-button">新規登録</button>
          <button onClick={handleLogin} className="login-button">ログイン</button>
        </div>
      </div>
    </div>
  );
}
