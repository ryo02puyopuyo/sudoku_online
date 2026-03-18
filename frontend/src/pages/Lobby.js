import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import '../App.css';

export default function Lobby() {
  const [rooms, setRooms] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchRooms = async () => {
      try {
        const response = await axios.get('/api/rooms');
        setRooms(response.data.rooms || []);
      } catch (err) {
        console.error('Failed to fetch rooms:', err);
        setError('ルーム情報の取得に失敗しました');
      } finally {
        setLoading(false);
      }
    };

    fetchRooms();
  }, []);

  const handleJoinRoom = (roomId) => {
    navigate(`/game/${roomId}`);
  };

  const handleBack = () => {
    navigate('/');
  };

  return (
    <div className="landing-container">
      <div className="landing-content" style={{ minWidth: '300px' }}>
        <h1 className="landing-title">ロビー</h1>
        <p className="landing-subtitle">参加するルームを選択してください</p>

        {loading ? (
          <p>読み込み中...</p>
        ) : error ? (
          <p style={{ color: 'red' }}>{error}</p>
        ) : (
          <div className="room-list">
            {rooms.length === 0 ? (
              <p>利用可能なルームがありません</p>
            ) : (
              rooms.map((room) => (
                <div key={room.id} className="room-item" style={{
                  background: 'rgba(255, 255, 255, 0.1)',
                  padding: '15px',
                  margin: '10px 0',
                  borderRadius: '10px',
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center'
                }}>
                  <div style={{ flex: '1', fontWeight: 'bold', fontSize: '1.2rem', textAlign: 'left' }}>
                    {room.name}
                  </div>
                  
                  <div style={{ flex: '1', textAlign: 'center' }}>
                    <button 
                      className="main-action-button" 
                      style={{ padding: '8px 20px', fontSize: '1rem', minWidth: '100px' }}
                      onClick={() => handleJoinRoom(room.id)}
                    >
                      参加
                    </button>
                  </div>

                  <div style={{ flex: '1', textAlign: 'center', fontSize: '1rem' }}>
                    👥 {room.playerCount}人
                  </div>

                  <div style={{ flex: '1', textAlign: 'right', fontSize: '1rem' }}>
                    <span style={{ color: '#ff6b6b' }}>T1: {room.scores?.team1 || 0}</span>
                    <span style={{ margin: '0 8px' }}>-</span>
                    <span style={{ color: '#4dabf7' }}>T2: {room.scores?.team2 || 0}</span>
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        <div className="landing-buttons" style={{ marginTop: '30px' }}>
          <button className="main-action-button" style={{ background: '#666' }} onClick={handleBack}>
            タイトルへ戻る
          </button>
        </div>
      </div>
    </div>
  );
}
