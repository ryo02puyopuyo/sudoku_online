package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv" // 1. パッケージをインポート
	"golang.org/x/crypto/bcrypt"
)

// User 構造体: DBのusersテーブルに対応
type User struct {
	ID            int
	Username      string
	Role          string
	MatchesPlayed int
	Wins          int
}

func main() {

	err := godotenv.Load("../.env")
	if err != nil {
		// .envファイルがなくてもエラーにはしない。本番環境などでは環境変数で渡されるため。
		log.Println("注意: .envファイルが読み込めませんでした。")
	}
	env := os.Getenv("APP_ENV")
	dsn := os.Getenv("DB_DSN")

	fmt.Printf("現在の環境: %s\n", env)
	// fmt.Printf("読み込んだDSN: %s\n", dsn) // デバッグ用に表示

	if dsn == "" {
		log.Fatal("エラー: 環境変数 DB_DSN が設定されていません。")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("データベースのオープンに失敗しました: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("データベースへの接続に失敗しました: %v", err)
	}

	fmt.Println("Docker上のMySQLサーバーへの接続に成功しました！🐳")

	if env == "development" {
		fmt.Println("これは開発環境用の処理です。")

		// --- 2. 一連の操作を実行 ---
		username := "Alice"
		password := "secure_password123"

		// ユーザー作成
		fmt.Printf("\n--- ユーザー '%s' を作成します ---\n", username)
		_, err = createUser(db, username, password, "user")
		if err != nil {
			log.Printf("ユーザー作成エラー: %v\n", err)
		} else {
			fmt.Printf("ユーザー '%s' を作成しました。\n", username)
		}

		// ユーザー情報を表示
		printUserDetails(db, username)

		// 試合結果を記録 (勝利)
		fmt.Printf("\n--- ユーザー '%s' の試合結果（勝利）を記録します ---\n", username)
		recordMatchResult(db, username, true)
		printUserDetails(db, username)

		// 試合結果を記録 (敗北)
		fmt.Printf("\n--- ユーザー '%s' の試合結果（敗北）を記録します ---\n", username)
		recordMatchResult(db, username, false)
		printUserDetails(db, username)

		// ユーザー削除
		//fmt.Printf("\n--- ユーザー '%s' を削除します ---\n", username)
		//err = deleteUser(db, username)
		//if err != nil {
		//	log.Printf("ユーザー削除エラー: %v", err)
		//} else {
		//	fmt.Printf("ユーザー '%s' を削除しました。\n", username)
		//}
		printUserDetails(db, username)

	}
}

// createUser: 新しいユーザーを作成し、パスワードをハッシュ化して保存する
func createUser(db *sql.DB, username, password, role string) (int64, error) {
	// パスワードをbcryptでハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("パスワードのハッシュ化に失敗: %w", err)
	}

	// DBに挿入
	result, err := db.Exec(
		"INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)",
		username, string(hashedPassword), role,
	)
	if err != nil {
		return 0, fmt.Errorf("ユーザーの挿入に失敗: %w", err)
	}

	return result.LastInsertId()
}

// deleteUser: 指定されたユーザー名のユーザーを削除する
func deleteUser(db *sql.DB, username string) error {
	_, err := db.Exec("DELETE FROM users WHERE username = ?", username)
	if err != nil {
		return fmt.Errorf("ユーザーの削除に失敗: %w", err)
	}
	return nil
}

// recordMatchResult: 試合結果を記録し、matches_playedとwinsを更新する
func recordMatchResult(db *sql.DB, username string, didWin bool) error {
	winValue := 0
	if didWin {
		winValue = 1
	}

	// 試合数(matches_played)は常に+1、勝利数(wins)は勝った場合のみ+1する
	_, err := db.Exec(
		"UPDATE users SET matches_played = matches_played + 1, wins = wins + ? WHERE username = ?",
		winValue, username,
	)
	if err != nil {
		return fmt.Errorf("試合結果の更新に失敗: %w", err)
	}
	return nil
}

// printUserDetails: ユーザーの詳細情報を取得して表示するヘルパー関数
func printUserDetails(db *sql.DB, username string) {
	var u User
	err := db.QueryRow(
		"SELECT id, username, role, matches_played, wins FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.Role, &u.MatchesPlayed, &u.Wins)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("ユーザー '%s' は見つかりませんでした。\n", username)
		} else {
			log.Printf("ユーザー情報の取得エラー: %v", err)
		}
		return
	}
	fmt.Printf("現在の状態 -> ID: %d, Name: %s, Role: %s, Played: %d, Wins: %d\n",
		u.ID, u.Username, u.Role, u.MatchesPlayed, u.Wins)
}
