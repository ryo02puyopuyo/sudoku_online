import { useEffect, useState } from "react";
import SudokuBoard from "./SudokuBoard";
import UserList from "./UserList";
import "./App.css";

function App() {
  const [ws, setWs] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [userList, setUserList] = useState([]);
  const [boardState, setBoardState] = useState(null); // サーバーから送られてくる盤面状態

  useEffect(() => {
    const socket = new WebSocket(process.env.REACT_APP_WS_URL);
    setWs(socket);

    socket.onopen = () => setIsConnected(true);
    socket.onclose = () => setIsConnected(false);
    socket.onerror = (error) => console.error("WebSocket error:", error);

    socket.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      switch (msg.type) {
        case "board_state":
          setBoardState(msg.payload);
          break;
        case "user_list":
          setUserList(msg.payload);
          break;
        default:
          break;
      }
    };

    return () => socket.close();
  }, []);

  // サーバーにメッセージを送信するためのヘルパー関数
  const sendMessage = (type, payload) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      const message = { type, payload };
      ws.send(JSON.stringify(message));
    }
  };

  // セル更新リクエストを送信する関数
  const handleCellUpdate = (row, col, value) => {
    sendMessage("cell_update", { row, col, value });
  };

  // 新しい問題のリクエストを送信する関数
  const requestNewPuzzle = () => {
    sendMessage("new_puzzle", {});
  };

  return (
    <div className="app-container">
      <UserList users={userList} />
      <div className="game-area">
        <h1>リアルタイムナンプレ</h1>
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