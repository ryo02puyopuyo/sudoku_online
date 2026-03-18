package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ryo02puyopuyo/sudoku_online/backend/db"
	"gorm.io/gorm"
)

type Auth struct {
	DB *gorm.DB
}

// Optional はトークンがあればContextにユーザーをセットし、なくてもゲストとして処理を続行する
func (amw *Auth) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// HTTPヘッダーから取得 (Authorization: Bearer <token>)
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// ヘッダーにない場合、URLパラメータから取得 (WebSocket接続用)
		if tokenString == "" {
			tokenString = r.URL.Query().Get("token")
		}

		ctx := r.Context()

		if tokenString != "" {
			user, err := db.VerifyJWT(tokenString)
			if err == nil && user != nil {
				ctx = context.WithValue(ctx, "user", user)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
