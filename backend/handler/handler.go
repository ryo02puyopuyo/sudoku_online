package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ryo02puyopuyo/sudoku_online/backend/db"
	"gorm.io/gorm"
)

func CheckHandler(w http.ResponseWriter, r *http.Request, database *gorm.DB) {
	log.Println("check handler called")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// JSONで { "username": "...", "password": "..." } を受け取る
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ok, err := db.CheckUser(database, creds.Username, creds.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if ok {
		log.Println("check user successfully")
	} else {
		log.Println("check user failed")
		return
	}

	// 成功レスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

// RegisterHandler は新規ユーザー登録を処理します
func RegisterHandler(w http.ResponseWriter, r *http.Request, database *gorm.DB) {
	log.Println("register handler called")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// JSONで { "username": "...", "password": "..." } を受け取る
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	user, err := db.RegisterUser(database, creds.Username, creds.Password, "user")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("User %s registered successfully", user.Username)
	// 成功レスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request, database *gorm.DB) {
	log.Println("login handler called")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// JSONリクエストの内容を受け取ってログ出力
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	log.Println("[TEST] Received data:", data)

	// レスポンス返却
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Request received successfully",
	})
}
