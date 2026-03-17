package tests

import (
	"os"
	"testing"

	"github.com/ryo02puyopuyo/sudoku_online/backend/db"
)

func init() {
	// テスト時にJWT_SECRETが未設定の場合、テスト用の値を設定
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test_secret_key_for_unit_tests")
	}
}

func TestGenerateAndVerifyJWT_Roundtrip(t *testing.T) {
	user := &db.User{ID: 42, Username: "testuser", Role: "user"}

	token, err := db.GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateJWT returned empty token")
	}

	restored, err := db.VerifyJWT(token)
	if err != nil {
		t.Fatalf("VerifyJWT failed: %v", err)
	}

	if restored.ID != user.ID {
		t.Errorf("expected ID %d, got %d", user.ID, restored.ID)
	}
	if restored.Username != user.Username {
		t.Errorf("expected Username %q, got %q", user.Username, restored.Username)
	}
	if restored.Role != user.Role {
		t.Errorf("expected Role %q, got %q", user.Role, restored.Role)
	}
}

func TestGenerateJWT_AdminRole(t *testing.T) {
	user := &db.User{ID: 1, Username: "admin", Role: "admin"}

	token, err := db.GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	restored, err := db.VerifyJWT(token)
	if err != nil {
		t.Fatalf("VerifyJWT failed: %v", err)
	}

	if restored.Role != "admin" {
		t.Errorf("expected admin role, got %q", restored.Role)
	}
}

func TestVerifyJWT_InvalidToken(t *testing.T) {
	_, err := db.VerifyJWT("invalid.token.string")
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}

func TestVerifyJWT_EmptyToken(t *testing.T) {
	_, err := db.VerifyJWT("")
	if err == nil {
		t.Error("expected error for empty token, got nil")
	}
}

func TestGenerateJWT_DifferentUsersProduceDifferentTokens(t *testing.T) {
	user1 := &db.User{ID: 1, Username: "alice", Role: "user"}
	user2 := &db.User{ID: 2, Username: "bob", Role: "user"}

	token1, err := db.GenerateJWT(user1)
	if err != nil {
		t.Fatalf("GenerateJWT for user1 failed: %v", err)
	}
	token2, err := db.GenerateJWT(user2)
	if err != nil {
		t.Fatalf("GenerateJWT for user2 failed: %v", err)
	}

	if token1 == token2 {
		t.Error("expected different tokens for different users")
	}
}

func TestDeleteUserToken_ReturnsNil(t *testing.T) {
	err := db.DeleteUserToken(nil, 1)
	if err != nil {
		t.Errorf("DeleteUserToken should return nil, got: %v", err)
	}
}
