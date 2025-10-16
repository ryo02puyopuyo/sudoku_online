import React from 'react';
import { Routes, Route } from 'react-router-dom';
import axios from 'axios'; // ★ axiosを追加
import Landing from './pages/Landing';
import Game from './pages/Game';
import './App.css';

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