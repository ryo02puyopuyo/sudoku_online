package db

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	sqlmysql "github.com/go-sql-driver/mysql" // エイリアス(別名)を付けます
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql" // こちらは 'mysql' のまま使います
	"gorm.io/gorm"
)

// User 構造体: 論理削除用のDeletedAtフィールドを削除
type User struct {
	ID            uint   `gorm:"primarykey"`
	Username      string `gorm:"unique;not null"`
	PasswordHash  string `gorm:"not null"`
	Role          string `gorm:"not null;default:'user'"`
	MatchesPlayed int    `gorm:"default:0"`
	Wins          int    `gorm:"default:0"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	UserTokens    []UserToken
}

// IsAdmin はユーザーが管理者かどうかを判定します
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// UserToken 構造体: 論理削除用のDeletedAtフィールドを削除
type UserToken struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"not null"`
	TokenHash string    `gorm:"unique;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// --- ヘルパー関数 ---

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// setupDatabase はDBへの接続を確立し、マイグレーションを行います
func setupDatabase() (*gorm.DB, error) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("注意: .envファイルが読み込めませんでした。")
	}
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("エラー: 環境変数 DB_DSN が設定されていません")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("データベースへの接続に失敗しました: %w", err)
	}
	fmt.Println("Docker上のMySQLサーバーへの接続に成功しました！🐳")

	// AutoMigrateは、DeletedAtフィールドがなければ論理削除用のカラムを作成しません
	err = db.AutoMigrate(&User{}, &UserToken{})
	if err != nil {
		return nil, fmt.Errorf("テーブルのマイグレーションに失敗しました: %w", err)
	}
	fmt.Println("テーブルのマイグレーションが完了しました。")

	return db, nil
}

func main() {
	db, err := setupDatabase()
	if err != nil {
		log.Fatal(err)
	}

	if os.Getenv("APP_ENV") == "development" {
		fmt.Println("これは開発環境用の処理です。")
		// 管理者ユーザーを作成（既に存在する場合はエラーが出ますが、テスト用なので問題ありません）
		//_, err := createUser(db, "admin", "admin", "admin")
		//if err != nil {
		//	fmt.Println(err) // エラー内容を表示
		//} else {
		//	fmt.Println("管理者ユーザー 'admin' を作成しました。")
		//}
		//runDevelopmentTests(db)

		recordMatchResult(db, "admin", false) // 例: admin の試合結果を勝利として記録
	}
}

// 開発用のテスト関数
func runDevelopmentTests(db *gorm.DB) {
	testUsername := "hard_delete_test"
	testPassword := "password123"

	// 事前にユーザーを削除
	deleteUser(db, testUsername)

	// 1. ユーザーを作成
	fmt.Printf("\n--- ユーザー '%s' を作成します ---\n", testUsername)
	_, err := createUser(db, testUsername, testPassword, "user")
	if err != nil {
		log.Fatalf("ユーザー作成が予期せず失敗しました: %v", err)
	}
	fmt.Println("ユーザー作成成功！")
	printUserDetails(db, testUsername)

	// 2. ユーザーを物理削除
	fmt.Printf("\n--- ユーザー '%s' を物理削除します ---\n", testUsername)
	err = deleteUser(db, testUsername)
	if err != nil {
		log.Fatalf("ユーザー削除が予期せず失敗しました: %v", err)
	}
	fmt.Println("ユーザー削除成功！")

	// 3. 削除されたことを確認（見つからないはず）
	printUserDetails(db, testUsername)
}

// --- データベース操作関数 ---

func createUser(db *gorm.DB, username, password, role string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("パスワードのハッシュ化に失敗: %w", err)
	}
	user := &User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         role,
	}
	result := db.Create(user)
	if result.Error != nil {
		var mysqlErr *sqlmysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, fmt.Errorf("ユーザー名 '%s' は既に使用されています", username)
		}
		return nil, fmt.Errorf("ユーザーの作成に失敗: %w", result.Error)
	}
	return user, nil
}

func loginUser(db *gorm.DB, username, password string) (string, error) {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return "", fmt.Errorf("ユーザーが見つからないか、パスワードが違います")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("ユーザーが見つからないか、パスワードが違います")
	}
	token, err := generateSecureToken(32)
	if err != nil {
		return "", fmt.Errorf("トークンの生成に失敗: %w", err)
	}
	tokenHash := hashToken(token)
	userToken := UserToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := db.Create(&userToken).Error; err != nil {
		return "", fmt.Errorf("トークンの保存に失敗: %w", err)
	}
	return token, nil
}

func logoutUser(db *gorm.DB, token string) error {
	tokenHash := hashToken(token)
	result := db.Where("token_hash = ?", tokenHash).Delete(&UserToken{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("無効なトークンです")
	}
	return nil
}

func findUserByToken(db *gorm.DB, token string) (*User, error) {
	tokenHash := hashToken(token)
	var userToken UserToken
	err := db.Where("token_hash = ?", tokenHash).First(&userToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("無効なトークンです")
		}
		return nil, err
	}
	if time.Now().After(userToken.ExpiresAt) {
		db.Delete(&userToken)
		return nil, fmt.Errorf("トークンの有効期限が切れています")
	}
	var user User
	if err := db.First(&user, userToken.UserID).Error; err != nil {
		return nil, fmt.Errorf("トークンに紐づくユーザーが見つかりません")
	}
	return &user, nil
}

func deleteUser(db *gorm.DB, username string) error {
	result := db.Unscoped().Where("username = ?", username).Delete(&User{})
	if result.Error != nil {
		return fmt.Errorf("ユーザーの削除に失敗: %w", result.Error)
	}
	return nil
}

func printUserDetails(db *gorm.DB, username string) {
	var u User
	err := db.Where("username = ?", username).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("ユーザー '%s' は見つかりませんでした。\n", username)
		} else {
			log.Printf("ユーザー情報の取得エラー: %v", err)
		}
		return
	}
	fmt.Printf("現在の状態 -> ID: %d, Name: %s, Role: %s, Played: %d, Wins: %d\n",
		u.ID, u.Username, u.Role, u.MatchesPlayed, u.Wins)
}

// ▼▼▼ これがご要望の、特定のユーザーの戦績を更新する関数です ▼▼▼
func recordMatchResult(db *gorm.DB, username string, didWin bool) error {
	var user User
	// まず、更新対象のユーザーをユーザー名でDBから取得
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return fmt.Errorf("ユーザー '%s' が見つかりませんでした: %w", username, err)
	}

	// Goのコード上で値を更新
	user.MatchesPlayed++ // 試合数は必ず+1
	if didWin {
		user.Wins++ // 勝った場合のみ勝利数を+1
	}

	// 更新された内容をDBに保存
	if err := db.Save(&user).Error; err != nil {
		return fmt.Errorf("試合結果の更新に失敗: %w", err)
	}
	return nil
}
