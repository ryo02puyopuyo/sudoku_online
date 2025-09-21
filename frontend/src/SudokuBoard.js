import { useState, useEffect } from "react";

export default function SudokuBoard({ ws }) {
  // 9x9 の空盤面を初期化
  const [board, setBoard] = useState(
    Array(9).fill(null).map(() => Array(9).fill(""))
  );

  const [selectedNumber, setSelectedNumber] = useState(null); // 選択中の数字

  // セルクリック処理
  const handleCellClick = (r, c) => {
    if (selectedNumber !== null) {
      setBoard(prev => {
        const newBoard = prev.map(row => [...row]);
        newBoard[r][c] = selectedNumber;
        return newBoard;
      });

      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ row: r, col: c, value: selectedNumber }));
      }
    }
  };

  // WebSocket 受信処理
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

  return (
    <div>
      {/* 数字選択UI */}
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
                    cursor: "pointer",
                    fontSize: "18px",
                    fontWeight: "bold",
                    background: cell === "" ? "white" : "#f0f8ff",
                    // 3x3 区切りを太線にする
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
