package util

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandFloat() float64 {
	return rand.Float64()
}

func boxIndex(r, c int) int {
	return (r/3)*3 + (c / 3)
}

func shuffle(nums []int) {
	for i := len(nums) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		nums[i], nums[j] = nums[j], nums[i]
	}
}

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
		var dfs func(idx int) bool
		dfs = func(idx int) bool {
			if idx == len(cellList) {
				return true
			}
			r, c := cellList[idx][0], cellList[idx][1]
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
					grid[r][c] = 0
					usedRow[r][n], usedCol[c][n], usedBox[boxIdx][n] = false, false, false
				}
			}
			return false
		}
		if dfs(0) {
			return grid, nil
		}
	}
	return grid, fmt.Errorf("failed to generate solved grid after %d attempts", maxAttempts)
}