import { useState, useEffect } from "react";

// 渡された puzzle プロパティに基づいて初期状態を生成するヘルパー関数
const getInitialState = (puzzle) => ({
  board: puzzle.map(row => row.map(cell => (cell === 0 ? "" : cell))),
  statusBoard: puzzle.map(row => row.map(cell => (cell === 0 ? "empty" : "fixed"))),
});

export default function SudokuBoard({ puzzle, solution, onNewGameClick }) {
  const [board, setBoard] = useState(getInitialState(puzzle).board);
  const [statusBoard, setStatusBoard] = useState(getInitialState(puzzle).statusBoard);
  const [selectedNumber, setSelectedNumber] = useState(null);

  // puzzleプロパティが変更されたら（＝新しい問題が届いたら）、盤面の表示をリセットする
  useEffect(() => {
    const initialState = getInitialState(puzzle);
    setBoard(initialState.board);
    setStatusBoard(initialState.statusBoard);
  }, [puzzle]);

  // セルがクリックされたときの処理
  const handleCellClick = (r, c) => {
    if (selectedNumber === null) return; // 数字が選択されていなければ何もしない
    // 固定マスは変更不可
    if (statusBoard[r][c] === "fixed") return;

    const value = selectedNumber;

    // 盤面の数字を更新
    setBoard(prev => {
      const newBoard = prev.map(row => [...row]);
      newBoard[r][c] = value;
      return newBoard;
    });

    // マスの状態（正解/不正解）を更新
    setStatusBoard(prev => {
      const newStatus = prev.map(row => [...row]);
      if (value === "") {
        newStatus[r][c] = "empty";
      } else if (parseInt(value, 10) === solution[r][c]) {
        newStatus[r][c] = "correct";
      } else {
        newStatus[r][c] = "wrong";
      }
      return newStatus;
    });
  };

  // マスの状態に応じてスタイルを返す
  const getCellStyle = (status) => {
    switch (status) {
      case "fixed":
        return { background: "#e9ecef", color: "black", fontWeight: "bold" };
      case "correct":
        return { background: "white", color: "#1976d2", fontWeight: "bold" };
      case "wrong":
        return { background: "#ffcdd2", color: "#c62828", fontWeight: "bold" };
      default: // empty
        return { background: "white", color: "#1976d2", fontWeight: "normal" };
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
          onClick={() => setSelectedNumber("")}
          className={selectedNumber === "" ? 'selected' : ''}
        >
          消
        </button>
      </div>

      <table className="sudoku-board" style={{ borderCollapse: 'collapse', border: '2px solid black' }}>
        <tbody>
          {board.map((row, rIdx) => (
            <tr key={rIdx}>
              {row.map((cell, cIdx) => (
                <td
                  key={cIdx}
                  onClick={() => handleCellClick(rIdx, cIdx)}
                  style={{
                    // ⬇️ ここからが修正・追加したスタイルです ⬇️
                    
                    // 基本のセルスタイル
                    width: '40px',
                    height: '40px',
                    textAlign: 'center',
                    fontSize: '20px',
                    cursor: statusBoard[rIdx][cIdx] === "fixed" ? "default" : "pointer",
                    
                    // 正解・不正解などに応じた色設定
                    ...getCellStyle(statusBoard[rIdx][cIdx]),

                    // 3x3ブロックを区切るための太い枠線
                    borderTop: '1px solid #ccc',
                    borderLeft: '1px solid #ccc',
                    borderRight: (cIdx + 1) % 3 === 0 ? '2px solid black' : '1px solid #ccc',
                    borderBottom: (rIdx + 1) % 3 === 0 ? '2px solid black' : '1px solid #ccc',
                  }}
                >
                  {cell}
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