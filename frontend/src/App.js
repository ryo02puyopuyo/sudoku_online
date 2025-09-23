import { useEffect, useState } from "react";
import SudokuBoard from "./SudokuBoard";
import UserList from "./UserList";
import ScoreBoard from "./ScoreBoard";
import "./App.css";

function App() {
  const [ws, setWs] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [myID, setMyID] = useState(null);
  const [players, setPlayers] = useState([]);
  const [scores, setScores] = useState({ team1: 0, team2: 0 });
  const [boardState, setBoardState] = useState(null);

  useEffect(() => {
    const socket = new WebSocket(process.env.REACT_APP_WS_URL);
    setWs(socket);

    socket.onopen = () => setIsConnected(true);
    socket.onclose = () => setIsConnected(false);
    socket.onerror = (error) => console.error("WebSocket error:", error);

    socket.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      switch (msg.type) {
        case "welcome":
          setMyID(msg.payload.yourID);
          setBoardState(msg.payload.boardState);
          break;
        case "board_state":
          setBoardState(msg.payload);
          break;
        case "user_list_update":
          setPlayers(msg.payload.players);
          setScores(msg.payload.scores);
          break;
        default:
          break;
      }
    };
    return () => socket.close();
  }, []);

  const sendMessage = (type, payload) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type, payload }));
    }
  };

  const handleCellUpdate = (row, col, value) => sendMessage("cell_update", { row, col, value });
  const requestNewPuzzle = () => sendMessage("new_puzzle", {});
  const handleChangeTeam = (team) => sendMessage("change_team", { team });

  return (
    <div className="app-container">
      <UserList users={players} myID={myID} onChangeTeam={handleChangeTeam} />
      <div className="game-area">
        <h1>リアルタイムナンプレ</h1>
        <ScoreBoard scores={scores} />
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