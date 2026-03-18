import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import LoginRegisterModal from '../components/LoginRegisterModal';
import '../App.css';
import axios from 'axios';
import { playSE } from '../components/Sound';

// 環境変数でデバッグパネルの表示を切り替え
// .env に REACT_APP_SHOW_DEBUG_PANEL=true を設定すると表示される
const SHOW_DEBUG_PANEL = process.env.REACT_APP_SHOW_DEBUG_PANEL === 'true';

export default function Landing() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const navigate = useNavigate();

  const announcements = [
    { date: '2026/03/18', content: ' 5個の卓を同時に遊べるようになりました。5コンボ10コンボごとにボーナス点獲得。ホットスポット実装' },
    { date: '2026/01/27', content: ' 無料のデプロイサービスを使っているため、しばらく遊ぶ人がいないと、ゲームの起動に1分ほどかかります' },
    { date: '2026/01/27', content: ' ログインシステム実装、ホットスポット実装' },
    { date: '2025/10/**', content: ' ナンプレバトルβ版を公開しました！' },
  ];

  const handleGuestJoin = () => {
    localStorage.removeItem('auth_token');
    navigate('/lobby');
  };

  const handleTestClick = async () => {
    try {
      const response = await axios.post('/api/test', { msg: 'Hello from frontend!' });
      alert('サーバーからのレスポンス: ' + JSON.stringify(response.data));
    } catch (err) {
      console.error('通信エラー:', err);
      alert('通信に失敗しました。');
    }
  };

  return (
    <div className="landing-container">
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

        {SHOW_DEBUG_PANEL && (
          <>
            <div className="test-button-area">
              <button onClick={handleTestClick} className="test-api-button">
                /api/test 通信テスト
              </button>
            </div>

            <div style={{
              background: 'rgba(255,255,255,0.1)',
              padding: '20px',
              borderRadius: '10px',
              marginTop: '20px'
            }}>
              <p>🔊 SEテストパネル</p>
              <button onClick={() => playSE('correct')}>決定音1</button>
              <button onClick={() => playSE('incorrect')}>ビープ音</button>
              <button onClick={() => playSE('hotspot')}>決定音2</button>
            </div>
          </>
        )}
      </div>

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