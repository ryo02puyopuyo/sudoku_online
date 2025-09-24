import { useEffect, useState, useRef } from "react";
import SudokuBoard from "./SudokuBoard";
import UserList from "./UserList";
import ScoreBoard from "./ScoreBoard";
import TeamSelector from "./TeamSelector";
import Chat from "./Chat";
import "./App.css";

function App() {
  const [isConnected, setIsConnected] = useState(false);
  const [myPlayer, setMyPlayer] = useState(null);
  const [players, setPlayers] = useState([]);
  const [scores, setScores] = useState({ team1: 0, team2: 0 });
  const [boardState, setBoardState] = useState(null);
  const [chatMessages, setChatMessages] = useState([]);
  
  const ws = useRef(null);

  useEffect(() => {
    const socket = new WebSocket(process.env.REACT_APP_WS_URL);
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
        default:
          break;
      }
    };

    return () => {
      socket.close();
    };
  }, []);

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
    <div className="app-container">
      <div className="sidebar-container">
        <UserList
          users={players}
          myPlayer={myPlayer}
          onChangeName={handleChangeName}
        />
        <Chat messages={chatMessages} onSendMessage={handleSendMessage} />
      </div>
      <div className="game-area">
        <h1>リアルタイムナンプレ</h1>
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
      </div>
    </div>
  );
}
export default App;