import React from 'react';
import { Routes, Route } from 'react-router-dom';
import axios from 'axios'; // ★ axiosを追加
import Landing from './pages/Landing';
import Game from './pages/Game';
import './App.css';

function App() {
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
    <div className="app-container">
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/game" element={<Game />} />
      </Routes>

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

export default App;