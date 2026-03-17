package tests

import (
	"encoding/json"
	"testing"

	"github.com/ryo02puyopuyo/sudoku_online/backend/models"
)

func TestPlayerJSON(t *testing.T) {
	p := models.Player{ID: "user-1", Name: "TestUser", Team: 1, Role: "admin"}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("failed to marshal Player: %v", err)
	}

	var decoded models.Player
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Player: %v", err)
	}

	if decoded.ID != p.ID || decoded.Name != p.Name || decoded.Team != p.Team || decoded.Role != p.Role {
		t.Errorf("Player roundtrip mismatch: got %+v, want %+v", decoded, p)
	}
}

func TestCellJSON(t *testing.T) {
	c := models.Cell{Value: 5, Status: "correct", FilledByTeam: 1, IsHotSpot: true}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("failed to marshal Cell: %v", err)
	}

	var decoded models.Cell
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Cell: %v", err)
	}

	if decoded != c {
		t.Errorf("Cell roundtrip mismatch: got %+v, want %+v", decoded, c)
	}
}

func TestScoreJSON(t *testing.T) {
	s := models.Score{Team1: 10, Team2: -3}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal Score: %v", err)
	}

	var decoded models.Score
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Score: %v", err)
	}

	if decoded != s {
		t.Errorf("Score roundtrip mismatch: got %+v, want %+v", decoded, s)
	}
}

func TestGameOverPayloadJSON(t *testing.T) {
	p := models.GameOverPayload{
		WinnerTeam:  2,
		FinalScores: models.Score{Team1: 5, Team2: 8},
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("failed to marshal GameOverPayload: %v", err)
	}

	var decoded models.GameOverPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal GameOverPayload: %v", err)
	}

	if decoded.WinnerTeam != p.WinnerTeam || decoded.FinalScores != p.FinalScores {
		t.Errorf("GameOverPayload roundtrip mismatch: got %+v, want %+v", decoded, p)
	}
}

func TestChatMessageJSON(t *testing.T) {
	m := models.ChatMessage{
		SenderName: "Alice",
		SenderTeam: 1,
		Message:    "Hello!",
		Timestamp:  "12:00",
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal ChatMessage: %v", err)
	}

	var decoded models.ChatMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ChatMessage: %v", err)
	}

	if decoded != m {
		t.Errorf("ChatMessage roundtrip mismatch: got %+v, want %+v", decoded, m)
	}
}

func TestServerMessageJSON(t *testing.T) {
	sm := models.ServerMessage{Type: "board_state", Payload: "test"}

	data, err := json.Marshal(sm)
	if err != nil {
		t.Fatalf("failed to marshal ServerMessage: %v", err)
	}

	var decoded models.ServerMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ServerMessage: %v", err)
	}

	if decoded.Type != sm.Type {
		t.Errorf("ServerMessage type mismatch: got %s, want %s", decoded.Type, sm.Type)
	}
}

func TestCellJSON_EmptyCell(t *testing.T) {
	c := models.Cell{Value: 0, Status: "empty", FilledByTeam: 0, IsHotSpot: false}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("failed to marshal Cell: %v", err)
	}

	var decoded models.Cell
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Cell: %v", err)
	}

	if decoded != c {
		t.Errorf("empty Cell roundtrip mismatch: got %+v, want %+v", decoded, c)
	}
}
