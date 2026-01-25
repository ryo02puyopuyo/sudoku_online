package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ryo02puyopuyo/sudoku_online/backend/db"
	"gorm.io/gorm"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type API struct {
	DB *gorm.DB
}

type TestPayload struct {
	Msg string `json:"msg"`
}

// テスト用ハンドラ
func (a *API) TestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTメソッドのみ許可されています", http.StatusMethodNotAllowed)
		return
	}

	var payload TestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "無効なJSONです", http.StatusBadRequest)
		return
	}

	log.Printf("[TEST] フロントエンドからメッセージを受信: %s", payload.Msg)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":           "ok",
		"received_message": payload.Msg,
	}
	json.NewEncoder(w).Encode(response)
}

// ユーザー登録ハンドラ
func (a *API) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "無効なリクエスト", http.StatusBadRequest)
		return
	}

	_, err := db.CreateUser(a.DB, creds.Username, creds.Password, "user")
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "登録完了"})
}

// ログインハンドラ（JWT方式へ完全修正）
func (a *API) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "無効なリクエスト", http.StatusBadRequest)
		return
	}

	// 1. ユーザーを検証し、JWTトークンを生成
	token, err := db.LoginUser(a.DB, creds.Username, creds.Password)
	if err != nil {
		// 認証失敗時は 401 Unauthorized
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// ★ 【重要】Cookie（Set-Cookie）の処理は完全に削除しました。
	// 代わりに、JSONのレスポンスボディに token を含めてフロントエンドに直接手渡します。

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// フロントエンドの const { token } = response.data が待っているデータ
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "ログイン成功",
		"token":    token, // これが localStorage に保存されます
		"username": creds.Username,
	})
}

// 現在のユーザー情報を返すエンドポイント
func (a *API) MeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// ミドルウェア(middleware/auth.go)が JWT を解析してセットしたユーザーを取得
	user, ok := r.Context().Value("user").(*db.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "guest",
			"msg":    "ログインしていません",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "authenticated",
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
		"wins":     user.Wins,
	})
}
