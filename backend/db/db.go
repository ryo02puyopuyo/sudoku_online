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

// JWTの署名に使用する秘密鍵（環境変数から取得）
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// JWTのデータ構造（Payload）
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
		log.Println("注意: .envファイルが読み込めませんでした。、読み込めない環境の場合okです")
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("エラー: DB_DSN が設定されていません")
	}

	log.Println("db_dsn成功")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false, // プーラー（PgBouncer）利用時は必須
	})

	log.Println("gorm open 成功、DB接続開始")

	// ★ 安定化設定：ゾンビ接続を3分で強制終了してリフレッシュ
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetConnMaxLifetime(time.Minute * 3) // 270秒タイムアウトを防ぐ
	}

	// ★ Userテーブルのみマイグレーション
	err = db.AutoMigrate(&User{})
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
		if strings.Contains(result.Error.Error(), "23505") || strings.Contains(result.Error.Error(), "duplicate key") {
			return nil, fmt.Errorf("ユーザー名 '%s' は既に使用されています", username)
		}
		return nil, fmt.Errorf("ユーザー作成失敗: %w", result.Error)
	}
	return user, nil
}

// ★ ログイン：DBにトークンを保存せず、JWTを生成して返す
func LoginUser(db *gorm.DB, username, password string) (string, error) {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return "", fmt.Errorf("ユーザー名またはパスワードが違います")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("ユーザー名またはパスワードが違います")
	}

	// JWTの生成
	return GenerateJWT(&user)
}

// トークンを生成する
func GenerateJWT(user *User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)), // 3日間有効
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// トークンを検証し、DBを叩かずにユーザー情報を復元する
func VerifyJWT(tokenString string) (*User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("トークンの解析に失敗しました: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// ★ DBにはアクセスせず、トークン内の情報からUserオブジェクトを作る
		return &User{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		}, nil
	}

	return nil, fmt.Errorf("無効なトークンです")
}

// ★ DeleteUserToken は JWT 方式では不要（クライアント側でトークンを捨てればOK）
func DeleteUserToken(db *gorm.DB, userID uint) error {
	return nil // 互換性のために残すだけ
}
