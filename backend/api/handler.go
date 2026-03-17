
package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ryo02puyopuyo/sudoku_online/backend/db"
	"github.com/ryo02puyopuyo/sudoku_online/backend/models"
	"github.com/ryo02puyopuyo/sudoku_online/backend/room"
	"gorm.io/gorm"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type API struct {
	DB          *gorm.DB
	RoomManager *room.RoomManager
}

type TestPayload struct {
	Msg string `json:"msg"`
}

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

func (a *API) RoomListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	rooms := a.RoomManager.GetAllRoomInfo()
	payload := models.RoomListPayload{
		Rooms: rooms,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "ログイン成功",
		"token":    token,
		"username": creds.Username,
	})
}

func (a *API) MeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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
