import React from 'react';
import { Routes, Route } from 'react-router-dom';
import axios from 'axios'; // ★ axiosを追加
import Landing from './pages/Landing';
import Game from './pages/Game';
import './App.css';

// これを1行追加するだけで、全てのAPIリクエストにCookieが自動添付されます
axios.defaults.withCredentials = true;
// ベースURLを設定しておくと便利です
axios.defaults.baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

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