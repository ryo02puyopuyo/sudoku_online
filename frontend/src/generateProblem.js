// sudoku-generator.js
// 実行例: node sudoku-generator.js
// これは「解（完成盤）」を生成するコードです。

// shuffle helper
function shuffle(array) {
  for (let i = array.length - 1; i > 0; --i) {
    const j = Math.floor(Math.random() * (i + 1));
    [array[i], array[j]] = [array[j], array[i]];
  }
  return array;
}

function boxIndex(r, c) {
  return Math.floor(r / 3) * 3 + Math.floor(c / 3);
}

function makeEmptyGrid() {
  return Array.from({ length: 9 }, () => Array(9).fill(0));
}

/**
 * generateSolvedGrid
 * - 塗り始めとして row0 をランダムに埋める
 * - 残マスを DFS で埋める（最初に見つかった解で終了）
 * returns solved 9x9 array (numbers 1..9)
 */
function generateSolvedGrid(maxAttempts = 10) {
  for (let attempt = 0; attempt < maxAttempts; ++attempt) {
    // grid, usage trackers
    const grid = makeEmptyGrid();

    // trackers: usedRow[r][n], usedCol[c][n], usedBox[b][n]
    const usedRow = Array.from({ length: 9 }, () => Array(10).fill(false));
    const usedCol = Array.from({ length: 9 }, () => Array(10).fill(false));
    const usedBox = Array.from({ length: 9 }, () => Array(10).fill(false));

    // fill top row with shuffled 1..9
    const nums = shuffle([1,2,3,4,5,6,7,8,9].slice());
    for (let c = 0; c < 9; ++c) {
      const v = nums[c];
      grid[0][c] = v;
      usedRow[0][v] = true;
      usedCol[c][v] = true;
      usedBox[boxIndex(0, c)][v] = true;
    }

    // make cell list for remaining empty cells (row-major)
    const cellList = [];
    for (let r = 0; r < 9; ++r) {
      for (let c = 0; c < 9; ++c) {
        if (grid[r][c] === 0) cellList.push([r, c]);
      }
    }

    // recursive DFS
    let solved = false;
    function dfs(idx) {
      if (idx === cellList.length) {
        solved = true;
        return true;
      }
      const [r, c] = cellList[idx];

      // compute candidates 1..9 not used
      const cand = [];
      for (let n = 1; n <= 9; ++n) {
        if (!usedRow[r][n] && !usedCol[c][n] && !usedBox[boxIndex(r,c)][n]) {
          cand.push(n);
        }
      }
      // randomize candidate order
      shuffle(cand);

      for (const n of cand) {
        // place
        grid[r][c] = n;
        usedRow[r][n] = true;
        usedCol[c][n] = true;
        usedBox[boxIndex(r,c)][n] = true;

        if (dfs(idx + 1)) return true;

        // undo
        grid[r][c] = 0;
        usedRow[r][n] = false;
        usedCol[c][n] = false;
        usedBox[boxIndex(r,c)][n] = false;
      }

      // no candidate leads to solution -> backtrack
      return false;
    }

    if (dfs(0)) {
      return grid; // solved grid
    }
    // else attempt again by reshuffling top row
  }

  throw new Error("Failed to generate a solved grid after many attempts");
}

/* -- optional: pretty print -- */
function printGrid(grid) {
  for (let r = 0; r < 9; ++r) {
    let line = "";
    for (let c = 0; c < 9; ++c) {
      line += (grid[r][c] || ".") + (c % 3 === 2 && c !== 8 ? " | " : " ");
    }
    console.log(line);
    if (r % 3 === 2 && r !== 8) console.log("---------------------");
  }
}

/* run example */
if (require.main === module) {
  const solved = generateSolvedGrid();
  console.log("Solved Sudoku:");
  printGrid(solved);
}

module.exports = { generateSolvedGrid, printGrid };