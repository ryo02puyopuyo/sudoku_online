package tests

import (
	"testing"

	"github.com/ryo02puyopuyo/sudoku_online/backend/game"
)

func TestGenerateSolvedGrid_ProducesValidGrid(t *testing.T) {
	grid, err := game.GenerateSolvedGrid(100)
	if err != nil {
		t.Fatalf("GenerateSolvedGrid returned error: %v", err)
	}

	for r := 0; r < 9; r++ {
		seen := [10]bool{}
		for c := 0; c < 9; c++ {
			v := grid[r][c]
			if v < 1 || v > 9 {
				t.Errorf("row %d, col %d: invalid value %d", r, c, v)
			}
			if seen[v] {
				t.Errorf("row %d: duplicate value %d", r, v)
			}
			seen[v] = true
		}
	}

	for c := 0; c < 9; c++ {
		seen := [10]bool{}
		for r := 0; r < 9; r++ {
			v := grid[r][c]
			if seen[v] {
				t.Errorf("col %d: duplicate value %d", c, v)
			}
			seen[v] = true
		}
	}

	for boxR := 0; boxR < 3; boxR++ {
		for boxC := 0; boxC < 3; boxC++ {
			seen := [10]bool{}
			for r := boxR * 3; r < boxR*3+3; r++ {
				for c := boxC * 3; c < boxC*3+3; c++ {
					v := grid[r][c]
					if seen[v] {
						t.Errorf("box (%d,%d): duplicate value %d", boxR, boxC, v)
					}
					seen[v] = true
				}
			}
		}
	}
}

func TestGenerateSolvedGrid_FailsWithZeroAttempts(t *testing.T) {
	_, err := game.GenerateSolvedGrid(0)
	if err == nil {
		t.Error("expected error with 0 attempts, got nil")
	}
}

func TestNewGame_InitializesCorrectly(t *testing.T) {
	g := game.NewGame()

	scores := g.GetScores()
	if scores.Team1 != 0 || scores.Team2 != 0 {
		t.Errorf("expected initial scores (0, 0), got (%d, %d)", scores.Team1, scores.Team2)
	}

	isOver, payload := g.GetGameOverState()
	if isOver {
		t.Error("expected game not over initially")
	}
	if payload != nil {
		t.Error("expected nil gameOverPayload initially")
	}

	board := g.GetBoard()
	hasFixed := false
	hasEmpty := false
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board[r][c].Status == "fixed" {
				hasFixed = true
			}
			if board[r][c].Status == "empty" {
				hasEmpty = true
			}
		}
	}
	if !hasFixed {
		t.Error("expected at least one fixed cell")
	}
	if !hasEmpty {
		t.Error("expected at least one empty cell")
	}
}

func TestNewGame_HasHotSpots(t *testing.T) {
	g := game.NewGame()
	board := g.GetBoard()

	hotSpotCount := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board[r][c].IsHotSpot {
				hotSpotCount++
			}
		}
	}

	if hotSpotCount != 3 {
		t.Errorf("expected 3 hot spots, got %d", hotSpotCount)
	}
}

func TestUpdateCell_FixedCellIsNotUpdated(t *testing.T) {
	g := game.NewGame()
	board := g.GetBoard()

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board[r][c].Status == "fixed" {
				completed, _ := g.UpdateCell(r, c, 5, 1)
				if completed {
					t.Error("fixed cell should not cause board completion")
				}
				newBoard := g.GetBoard()
				if newBoard[r][c].Value != board[r][c].Value {
					t.Error("fixed cell value should not change")
				}
				return
			}
		}
	}
	t.Fatal("no fixed cell found for testing")
}

func TestUpdateCell_CorrectAnswer(t *testing.T) {
	g := game.NewGame()
	board := g.GetBoard()

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board[r][c].Status == "empty" && !board[r][c].IsHotSpot {
				correctValue := g.Solution[r][c]
				g.UpdateCell(r, c, correctValue, 1)

				newBoard := g.GetBoard()
				if newBoard[r][c].Status != "correct" {
					t.Errorf("expected status 'correct', got '%s'", newBoard[r][c].Status)
				}
				if newBoard[r][c].Value != correctValue {
					t.Errorf("expected value %d, got %d", correctValue, newBoard[r][c].Value)
				}
				if newBoard[r][c].FilledByTeam != 1 {
					t.Errorf("expected filledByTeam 1, got %d", newBoard[r][c].FilledByTeam)
				}

				scores := g.GetScores()
				if scores.Team1 != 1 {
					t.Errorf("expected Team1 score 1, got %d", scores.Team1)
				}
				return
			}
		}
	}
}

func TestUpdateCell_WrongAnswer(t *testing.T) {
	g := game.NewGame()
	board := g.GetBoard()

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board[r][c].Status == "empty" {
				correctValue := g.Solution[r][c]
				wrongValue := (correctValue % 9) + 1

				g.UpdateCell(r, c, wrongValue, 2)

				newBoard := g.GetBoard()
				if newBoard[r][c].Status != "wrong" {
					t.Errorf("expected status 'wrong', got '%s'", newBoard[r][c].Status)
				}

				scores := g.GetScores()
				if scores.Team2 != -1 {
					t.Errorf("expected Team2 score -1, got %d", scores.Team2)
				}
				return
			}
		}
	}
}

func TestUpdateCell_ClearCell(t *testing.T) {
	g := game.NewGame()
	board := g.GetBoard()

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board[r][c].Status == "empty" {
				correctValue := g.Solution[r][c]
				wrongValue := (correctValue % 9) + 1
				g.UpdateCell(r, c, wrongValue, 1)

				g.UpdateCell(r, c, 0, 1)

				newBoard := g.GetBoard()
				if newBoard[r][c].Status != "empty" {
					t.Errorf("expected status 'empty' after clear, got '%s'", newBoard[r][c].Status)
				}
				if newBoard[r][c].Value != 0 {
					t.Errorf("expected value 0 after clear, got %d", newBoard[r][c].Value)
				}
				return
			}
		}
	}
}

func TestSetScore(t *testing.T) {
	g := game.NewGame()

	g.SetScore(1, 100)
	g.SetScore(2, 50)

	scores := g.GetScores()
	if scores.Team1 != 100 {
		t.Errorf("expected Team1 score 100, got %d", scores.Team1)
	}
	if scores.Team2 != 50 {
		t.Errorf("expected Team2 score 50, got %d", scores.Team2)
	}
}

func TestReset_ClearsGameState(t *testing.T) {
	g := game.NewGame()

	g.SetScore(1, 50)
	g.SetScore(2, 30)
	g.Reset()

	scores := g.GetScores()
	if scores.Team1 != 0 || scores.Team2 != 0 {
		t.Errorf("expected scores (0, 0) after reset, got (%d, %d)", scores.Team1, scores.Team2)
	}

	isOver, _ := g.GetGameOverState()
	if isOver {
		t.Error("expected game not over after reset")
	}
}
