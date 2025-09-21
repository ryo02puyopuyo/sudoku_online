package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // CORS対策
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string)

// WebSocket 接続の処理
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()
	clients[ws] = true

	for {
		var msg string
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}
}

// メッセージをブロードキャスト
func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			client.WriteJSON(msg)
		}
	}
}

func main() {
	// WebSocket ハンドラー
	http.HandleFunc("/ws", handleConnections)

	// HTTP ルート "/" ハンドラー
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Go server is running!")
	})

	// メッセージブロードキャストをゴルーチンで起動
	go handleMessages()

	fmt.Println("Go WebSocket server running on :8080")
	http.ListenAndServe(":8080", nil)
}
