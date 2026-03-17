package db

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

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

func Connect() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("注意: .envファイルが読み込めませんでした。環境変数を直接使用します")
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("エラー: DB_DSN が設定されていません")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false, // PgBouncer利用時はfalse必須
	})

	// コネクションプール設定
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetConnMaxLifetime(time.Minute * 3)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		return nil, fmt.Errorf("マイグレーション失敗: %w", err)
	}

	log.Println("DB接続とマイグレーション成功")
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
		if strings.Contains(result.Error.Error(), "23505") || strings.Contains(result.Error.Error(), "duplicate key") {
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
	return GenerateJWT(&user)
}

func GenerateJWT(user *User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// VerifyJWT はJWTトークンを検証し、DBアクセスなしでユーザー情報を復元する
func VerifyJWT(tokenString string) (*User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("トークンの解析に失敗しました: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &User{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		}, nil
	}

	return nil, fmt.Errorf("無効なトークンです")
}

// DeleteUserToken はJWT方式では不要（互換性のために残す）
func DeleteUserToken(db *gorm.DB, userID uint) error {
	return nil
}
