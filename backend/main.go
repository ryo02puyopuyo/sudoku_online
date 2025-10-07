package main

import (
	"log"
	"net/http"

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

	// 4. サーバーを起動
	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
