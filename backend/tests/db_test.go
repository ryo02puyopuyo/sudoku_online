package tests

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ryo02puyopuyo/sudoku_online/backend/db"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// テスト用DB接続を作成し、テーブルを初期化する
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		// デフォルト: docker-compose.yml の PostgreSQL に接続
		dsn = "host=localhost user=sudoku_user password=password dbname=sudoku_db port=5432 sslmode=disable TimeZone=Asia/Tokyo"
	}

	testDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("テスト用DB接続に失敗（Dockerが起動していない可能性あり）: %v", err)
	}

	// テスト用テーブルのマイグレーション
	if err := testDB.AutoMigrate(&db.User{}); err != nil {
		t.Fatalf("マイグレーション失敗: %v", err)
	}

	return testDB
}

// テスト後にテストユーザーをクリーンアップする
func cleanupTestUsers(t *testing.T, testDB *gorm.DB, prefix string) {
	t.Helper()
	testDB.Where("username LIKE ?", prefix+"%").Delete(&db.User{})
}

// テスト用のユニークなユーザー名を生成する
func testUsername(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func TestCreateUser_Success(t *testing.T) {
	testDB := setupTestDB(t)
	username := testUsername("test_create")
	defer cleanupTestUsers(t, testDB, "test_create")

	user, err := db.CreateUser(testDB, username, "password123", "user")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected non-zero user ID")
	}
	if user.Username != username {
		t.Errorf("expected username %q, got %q", username, user.Username)
	}
	if user.Role != "user" {
		t.Errorf("expected role 'user', got %q", user.Role)
	}
	// パスワードがハッシュ化されていること
	if user.PasswordHash == "password123" {
		t.Error("password should be hashed, not stored in plaintext")
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	testDB := setupTestDB(t)
	username := testUsername("test_dup")
	defer cleanupTestUsers(t, testDB, "test_dup")

	_, err := db.CreateUser(testDB, username, "password123", "user")
	if err != nil {
		t.Fatalf("first CreateUser failed: %v", err)
	}

	_, err = db.CreateUser(testDB, username, "password456", "user")
	if err == nil {
		t.Error("expected error for duplicate username, got nil")
	}
}

func TestLoginUser_Success(t *testing.T) {
	testDB := setupTestDB(t)
	username := testUsername("test_login")
	defer cleanupTestUsers(t, testDB, "test_login")

	_, err := db.CreateUser(testDB, username, "mypassword", "user")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	token, err := db.LoginUser(testDB, username, "mypassword")
	if err != nil {
		t.Fatalf("LoginUser failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// 返されたトークンが有効であること
	restored, err := db.VerifyJWT(token)
	if err != nil {
		t.Fatalf("VerifyJWT failed for login token: %v", err)
	}
	if restored.Username != username {
		t.Errorf("expected username %q in token, got %q", username, restored.Username)
	}
}

func TestLoginUser_WrongPassword(t *testing.T) {
	testDB := setupTestDB(t)
	username := testUsername("test_wrongpw")
	defer cleanupTestUsers(t, testDB, "test_wrongpw")

	_, err := db.CreateUser(testDB, username, "correctpassword", "user")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	_, err = db.LoginUser(testDB, username, "wrongpassword")
	if err == nil {
		t.Error("expected error for wrong password, got nil")
	}
}

func TestLoginUser_NonexistentUser(t *testing.T) {
	testDB := setupTestDB(t)

	_, err := db.LoginUser(testDB, "nonexistent_user_xyz", "password")
	if err == nil {
		t.Error("expected error for nonexistent user, got nil")
	}
}

func TestCreateUser_AdminRole(t *testing.T) {
	testDB := setupTestDB(t)
	username := testUsername("test_admin")
	defer cleanupTestUsers(t, testDB, "test_admin")

	user, err := db.CreateUser(testDB, username, "adminpass", "admin")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if user.Role != "admin" {
		t.Errorf("expected role 'admin', got %q", user.Role)
	}

	// ログインしてトークン内のロールも確認
	token, err := db.LoginUser(testDB, username, "adminpass")
	if err != nil {
		t.Fatalf("LoginUser failed: %v", err)
	}

	restored, err := db.VerifyJWT(token)
	if err != nil {
		t.Fatalf("VerifyJWT failed: %v", err)
	}
	if restored.Role != "admin" {
		t.Errorf("expected admin role in token, got %q", restored.Role)
	}
}

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	os.Exit(m.Run())
}
