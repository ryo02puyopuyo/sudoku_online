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
func SetupDatabase() (*gorm.DB, error) {
	err := godotenv.Load("./.env")
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

// 開発用のテスト関数

// --- データベース操作関数 ---

func RegisterUser(db *gorm.DB, username, password, role string) (*User, error) {
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

func LoginUser(db *gorm.DB, username, password string) (string, error) {
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

func LogoutUser(db *gorm.DB, token string) error {
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

func FindUserByToken(db *gorm.DB, token string) (*User, error) {
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

func DeleteUser(db *gorm.DB, username string) error {
	result := db.Unscoped().Where("username = ?", username).Delete(&User{})
	if result.Error != nil {
		return fmt.Errorf("ユーザーの削除に失敗: %w", result.Error)
	}
	return nil
}

func PrintUserDetails(db *gorm.DB, username string) {
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
func RecordMatchResult(db *gorm.DB, username string, didWin bool) error {
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

// CheckUser はユーザー認証を行います (トークン発行なし)。
func CheckUser(db *gorm.DB, username, password string) (bool, error) {
	var user User
	// 1. ユーザー名でユーザーを取得
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ユーザーが見つからない場合は、セキュリティのため「ユーザー名またはパスワードが違います」と返す
			return false, fmt.Errorf("ユーザー名またはパスワードが違います")
		}
		// その他のDBエラー
		return false, fmt.Errorf("データベースエラー: %w", err)
	}

	// 2. パスワードのハッシュを比較
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			// パスワードが一致しない場合も、同様に「ユーザー名またはパスワードが違います」と返す
			return false, fmt.Errorf("ユーザー名またはパスワードが違います")
		}
		// その他のbcryptエラー
		return false, fmt.Errorf("認証エラー: %w", err)
	}

	// 3. 認証成功
	return true, nil
}
