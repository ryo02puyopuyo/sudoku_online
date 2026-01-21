package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

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

func (a *API) TestHandler(w http.ResponseWriter, r *http.Request) {
	// 1. POSTメソッド以外は受け付けない
	if r.Method != http.MethodPost {
		http.Error(w, "POSTメソッドのみ許可されています", http.StatusMethodNotAllowed)
		return
	}

	// 2. フロントエンドから送られてきたJSONデータをデコード
	var payload TestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "無効なJSONです", http.StatusBadRequest)
		return
	}

	// 3. 受け取ったデータをサーバーのコンソールにログとして出力
	log.Printf("[TEST] フロントエンドからメッセージを受信: %s", payload.Msg)

	// 4. フロントエンドに応答を返す
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":           "ok",
		"received_message": payload.Msg, // 受け取ったメッセージをそのまま返す
	}
	json.NewEncoder(w).Encode(response)
}

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
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "登録完了"})
}

func (a *API) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "無効なリクエスト", http.StatusBadRequest)
		return
	}
	token, err := db.LoginUser(a.DB, creds.Username, creds.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "ログイン成功"})
}

// 【修正点】現在のユーザー情報を返すエンドポイント
func (a *API) MeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// コンテキストからユーザー情報を取得
	user, ok := r.Context().Value("user").(*db.User)
	if !ok {
		// 未ログインの場合
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "guest",
			"msg":    "ログインしていません",
		})
		return
	}

	// ログイン済みの場合はユーザー詳細を返す
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "authenticated",
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
		"wins":     user.Wins, // 現在の勝利数など
	})
}
