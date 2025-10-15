package main

import (
	"log"
	"net/http"
	"encoding/json" // ← これを追加！

	"github.com/ryo02puyopuyo/sudoku_online/backend/game"
	"github.com/ryo02puyopuyo/sudoku_online/backend/hub"
)

func main() {
	// 1. ゲーム状態管理オブジェクトを生成
	gameInstance := game.NewGame()

	// 2. WebSocketハブを生成し、ゲームインスタンスを渡す
	hubInstance := hub.NewHub(gameInstance)

	// 3. HTTPルートを設定
	http.HandleFunc("/ws", hubInstance.ServeWs) // WebSocket接続をハブに任せる
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/test", TestHandler)

	// 4. サーバーを起動
	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
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