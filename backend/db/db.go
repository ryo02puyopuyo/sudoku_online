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

	sqlmysql "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID            uint   `gorm:"primarykey"`
	Username      string `gorm:"unique;not null"`
	PasswordHash  string `gorm:"not null"`
	Role          string `gorm:"not null;default:'user'"`
	MatchesPlayed int    `gorm:"default:0"`
	Wins          int    `gorm:"default:0"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UserToken struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"not null"`
	TokenHash string    `gorm:"unique;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func Connect() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("注意: .envファイルが読み込めませんでした。")
	}
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("エラー: DB_DSN が設定されていません")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("DB接続失敗: %w", err)
	}
	err = db.AutoMigrate(&User{}, &UserToken{})
	if err != nil {
		return nil, fmt.Errorf("マイグレーション失敗: %w", err)
	}
	fmt.Println("DB接続とマイグレーション成功")
	return db, nil
}

func CreateUser(db *gorm.DB, username, password, role string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ハッシュ化失敗: %w", err)
	}
	user := &User{Username: username, PasswordHash: string(hashedPassword), Role: role}
	result := db.Create(user)
	if result.Error != nil {
		var mysqlErr *sqlmysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, fmt.Errorf("ユーザー名 '%s' は既に使用されています", username)
		}
		return nil, fmt.Errorf("ユーザー作成失敗: %w", result.Error)
	}
	return user, nil
}

func LoginUser(db *gorm.DB, username, password string) (string, error) {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return "", fmt.Errorf("ユーザー名またはパスワードが違います")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("ユーザー名またはパスワードが違います")
	}
	token, err := generateSecureToken(32)
	if err != nil {
		return "", fmt.Errorf("トークン生成失敗: %w", err)
	}
	tokenHash := hashToken(token)
	userToken := UserToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := db.Create(&userToken).Error; err != nil {
		return "", fmt.Errorf("トークン保存失敗: %w", err)
	}
	return token, nil
}

func FindUserByToken(db *gorm.DB, token string) (*User, error) {
	tokenHash := hashToken(token)
	var userToken UserToken
	err := db.Where("token_hash = ?", tokenHash).First(&userToken).Error
	if err != nil {
		return nil, fmt.Errorf("無効なトークンです")
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
