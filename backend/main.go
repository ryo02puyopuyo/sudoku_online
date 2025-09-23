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



var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// --- サーバーからクライアントへ送るメッセージの統一形式 ---
type ServerMessage struct {
	Type    string      `json:"type"`    // メッセージの種類 ("puzzle_state" or "user_list")
	Payload interface{} `json:"payload"` // 実際のデータ
}

// PuzzleState はパズルの状態を保持
type PuzzleState struct {
	Puzzle   [9][9]int `json:"puzzle"`
	Solution [9][9]int `json:"solution"`
}

// --- グローバル変数として、サーバーに一つだけの状態を管理 ---
var (
	mu            sync.Mutex
	// 接続中のクライアントを<接続情報: ユーザーID>の形式で保持
	clients       = make(map[*websocket.Conn]string)
	currentPuzzle PuzzleState
	nextUserID    = 1 // ユーザーIDを発行するためのカウンター
)

// 最新のメンバー一覧を作成し、全員にブロードキャストする
func broadcastUserList() {
	mu.Lock()
	defer mu.Unlock()

	var userList []string
	// clientsマップからIDだけを収集
	for _, userID := range clients {
		userList = append(userList, userID)
	}

	message := ServerMessage{Type: "user_list", Payload: userList}

	log.Printf("Broadcasting user list to %d clients...", len(clients))
	for client := range clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("User list broadcast error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// 最新のパズルを全員にブロードキャストする
func broadcastPuzzleState() {
	mu.Lock()
	defer mu.Unlock()

	message := ServerMessage{Type: "puzzle_state", Payload: currentPuzzle}

	log.Printf("Broadcasting puzzle to %d clients...", len(clients))
	for client := range clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("Puzzle broadcast error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
	log.Println("Puzzle broadcast complete.")
}

// 新しいパズルを生成して保存する
func generateNewPuzzle() {
	mu.Lock()
	defer mu.Unlock()
	// (この関数の内部は変更なし、ロックを追加しただけ)
	solution, err := util.GenerateSolvedGrid(1000)
	if err != nil {
		log.Printf("Error generating new grid: %v", err)
		return
	}
	puzzle := createPuzzleFromSolution(solution, 0.5)
	currentPuzzle = PuzzleState{
		Puzzle:   puzzle,
		Solution: solution,
	}
	log.Println("A new puzzle has been generated and stored.")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// ---- 1. クライアント接続時の処理 ----
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

	// 2. 接続してきたクライアントに、まず現在のパズルを送信
	mu.Lock()
	initialMessage := ServerMessage{Type: "puzzle_state", Payload: currentPuzzle}
	if err := conn.WriteJSON(initialMessage); err != nil {
		mu.Unlock()
		log.Printf("Initial puzzle send error: %v", err)
		return
	}
	mu.Unlock()

	// 3. 全員に最新のメンバー一覧をブロードキャスト
	broadcastUserList()

	// ---- 4. クライアントからのメッセージを待つループ ----
	for {
		var msg string
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		if msg == "new_puzzle" {
			log.Printf("Received 'new_puzzle' request from %s.", userID)
			generateNewPuzzle()      // 新しい問題を生成し
			broadcastPuzzleState() // 全員にブロードキャストする
		}
	}
}

func main() {
	generateNewPuzzle()
	http.HandleFunc("/ws", handleConnections)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// (createPuzzleFromSolution 関数は変更なし)
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