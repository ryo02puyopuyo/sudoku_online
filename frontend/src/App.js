import React from 'react';
import { Routes, Route } from 'react-router-dom';
import axios from 'axios'; 
import Landing from './pages/Landing';
import Game from './pages/Game';
import './App.css';


// ベースURLを設定しておくと便利です
axios.defaults.baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
//cookkie off
axios.defaults.withCredentials = false;

// 3. 【重要】Axios インターセプターの設定
// 全てのリクエストが送信される直前に、localStorage からトークンを拾ってヘッダーにセットします
axios.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      // Go側のミドルウェアが期待している "Bearer <トークン>" 形式でセット
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

function App() {


  return (
    <div className="app-container">
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/game" element={<Game />} />
      </Routes>
    </div>
  );
}

export default App;