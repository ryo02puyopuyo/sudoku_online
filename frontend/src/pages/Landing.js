import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import LoginRegisterModal from '../components/LoginRegisterModal';
import '../App.css';

export default function Landing() {
const [isModalOpen, setIsModalOpen] = useState(false);
  const navigate = useNavigate();

  return (
    <div className="landing-container">
      <div className="landing-content">
        <h1 className="landing-title">ナンプレバトル</h1>
        <p className="landing-subtitle">リアルタイム協力＆対戦ナンプレ</p>
        <div className="landing-buttons">
          <button 
            className="main-action-button"
            onClick={() => navigate('/game')} // ボタンを押すと /game に移動
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
    </div>
  );
}