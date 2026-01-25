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

// Optional はユーザーがいればContextにセットし、いなくても次の処理へ進めます
func (amw *Auth) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// 1. HTTPヘッダーからトークンを取得 (API用)
		// 形式: Authorization: Bearer <token>
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. ヘッダーにない場合、URLパラメータから取得 (WebSocket接続用)
		// 形式: ws://.../ws?token=<token>
		if tokenString == "" {
			tokenString = r.URL.Query().Get("token")
		}

		ctx := r.Context()

		if tokenString != "" {
			// ★ db.VerifyJWT を呼び出す (DBアクセスなし)
			user, err := db.VerifyJWT(tokenString)
			if err == nil && user != nil {
				// 有効なトークンならContextにユーザー情報を入れる
				ctx = context.WithValue(ctx, "user", user)
			}
			// トークンが無効な場合は、ログを出さずにそのまま(ゲスト扱い)進めます
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
