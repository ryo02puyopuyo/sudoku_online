package main

import (
	// ← これを追加！
	"log"
	"net/http"

	"github.com/ryo02puyopuyo/sudoku_online/backend/db"
	"github.com/ryo02puyopuyo/sudoku_online/backend/game"
	"github.com/ryo02puyopuyo/sudoku_online/backend/handler"
	"github.com/ryo02puyopuyo/sudoku_online/backend/hub"
)

func main() {
	// 1. ゲーム状態管理オブジェクトを生成
	gameInstance := game.NewGame()

	// 2. WebSocketハブを生成し、ゲームインスタンスを渡す
	hubInstance := hub.NewHub(gameInstance)

	//db接続

	database, err := db.SetupDatabase()
	if err != nil {
		log.Fatal(err)
	}

	// 3. HTTPルートを設定
	http.HandleFunc("/ws", hubInstance.ServeWs) // WebSocket接続をハブに任せる
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/test", handler.TestHandler)
	//test
	http.HandleFunc("/api/check", func(w http.ResponseWriter, r *http.Request) {
		handler.CheckHandler(w, r, database)
	})

	http.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		handler.RegisterHandler(w, r, database)
	})
	//まだ
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		handler.LoginHandler(w, r, database)
	})

	// 4. サーバーを起動
	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
