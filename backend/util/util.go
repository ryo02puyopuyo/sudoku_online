package util

import (
	"fmt"
	"math/rand"
	"time"
)

// パッケージ初期化時に一度だけ乱数のシードを設定
func init() {
	rand.Seed(time.Now().UnixNano())
}

// 0.0から1.0の間のfloat64を返す
func RandFloat() float64 {
	return rand.Float64()
}

// 座標(r, c)が属する3x3のボックス番号(0-8)を返す
func boxIndex(r, c int) int {
	return (r/3)*3 + (c / 3)
}

// intのスライスをシャッフルする
func shuffle(nums []int) {
	for i := len(nums) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		nums[i], nums[j] = nums[j], nums[i]
	}
}

// 数独の完全な盤面を生成する (バックトラッキング)
func GenerateSolvedGrid(maxAttempts int) ([9][9]int, error) {
	var grid [9][9]int
	for attempt := 0; attempt < maxAttempts; attempt++ {
		grid = [9][9]int{}
		usedRow := [9][10]bool{}
		usedCol := [9][10]bool{}
		usedBox := [9][10]bool{}

		cellList := make([][2]int, 0, 81)
		for r := 0; r < 9; r++ {
			for c := 0; c < 9; c++ {
				cellList = append(cellList, [2]int{r, c})
			}
		}

		// DFS (深さ優先探索) バックトラッキング
		var dfs func(idx int) bool
		dfs = func(idx int) bool {
			if idx == len(cellList) {
				return true // すべてのセルが埋まった
			}
			r, c := cellList[idx][0], cellList[idx][1]

			// そのセルに入れられる数字の候補
			nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
			shuffle(nums)

			for _, n := range nums {
				boxIdx := boxIndex(r, c)
				if !usedRow[r][n] && !usedCol[c][n] && !usedBox[boxIdx][n] {
					grid[r][c] = n
					usedRow[r][n], usedCol[c][n], usedBox[boxIdx][n] = true, true, true

					if dfs(idx + 1) {
						return true
					}

					// バックトラック
					grid[r][c] = 0
					usedRow[r][n], usedCol[c][n], usedBox[boxIdx][n] = false, false, false
				}
			}
			return false
		}

		if dfs(0) {
			return grid, nil // 生成成功
		}
	}
	return grid, fmt.Errorf("failed to generate solved grid after %d attempts", maxAttempts)
}