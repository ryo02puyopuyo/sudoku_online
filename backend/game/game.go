package game

import (
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/ryo02puyopuyo/sudoku_online/backend/models"
)

type Game struct {
	mu                  sync.Mutex
	Board               [9][9]models.Cell
	Solution            [9][9]int
	Scores              models.Score
	IsOver              bool
	LastGameOverPayload *models.GameOverPayload
}

// セル更新の結果
type UpdateResult int

const (
	ResultNone      UpdateResult = iota // 変化なし（fixedマスなど）
	ResultCorrect                       // 正解
	ResultIncorrect                     // 不正解
	ResultHotSpot                       // ホットスポット正解
	ResultEmpty                         // マスを空にした
)

func NewGame() *Game {
	g := &Game{}
	g.Reset()
	return g
}

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
	var emptyCellCoords [][2]int

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if puzzle[r][c] != 0 {
				board[r][c] = models.Cell{Value: puzzle[r][c], Status: "fixed", IsHotSpot: false}
			} else {
				board[r][c] = models.Cell{Value: 0, Status: "empty", IsHotSpot: false}
				emptyCellCoords = append(emptyCellCoords, [2]int{r, c})
			}
		}
	}

	// ホットスポットをランダムに3か所設定
	rand.Shuffle(len(emptyCellCoords), func(i, j int) {
		emptyCellCoords[i], emptyCellCoords[j] = emptyCellCoords[j], emptyCellCoords[i]
	})
	for i := 0; i < 3 && i < len(emptyCellCoords); i++ {
		r, c := emptyCellCoords[i][0], emptyCellCoords[i][1]
		board[r][c].IsHotSpot = true
	}

	g.Board = board
	log.Println("A new board state has been generated and game state has been reset.")
}

// UpdateCell はセルの更新を処理し、(盤面完成, ホットスポットヒット) を返す
func (g *Game) UpdateCell(row, col, value int, playerTeam int) (bool, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	currentCell := g.Board[row][col]
	isHotSpotHit := false

	if currentCell.Status == "fixed" || currentCell.Status == "correct" {
		return false, false
	}
	if currentCell.Status == "wrong" && currentCell.Value == value {
		return false, false
	}

	if value == 0 {
		g.Board[row][col] = models.Cell{Value: 0, Status: "empty"}
	} else if value == g.Solution[row][col] {
		g.Board[row][col] = models.Cell{Value: value, Status: "correct", FilledByTeam: playerTeam, IsHotSpot: currentCell.IsHotSpot}

		points := 1
		if currentCell.IsHotSpot {
			isHotSpotHit = true
			points = 3
		}
		if playerTeam == 1 {
			g.Scores.Team1 += points
		} else {
			g.Scores.Team2 += points
		}
	} else {
		g.Board[row][col] = models.Cell{Value: value, Status: "wrong", FilledByTeam: playerTeam}
		if playerTeam == 1 {
			g.Scores.Team1--
		} else {
			g.Scores.Team2--
		}
	}

	// 全セルが埋まったかチェック
	isFull := true
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g.Board[r][c].Status != "correct" && g.Board[r][c].Status != "fixed" {
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
			winner = 0
		}

		gameOverPayload := models.GameOverPayload{
			WinnerTeam:  winner,
			FinalScores: g.Scores,
		}
		g.LastGameOverPayload = &gameOverPayload
		log.Println("Game Over!")
	}

	return isFull, isHotSpotHit
}

func (g *Game) SetScore(team int, points int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	switch team {
	case 1:
		g.Scores.Team1 = points
	case 2:
		g.Scores.Team2 = points
	default:
		log.Printf("Invalid team number: %d", team)
	}
}

func (g *Game) GetBoard() [9][9]models.Cell {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.Board
}

func (g *Game) GetScores() models.Score {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.Scores
}

func (g *Game) GetGameOverState() (bool, *models.GameOverPayload) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.IsOver, g.LastGameOverPayload
}

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

// GenerateSolvedGrid はバックトラッキングで有効な数独解答グリッドを生成する
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
