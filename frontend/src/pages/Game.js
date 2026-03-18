import { useEffect, useState, useRef } from "react";
import SudokuBoard from "../components/SudokuBoard";
import UserList from "../components/UserList";
import ScoreBoard from "../components/ScoreBoard";
import TeamSelector from "../components/TeamSelector";
import Chat from "../components/Chat";
import ResultModal from "../components/ResultModal";
import { useParams, useNavigate } from "react-router-dom";

export default function Game() {
  const { roomId } = useParams();
  const navigate = useNavigate();

  const [isConnected, setIsConnected] = useState(false);
  const [myPlayer, setMyPlayer] = useState(null);
  const [players, setPlayers] = useState([]);
  const [scores, setScores] = useState({ team1: 0, team2: 0 });
  const [boardState, setBoardState] = useState(null);
  const [chatMessages, setChatMessages] = useState([]);
  const [gameOverResult, setGameOverResult] = useState(null);
  const [isSidebarOpen, setIsSidebarOpen] = useState(true);
  const [roomName, setRoomName] = useState("");

  const ws = useRef(null);

  useEffect(() => {
    if (!roomId) {
      navigate('/lobby');
      return;
    }

    const token = localStorage.getItem('auth_token');
    const wsUrl = token
      ? `${process.env.REACT_APP_WS_URL}?token=${token}&room=${roomId}`
      : `${process.env.REACT_APP_WS_URL}?room=${roomId}`;

    const socket = new WebSocket(wsUrl);
    ws.current = socket;

    socket.onopen = () => setIsConnected(true);
    socket.onclose = () => setIsConnected(false);
    socket.onerror = (error) => console.error("WebSocket error:", error);

    socket.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      switch (msg.type) {
        case "welcome":
          setMyPlayer(msg.payload.yourPlayer);
          setBoardState(msg.payload.boardState);
          if (msg.payload.roomName) {
            setRoomName(msg.payload.roomName);
          }
          break;
        case "board_state":
          setBoardState(msg.payload);
          break;
        case "user_list_update":
          setPlayers(msg.payload.players);
          setScores(msg.payload.scores);
          setMyPlayer(currentMyPlayer => {
            if (!currentMyPlayer) return null;
            const me = msg.payload.players.find(p => p.id === currentMyPlayer.id);
            return me ? me : currentMyPlayer;
          });
          break;
        case "new_chat_message":
          setChatMessages(prevMessages => [...prevMessages, msg.payload]);
          break;
        case "game_over":
          setGameOverResult(msg.payload);
          break;
        case "new_game_started":
          setGameOverResult(null);
          break;
        default:
          break;
      }
    };

    return () => {
      socket.close();
    };
  }, [roomId, navigate]);

  const sendMessage = (type, payload) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type, payload }));
    }
  };

  const handleCellUpdate = (row, col, value) => sendMessage("cell_update", { row, col, value });
  const requestNewPuzzle = () => sendMessage("new_puzzle", {});
  const handleChangeTeam = (team) => sendMessage("change_team", { team });
  const handleChangeName = (name) => sendMessage("change_name", { name });
  const handleSendMessage = (message) => sendMessage("send_chat_message", { message });

  return (
    <>
      <ResultModal result={gameOverResult} onNewGame={requestNewPuzzle} />

      {isSidebarOpen && (
        <div className="sidebar-container">
          <UserList
            users={players}
            myPlayer={myPlayer}
            onChangeName={handleChangeName}
          />
          <Chat messages={chatMessages} onSendMessage={handleSendMessage} />
        </div>
      )}

      <div className="game-area">
        <h1>{roomName ? roomName : "ナンプレバトル"}</h1>
        <ScoreBoard scores={scores} />
        <TeamSelector myPlayer={myPlayer} onChangeTeam={handleChangeTeam} />
        {!isConnected && <p>サーバーに接続中...</p>}
        {boardState ? (
          <SudokuBoard
            boardState={boardState}
            onCellClick={handleCellUpdate}
            onNewGameClick={requestNewPuzzle}
          />
        ) : (
          isConnected && <p>問題を読み込んでいます...</p>
        )}

        <button
          onClick={() => setIsSidebarOpen(!isSidebarOpen)}
          className="sidebar-toggle-main-button"
        >
          {isSidebarOpen ? 'チャット/メンバーを隠す' : 'チャット/メンバーを表示'}
        </button>

        <div style={{ marginTop: '20px', textAlign: 'center' }}>
          <button
            onClick={() => navigate('/lobby')}
            className="main-action-button"
            style={{ background: '#666' }}
          >
            ロビーに戻る
          </button>
        </div>
      </div>
    </>
  );
}