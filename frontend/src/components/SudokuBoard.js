import { useState } from "react";

export default function SudokuBoard({ boardState, onCellClick, onNewGameClick }) {
  const [selectedNumber, setSelectedNumber] = useState(null);

  const handleCellClick = (r, c) => {
    if (selectedNumber === null) return;
    onCellClick(r, c, selectedNumber);
  };

const getCellStyle = (cell) => {
    let baseStyle = {};
    
    // 既存のステータス別スタイル
    switch (cell.status) {
      case "fixed":
        baseStyle = { background: "#e9ecef", color: "black", fontWeight: "bold" };
        break;
      case "correct":
        const teamColor = cell.filledByTeam === 1
          ? { background: "#e3f2fd", color: "#1976d2" }
          : { background: "#e8f5e9", color: "#2e7d32" };
        baseStyle = { ...teamColor, fontWeight: "bold" };
        break;
      case "wrong":
        baseStyle = { background: "#ffcdd2", color: "#c62828", fontWeight: "bold" };
        break;
      default:
        baseStyle = { background: "white" };
    }

    // 【修正点】ホットスポット用の装飾を追加
    if (cell.isHotSpot) {
      // まだ入力されていないホットスポットは少し黄色く光らせる
      if (cell.status === "empty") {
        baseStyle.background = "#fffde7"; 
      }
      // 強調するための太い枠線（box-shadowを使うと既存のborderを壊しません）
      baseStyle.boxShadow = "inset 0 0 0 3px #ffd700"; 
      baseStyle.zIndex = 1; // 枠線を最前面に
    }

    return baseStyle;
  };

  return (
    <div>
      <div className="controls">
        {[1, 2, 3, 4, 5, 6, 7, 8, 9].map(num => (
          <button
            key={num}
            onClick={() => setSelectedNumber(num)}
            className={selectedNumber === num ? 'selected' : ''}
          >
            {num}
          </button>
        ))}
        <button
          onClick={() => setSelectedNumber(0)}
          className={selectedNumber === 0 ? 'selected' : ''}
        >
          消
        </button>
      </div>
      <table className="sudoku-board" style={{ borderCollapse: 'collapse', border: '2px solid black' }}>
        <tbody>
          {boardState.map((row, rIdx) => (
            <tr key={rIdx}>
              {row.map((cell, cIdx) => {
                const isEditable = cell.status !== 'fixed' && cell.status !== 'correct';
                return (
                  <td
                    key={cIdx}
                    onClick={() => isEditable && handleCellClick(rIdx, cIdx)}
                    style={{
                      width: '40px',
                      height: '40px',
                      textAlign: 'center',
                      fontSize: '20px',
                      cursor: isEditable ? 'pointer' : 'default',
                      ...getCellStyle(cell),
                      borderTop: '1px solid #ccc',
                      borderLeft: '1px solid #ccc',
                      borderRight: (cIdx + 1) % 3 === 0 ? '2px solid black' : '1px solid #ccc',
                      borderBottom: (rIdx + 1) % 3 === 0 ? '2px solid black' : '1px solid #ccc',
                    }}
                  >
                    {cell.value !== 0 ? cell.value : ''}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
      <button onClick={onNewGameClick} className="new-game-button">
        新しい問題
      </button>
    </div>
  );
}