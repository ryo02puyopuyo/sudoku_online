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

function generateSolvedGrid(maxAttempts = 10) {
  for (let attempt = 0; attempt < maxAttempts; ++attempt) {
    const grid = makeEmptyGrid();
    const usedRow = Array.from({ length: 9 }, () => Array(10).fill(false));
    const usedCol = Array.from({ length: 9 }, () => Array(10).fill(false));
    const usedBox = Array.from({ length: 9 }, () => Array(10).fill(false));

    const nums = shuffle([1,2,3,4,5,6,7,8,9].slice());
    for (let c = 0; c < 9; ++c) {
      const v = nums[c];
      grid[0][c] = v;
      usedRow[0][v] = true;
      usedCol[c][v] = true;
      usedBox[boxIndex(0,c)][v] = true;
    }

    const cellList = [];
    for (let r = 0; r < 9; ++r)
      for (let c = 0; c < 9; ++c)
        if (grid[r][c] === 0) cellList.push([r,c]);

    let solved = false;
    function dfs(idx) {
      if (idx === cellList.length) { solved = true; return true; }
      const [r,c] = cellList[idx];
      const cand = [];
      for (let n = 1; n <= 9; ++n)
        if (!usedRow[r][n] && !usedCol[c][n] && !usedBox[boxIndex(r,c)][n])
          cand.push(n);
      shuffle(cand);
      for (const n of cand) {
        grid[r][c] = n; usedRow[r][n]=true; usedCol[c][n]=true; usedBox[boxIndex(r,c)][n]=true;
        if (dfs(idx+1)) return true;
        grid[r][c]=0; usedRow[r][n]=false; usedCol[c][n]=false; usedBox[boxIndex(r,c)][n]=false;
      }
      return false;
    }
    if (dfs(0)) return grid;
  }
  throw new Error("Failed to generate solved grid");
}

module.exports = { generateSolvedGrid };
