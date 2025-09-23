import { useState } from "react";

export default function SudokuBoard({ boardState, onCellClick, onNewGameClick }) {
  const [selectedNumber, setSelectedNumber] = useState(null);
  // boardState (2次元配列) を1次元配列に変換し、値が0のセルの数を数える
  const remainingCells = boardState.flat().filter(cell => cell.value === 0).length;

  const handleCellClick = (r, c) => {
    if (selectedNumber === null) return;
    onCellClick(r, c, selectedNumber);
  };

  const getCellStyle = (cell) => {
    switch (cell.status) {
      case "fixed":
        return { background: "#e9ecef", color: "black", fontWeight: "bold", cursor: "default" };
      case "correct":
        return { background: "white", color: "#1976d2", fontWeight: "bold" };
      case "wrong":
        return { background: "#ffcdd2", color: "#c62828", fontWeight: "bold" };
      default: // empty
        return { background: "white", color: "#1976d2" };
    }
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
          onClick={() => setSelectedNumber(0)} // 0を「消す」としてサーバーに送信
          className={selectedNumber === 0 ? 'selected' : ''}
        >
          消
        </button>
      </div>

      {/* 残りマスとゲーム完了メッセージを表示するエリア */}
      <div className="game-info">
        {remainingCells > 0 ? (
          <span>残りマス: <strong>{remainingCells}</strong></span>
        ) : (
          <span className="game-complete">🎉コンプリート！🎉</span>
        )}
      </div>

      <table className="sudoku-board" style={{ borderCollapse: 'collapse', border: '2px solid black' }}>
        <tbody>
          {boardState.map((row, rIdx) => (
            <tr key={rIdx}>
              {row.map((cell, cIdx) => (
                <td
                  key={cIdx}
                  onClick={() => cell.status !== 'fixed' && handleCellClick(rIdx, cIdx)}
                  style={{
                    width: '40px',
                    height: '40px',
                    textAlign: 'center',
                    fontSize: '20px',
                    cursor: cell.status === 'fixed' ? 'default' : 'pointer',
                    ...getCellStyle(cell),
                    borderTop: '1px solid #ccc',
                    borderLeft: '1px solid #ccc',
                    borderRight: (cIdx + 1) % 3 === 0 ? '2px solid black' : '1px solid #ccc',
                    borderBottom: (rIdx + 1) % 3 === 0 ? '2px solid black' : '1px solid #ccc',
                  }}
                >
                  {cell.value !== 0 ? cell.value : ''}
                </td>
              ))}
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