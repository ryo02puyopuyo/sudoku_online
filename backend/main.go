package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
  "sync"

	"github.com/gorilla/websocket"
	"github.com/ryo02puyopuyo/sudoku_online/backend/util" // パスはプロジェクトに合わせてください
)

// WebSocketのアップグレーダー設定
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // すべてのオリジンを許可
}

// --- サーバー・クライアント間のメッセージ構造体 ---

// サーバーからクライアントへ送るメッセージ
type ServerMessage struct {
	Type    string      `json:"type"`    // メッセージの種類 ("board_state" or "user_list")
	Payload interface{} `json:"payload"` // 実際のデータ
}

// クライアントからサーバーへ送るメッセージ
type ClientMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// --- 盤面の状態を定義する構造体 ---

// 1マスごとの状態
type Cell struct {
	Value  int    `json:"value"`  // 0は空
	Status string `json:"status"` // "fixed", "correct", "wrong", "empty"
}

// --- サーバー全体で共有するグローバルな状態 ---
var (
	mu              sync.Mutex                     // 状態への同時アクセスを防ぐロック
	clients         = make(map[*websocket.Conn]string) // <接続情報: ユーザーID>
	currentBoard    [9][9]Cell                     // 現在の盤面の全セルの状態
	currentSolution [9][9]int                      // 現在の盤面の解答
	nextUserID      = 1                            // ユーザーID発行カウンター
)

// 新しいパズルと盤面状態を生成し、グローバル変数を更新する
func generateNewBoardState() {
	mu.Lock()
	defer mu.Unlock()

	solution, err := util.GenerateSolvedGrid(1000)
	if err != nil {
		log.Printf("Error generating grid: %v", err)
		return
	}
	currentSolution = solution
	puzzle := createPuzzleFromSolution(solution, 0.5) // 50%のマスを空にする

	var board [9][9]Cell
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if puzzle[r][c] != 0 {
				board[r][c] = Cell{Value: puzzle[r][c], Status: "fixed"}
			} else {
				board[r][c] = Cell{Value: 0, Status: "empty"}
			}
		}
	}
	currentBoard = board
	log.Println("A new board state has been generated and stored.")
}

// 最新の盤面状態を全員にブロードキャストする
func broadcastBoardState() {
	mu.Lock()
	defer mu.Unlock()

	message := ServerMessage{Type: "board_state", Payload: currentBoard}

	log.Printf("Broadcasting board state to %d clients...", len(clients))
	for client := range clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("Board state broadcast error: %v. Removing client.", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// 最新のメンバー一覧を全員にブロードキャストする
func broadcastUserList() {
	mu.Lock()
	defer mu.Unlock()

	var userList []string
	for _, userID := range clients {
		userList = append(userList, userID)
	}
	message := ServerMessage{Type: "user_list", Payload: userList}

	log.Printf("Broadcasting user list to %d clients...", len(clients))
	for client := range clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("User list broadcast error: %v. Removing client.", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// WebSocket接続ごとの処理
func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// 1. クライアント接続時の処理
	var userID string
	mu.Lock()
	userID = fmt.Sprintf("Player %d", nextUserID)
	nextUserID++
	clients[conn] = userID // 新しいクライアントをIDと共にマップに追加
	mu.Unlock()
	log.Printf("%s connected. Total clients: %d", userID, len(clients))

	// 接続が終了したとき（関数を抜けるとき）に必ず実行
	defer func() {
		mu.Lock()
		delete(clients, conn) // クライアントをマップから削除
		mu.Unlock()
		conn.Close()
		log.Printf("%s disconnected. Total clients: %d", userID, len(clients))
		broadcastUserList() // メンバーが抜けたことを全員に通知
	}()

	// 2. 接続してきたクライアントに、まず現在の盤面状態を送信
	mu.Lock()
	initialMessage := ServerMessage{Type: "board_state", Payload: currentBoard}
	if err := conn.WriteJSON(initialMessage); err != nil {
		mu.Unlock()
		log.Printf("Initial puzzle send error for %s: %v", userID, err)
		return
	}
	mu.Unlock()

	// 3. 全員に最新のメンバー一覧をブロードキャスト
	broadcastUserList()

	// 4. クライアントからのメッセージを待つループ
	for {
		var msg ClientMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break // 読み取りエラー＝接続が切れたと判断
		}

		switch msg.Type {
		case "new_puzzle":
			log.Printf("Received 'new_puzzle' request from %s.", userID)
			generateNewBoardState()
			broadcastBoardState()

		case "cell_update":
			payload, ok := msg.Payload.(map[string]interface{})
			if !ok { continue }

			row := int(payload["row"].(float64))
			col := int(payload["col"].(float64))
			value := int(payload["value"].(float64))

			mu.Lock()
			if currentBoard[row][col].Status != "fixed" {
				if value == 0 {
					currentBoard[row][col] = Cell{Value: 0, Status: "empty"}
				} else if value == currentSolution[row][col] {
					currentBoard[row][col] = Cell{Value: value, Status: "correct"}
				} else {
					currentBoard[row][col] = Cell{Value: value, Status: "wrong"}
				}
			}
			mu.Unlock()

			broadcastBoardState()
		}
	}
}

func main() {
	generateNewBoardState() // サーバー起動時に最初の盤面状態を生成
	http.HandleFunc("/ws", handleConnections)
	http.Handle("/", http.FileServer(http.Dir("./static"))) // フロントエンドのビルドファイルを配信

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// 解答から問題を作成するヘルパー関数
func createPuzzleFromSolution(solution [9][9]int, difficulty float64) [9][9]int {
	var puzzle [9][9]int
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if util.RandFloat() < difficulty {
				puzzle[r][c] = 0
			} else {
				puzzle[r][c] = solution[r][c]
			}
		}
	}
	return puzzle
}