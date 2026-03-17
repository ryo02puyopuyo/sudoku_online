import React from 'react';
import { Routes, Route } from 'react-router-dom';
import axios from 'axios';
import Landing from './pages/Landing';
import Game from './pages/Game';
import './App.css';

axios.defaults.baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080';
axios.defaults.withCredentials = false;

// リクエスト送信前にlocalStorageからJWTトークンをAuthorizationヘッダーにセット
axios.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token');
    if (token) {
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