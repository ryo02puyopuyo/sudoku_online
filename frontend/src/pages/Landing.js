import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import LoginRegisterModal from '../components/LoginRegisterModal';
import '../App.css';
import axios from 'axios';

export default function Landing() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const navigate = useNavigate();

  const announcements = [
    { date: '2026/02/27', content: ' 無料のデプロイサービスを使っているため、しばらく遊ぶ人がいないと、ゲームの起動に1分ほどかかります' },
    { date: '2026/01/27', content: ' ログインシステム実装、ホットスポット実装' },
    { date: '2025/10/**', content: ' ナンプレバトルβ版を公開しました！' },
  ];

  const handleGuestJoin = () => {
    localStorage.removeItem('auth_token');
    navigate('/game');
  };

  //テストボタン
  const handleTestClick = async () => {
    try {
      const response = await axios.post('/api/test', { msg: 'Hello from frontend!' });
      alert('サーバーからのレスポンス: ' + JSON.stringify(response.data));
    } catch (err) {
      console.error('❌ 通信エラー:', err);
      alert('通信に失敗しました。');
    }
  };

  return (
    <div className="landing-container">

      {/* 1. メインコンテンツ */}
      <div className="landing-content">
        <h1 className="landing-title">ナンプレバトル</h1>
        <p className="landing-subtitle">リアルタイム協力＆対戦ナンプレ</p>
        
        <div className="landing-buttons-container">
          <div className="landing-buttons">
            <button className="main-action-button" onClick={handleGuestJoin}>
              ゲストとして参加!
            </button>
          </div>

          <div className="landing-buttons">
            <button className="main-action-button" onClick={() => setIsModalOpen(true)}>
              ログイン・新規登録
            </button>
          </div>
        </div>

        {/* テスト用ボタン*/}
        <div className="test-button-area">
          <button onClick={handleTestClick} className="test-api-button">
            /api/test 通信テスト
          </button>
        </div>
        
      </div>

      {/* 2. お知らせ欄*/}
      <div className="announcement-wrapper">
        <h2 className="announcement-title">お知らせ</h2>
        <div className="announcement-list">
          {announcements.map((item, index) => (
            <div key={index} className="announcement-item">
              <span className="announcement-date">{item.date}</span>
              <span className="announcement-text">{item.content}</span>
            </div>
          ))}
        </div>
      </div>

      <LoginRegisterModal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
      />
    </div>
  );
}