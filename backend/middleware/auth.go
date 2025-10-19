package middleware

import (
	"context"
	"net/http"

	"github.com/ryo02puyopuyo/sudoku_online/backend/db"
	"gorm.io/gorm"
)

type Auth struct {
	DB *gorm.DB
}

func (amw *Auth) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		ctx := r.Context()

		if err == nil && cookie.Value != "" {
			user, err := db.FindUserByToken(amw.DB, cookie.Value)
			if err == nil && user != nil {
				ctx = context.WithValue(ctx, "user", user)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
