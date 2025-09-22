import { useState, useEffect } from "react";

export default function SudokuBoard({ ws, puzzle, solution }) {
  const [board, setBoard] = useState(
    puzzle.map(row => row.map(cell => (cell === 0 ? "" : cell)))
  );
  const [statusBoard, setStatusBoard] = useState(
    puzzle.map(row => row.map(cell => (cell === 0 ? "empty" : "fixed")))
  );
  const [selectedNumber, setSelectedNumber] = useState(null);

  const handleCellClick = (r, c) => {
    if (selectedNumber === null) return;
    if (statusBoard[r][c] === "fixed" || statusBoard[r][c] === "correct") return;

    const value = selectedNumber;

    setBoard(prev => {
      const newBoard = prev.map(row => [...row]);
      newBoard[r][c] = value;
      return newBoard;
    });

    setStatusBoard(prev => {
      const newStatus = prev.map(row => [...row]);
      if (value === "") newStatus[r][c] = "empty";
      else if (parseInt(value) === solution[r][c]) newStatus[r][c] = "correct";
      else newStatus[r][c] = "wrong";
      return newStatus;
    });

    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ row: r, col: c, value }));
    }
  };

  useEffect(() => {
    if (!ws) return;
    ws.onmessage = (event) => {
      const { row, col, value } = JSON.parse(event.data);
      setBoard(prev => {
        const newBoard = prev.map(r => [...r]);
        newBoard[row][col] = value;
        return newBoard;
      });
    };
  }, [ws]);

  const getCellStyle = (status) => {
    switch (status) {
      case "fixed":
      case "correct":
        return { background: "#dcdcdc" }; // 灰色
      case "wrong":
        return { background: "#ffcccc" }; // 赤
      default:
        return { background: "white" };
    }
  };

  return (
    <div>
      {/* 数字選択 */}
      <div style={{ marginBottom: "10px" }}>
        {[1,2,3,4,5,6,7,8,9].map(num => (
          <button
            key={num}
            onClick={() => setSelectedNumber(num)}
            style={{
              margin: "2px",
              background: selectedNumber === num ? "lightblue" : "white"
            }}
          >
            {num}
          </button>
        ))}
        <button
          onClick={() => setSelectedNumber("")}
          style={{
            margin: "2px",
            background: selectedNumber === "" ? "lightblue" : "white"
          }}
        >
          消す
        </button>
      </div>

      {/* 盤面 */}
      <table style={{ borderCollapse: "collapse" }}>
        <tbody>
          {board.map((row, rIdx) => (
            <tr key={rIdx}>
              {row.map((cell, cIdx) => (
                <td
                  key={cIdx}
                  onClick={() => handleCellClick(rIdx, cIdx)}
                  style={{
                    width: "35px",
                    height: "35px",
                    textAlign: "center",
                    border: "1px solid black",
                    cursor:
                      statusBoard[rIdx][cIdx] === "fixed" ||
                      statusBoard[rIdx][cIdx] === "correct"
                        ? "default"
                        : "pointer",
                    fontSize: "18px",
                    fontWeight: "bold",
                    ...getCellStyle(statusBoard[rIdx][cIdx]),
                    borderRight:
                      (cIdx + 1) % 3 === 0 && cIdx !== 8
                        ? "3px solid black"
                        : "1px solid black",
                    borderBottom:
                      (rIdx + 1) % 3 === 0 && rIdx !== 8
                        ? "3px solid black"
                        : "1px solid black"
                  }}
                >
                  {cell}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
