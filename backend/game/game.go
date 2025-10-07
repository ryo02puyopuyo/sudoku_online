package game

import (
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/ryo02puyopuyo/sudoku_online/backend/models"
)

// Game はゲーム全体のシングルトルンな状態を管理します
type Game struct {
	mu                  sync.Mutex
	Board               [9][9]models.Cell
	Solution            [9][9]int
	Scores              models.Score
	IsOver              bool
	LastGameOverPayload *models.GameOverPayload
}

// NewGame は新しいGameインスタンスを生成して返します
func NewGame() *Game {
	g := &Game{}
	g.Reset()
	return g
}

// Reset はゲームの状態を初期化します
func (g *Game) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Scores = models.Score{Team1: 0, Team2: 0}
	g.IsOver = false
	g.LastGameOverPayload = nil

	solution, err := GenerateSolvedGrid(1000)
	if err != nil {
		log.Fatalf("Error generating grid: %v. Server cannot start.", err)
		return
	}
	g.Solution = solution
	puzzle := createPuzzleFromSolution(solution, 0.5)

	var board [9][9]models.Cell
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if puzzle[r][c] != 0 {
				board[r][c] = models.Cell{Value: puzzle[r][c], Status: "fixed"}
			} else {
				board[r][c] = models.Cell{Value: 0, Status: "empty"}
			}
		}
	}
	g.Board = board
	log.Println("A new board state has been generated and game state has been reset.")
}

// UpdateCell はセルの更新ロジックを担当し、ゲームが終了したかどうかを返します
func (g *Game) UpdateCell(row, col, value int, playerTeam int) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	currentCell := g.Board[row][col]

	// 編集不可のセルは更新しない
	if currentCell.Status == "fixed" || currentCell.Status == "correct" {
		return false
	}
	// 間違っていて同じ数字が入力された場合は何もしない
	if currentCell.Status == "wrong" && currentCell.Value == value {
		return false
	}

	// セル状態を更新
	if value == 0 {
		g.Board[row][col] = models.Cell{Value: 0, Status: "empty"}
	} else if value == g.Solution[row][col] {
		g.Board[row][col] = models.Cell{Value: value, Status: "correct", FilledByTeam: playerTeam}
		if playerTeam == 1 {
			g.Scores.Team1++
		} else {
			g.Scores.Team2++
		}
	} else {
		g.Board[row][col] = models.Cell{Value: value, Status: "wrong", FilledByTeam: playerTeam}
		if playerTeam == 1 {
			g.Scores.Team1--
		} else {
			g.Scores.Team2--
		}
	}

	// 全てのセルが埋まったかチェック
	isFull := true
	for r_check := 0; r_check < 9; r_check++ {
		for c_check := 0; c_check < 9; c_check++ {
			if g.Board[r_check][c_check].Status != "correct" && g.Board[r_check][c_check].Status != "fixed" {
				isFull = false
				break
			}
		}
	}

	if isFull {
		g.IsOver = true
		var winner int
		if g.Scores.Team1 > g.Scores.Team2 {
			winner = 1
		} else if g.Scores.Team2 > g.Scores.Team1 {
			winner = 2
		} else {
			winner = 0 // Draw
		}

		gameOverPayload := models.GameOverPayload{
			WinnerTeam:  winner,
			FinalScores: g.Scores,
		}
		g.LastGameOverPayload = &gameOverPayload
		log.Println("Game Over!")
	}

	return isFull
}

// GetBoard は現在の盤面のスナップショットを返します
func (g *Game) GetBoard() [9][9]models.Cell {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.Board
}

// GetScores は現在のスコアのスナップショットを返します
func (g *Game) GetScores() models.Score {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.Scores
}

// GetGameOverState は現在のゲームオーバー状態を返します
func (g *Game) GetGameOverState() (bool, *models.GameOverPayload) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.IsOver, g.LastGameOverPayload
}

// 解答から問題を作成
func createPuzzleFromSolution(solution [9][9]int, difficulty float64) [9][9]int {
	var puzzle [9][9]int
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if rand.Float64() < difficulty {
				puzzle[r][c] = 0
			} else {
				puzzle[r][c] = solution[r][c]
			}
		}
	}
	return puzzle
}

// 解答済みの数独グリッドを生成
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
				boxIdx := (r/3)*3 + (c / 3)
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

func shuffle(nums []int) {
	for i := len(nums) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		nums[i], nums[j] = nums[j], nums[i]
	}
}
