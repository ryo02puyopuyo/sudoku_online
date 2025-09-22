import { useEffect, useState } from "react";
import SudokuBoard from "./SudokuBoard";
import { generateSolvedGrid } from "./generateProblem";

function App() {
  const [ws, setWs] = useState(null);
  const [puzzle, setPuzzle] = useState([]);
  const [solution, setSolution] = useState([]);

  useEffect(() => {
    // WebSocket
    const socket = new WebSocket(process.env.REACT_APP_WS_URL);
    setWs(socket);
    return () => socket.close();
  }, []);

  useEffect(() => {
    // 問題と解を生成
    const solved = generateSolvedGrid();
    const puzzleGrid = solved.map(row =>
      row.map(cell => (Math.random() < 0.5 ? 0 : cell)) // 半分を空に
    );
    setSolution(solved);
    setPuzzle(puzzleGrid);
  }, []);

  return (
    <div>
      <h1>リアルタイムナンプレ</h1>
      {puzzle.length > 0 && solution.length > 0 && (
        <SudokuBoard ws={ws} puzzle={puzzle} solution={solution} />
      )}
    </div>
  );
}

export default App;
