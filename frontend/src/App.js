// src/App.js

import { useEffect, useState } from "react";
import SudokuBoard from "./SudokuBoard";
import UserList from "./UserList"; // UserListコンポーネントをインポート
import "./App.css";

function App() {
  const [ws, setWs] = useState(null);
  const [puzzle, setPuzzle] = useState([]);
  const [solution, setSolution] = useState([]);
  const [isConnected, setIsConnected] = useState(false);
  const [userList, setUserList] = useState([]); // メンバー一覧を保持するstateを追加

  useEffect(() => {
    const socket = new WebSocket(process.env.REACT_APP_WS_URL);

    socket.onopen = () => {
      console.log("WebSocket connected");
      setIsConnected(true);
      setWs(socket);
    };

    // メッセージ受信処理を修正
    socket.onmessage = (event) => {
      const msg = JSON.parse(event.data);

      // メッセージのタイプに応じて、更新するstateを切り替える
      switch (msg.type) {
        case "puzzle_state":
          setPuzzle(msg.payload.puzzle);
          setSolution(msg.payload.solution);
          console.log("Puzzle state updated");
          break;
        case "user_list":
          setUserList(msg.payload);
          console.log("User list updated:", msg.payload);
          break;
        default:
          console.warn("Received unknown message type:", msg.type);
      }
    };

    socket.onclose = () => {
      console.log("WebSocket disconnected");
      setIsConnected(false);
    };

    socket.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    return () => socket.close();
  }, []);

  const requestNewPuzzle = () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify("new_puzzle"));
      console.log("Requested a new puzzle from the server");
    }
  };

  return (
    <div className="app-container">
      {/* メンバー一覧コンポーネントを追加 */}
      <UserList users={userList} />

      <div className="game-area">
        <h1>リアルタイムナンプレ</h1>
        {!isConnected && <p>サーバーに接続中...</p>}
        {puzzle.length > 0 && solution.length > 0 ? (
          <SudokuBoard
            puzzle={puzzle}
            solution={solution}
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