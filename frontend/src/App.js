import { useEffect, useState } from "react";
import SudokuBoard from "./SudokuBoard";

function App() {
  const [ws, setWs] = useState(null);

  useEffect(() => {
    const socket = new WebSocket("ws://localhost:8080/ws");
    setWs(socket);

    return () => socket.close();
  }, []);

  return (
    <div>
      <h1>リアルタイムナンプレ</h1>
      <SudokuBoard ws={ws} />
    </div>
  );
}

export default App;