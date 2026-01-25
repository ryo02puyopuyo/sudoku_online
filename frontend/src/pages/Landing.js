import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import LoginRegisterModal from '../components/LoginRegisterModal';
import '../App.css';
import axios from 'axios';

export default function Landing() {
const [isModalOpen, setIsModalOpen] = useState(false);
  const navigate = useNavigate();

  // ★ 変更点：ゲスト参加用の関数を作成
    const handleGuestJoin = () => {
        // localStorage を空にして、確実にゲストとして扱う
        localStorage.removeItem('auth_token');
        navigate('/game');
    };

    // テストボタンのクリック処理
  const handleTestClick = async () => {
    try {
      const response = await axios.post('/api/test', { msg: 'Hello from frontend!' });
      alert('サーバーからのレスポンス: ' + JSON.stringify(response.data));
      console.log('✅ /api/test レスポンス:', response.data);
    } catch (err) {
      console.error('❌ 通信エラー:', err);
      alert('通信に失敗しました。サーバーが起動しているか確認してください。');
    }
  };

  return (
    <div className="landing-container">
      <div className="landing-content">
        <h1 className="landing-title">ナンプレバトル</h1>
        <p className="landing-subtitle">リアルタイム協力＆対戦ナンプレ</p>
        <div className="landing-buttons">
          <button 
            className="main-action-button"
            onClick={handleGuestJoin} // ★ 修正
          >
            ゲストとして参加!
          </button>
        </div>


        <div className="landing-buttons">
          <button 
            className="main-action-button"
            onClick={() => setIsModalOpen(true)}
          >
            ログイン・新規登録
          </button>
        </div>
      </div>
            <LoginRegisterModal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
      />

      {/* ✅ テスト用ボタン */}
      <div style={{ textAlign: 'center', marginTop: '20px' }}>
        <button
          onClick={handleTestClick}
          style={{
            padding: '10px 20px',
            borderRadius: '8px',
            backgroundColor: '#007bff',
            color: 'white',
            border: 'none',
            cursor: 'pointer',
          }}
        >
          /api/test 通信テスト
        </button>
      </div>
    </div>
    
  );
}