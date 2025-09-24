package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ryo02puyopuyo/sudoku_online/backend/util" // パスはプロジェクトに合わせてください
)

// --- 構造体定義 ---
type ServerMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
type ClientMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Team int    `json:"team"`
}
type Cell struct {
	Value        int    `json:"value"`
	Status       string `json:"status"`
	FilledByTeam int    `json:"filledByTeam"`
}
type Score struct {
	Team1 int `json:"team1"`
	Team2 int `json:"team2"`
}
type UserListUpdatePayload struct {
	Players []Player `json:"players"`
	Scores  Score    `json:"scores"`
}
type WelcomePayload struct {
	YourPlayer Player     `json:"yourPlayer"`
	BoardState [9][9]Cell `json:"boardState"`
}
type ChatMessage struct {
	SenderName string `json:"senderName"`
	SenderTeam int    `json:"senderTeam"`
	Message    string `json:"message"`
	Timestamp  string `json:"timestamp"`
}
type GameOverPayload struct {
	WinnerTeam  int   `json:"winnerTeam"`
	FinalScores Score `json:"finalScores"`
}

// --- グローバル変数 ---
var (
	mu                  sync.Mutex
	clients             = make(map[*websocket.Conn]*Player)
	currentBoard        [9][9]Cell
	currentSolution     [9][9]int
	nextUserID          = 1
	currentScores       Score
	isGameOver          bool
	lastGameOverPayload *GameOverPayload
	upgrader            = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// 新しいゲームを開始する際に状態をリセット
func generateNewBoardState() {
	mu.Lock()
	defer mu.Unlock()

	currentScores = Score{Team1: 0, Team2: 0}
	isGameOver = false
	lastGameOverPayload = nil

	solution, err := util.GenerateSolvedGrid(1000)
	if err != nil {
		log.Printf("Error generating grid: %v", err)
		return
	}
	currentSolution = solution
	puzzle := createPuzzleFromSolution(solution, 0.5)
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
	log.Println("A new board state has been generated and game state has been reset.")
}

// 最新の盤面状態を全員にブロードキャスト
func broadcastBoardState() {
	mu.Lock()
	defer mu.Unlock()
	message := ServerMessage{Type: "board_state", Payload: currentBoard}
	for client := range clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("Board state broadcast error: %v. Removing client.", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// 最新のメンバー一覧とスコアを全員にブロードキャスト
func broadcastUserListUpdate() {
	mu.Lock()
	defer mu.Unlock()

	var userList []Player
	for _, player := range clients {
		userList = append(userList, *player)
	}

	payload := UserListUpdatePayload{Players: userList, Scores: currentScores}
	message := ServerMessage{Type: "user_list_update", Payload: payload}
	for client := range clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("User list broadcast error: %v. Removing client.", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// ゲーム終了を全員にブロードキャスト
func broadcastGameOver(payload GameOverPayload) {
	mu.Lock()
	defer mu.Unlock()

	msg := ServerMessage{Type: "game_over", Payload: payload}
	for client := range clients {
		if err := client.WriteJSON(msg); err != nil {
			log.Printf("Game Over broadcast error: %v. Removing client.", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// チャットメッセージを全員にブロードキャスト
func broadcastChatMessage(message ChatMessage) {
	mu.Lock()
	defer mu.Unlock()

	msg := ServerMessage{Type: "new_chat_message", Payload: message}

	for client := range clients {
		if err := client.WriteJSON(msg); err != nil {
			log.Printf("Chat broadcast error: %v. Removing client.", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// 新しいゲームが始まったことを全員に通知する
func broadcastNewGameStarted() {
	mu.Lock()
	defer mu.Unlock()

	msg := ServerMessage{Type: "new_game_started"}

	for client := range clients {
		if err := client.WriteJSON(msg); err != nil {
			log.Printf("New Game Started broadcast error: %v. Removing client.", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// WebSocket接続ごとの処理
func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	mu.Lock()
	playerID := fmt.Sprintf("Player %d", nextUserID)
	player := &Player{ID: playerID, Name: playerID, Team: 1}
	nextUserID++
	clients[conn] = player
	welcomePayload := WelcomePayload{
		YourPlayer: *player,
		BoardState: currentBoard,
	}
	mu.Unlock()

	log.Printf("%s (%s) connected.", player.ID, player.Name)

	defer func() {
		mu.Lock()
		delete(clients, conn)
		mu.Unlock()
		conn.Close()
		log.Printf("%s disconnected.", player.ID)
		broadcastUserListUpdate()
	}()

	if err := conn.WriteJSON(ServerMessage{Type: "welcome", Payload: welcomePayload}); err != nil {
		return
	}

	mu.Lock()
	if isGameOver {
		if err := conn.WriteJSON(ServerMessage{Type: "game_over", Payload: *lastGameOverPayload}); err != nil {
			log.Printf("Could not send late game_over message to %s", player.ID)
		}
	}
	mu.Unlock()

	broadcastUserListUpdate()

	for {
		var msg ClientMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		mu.Lock()
		if isGameOver && msg.Type != "new_puzzle" {
			mu.Unlock()
			continue
		}
		mu.Unlock()

		switch msg.Type {
		case "new_puzzle":
			generateNewBoardState()
			broadcastNewGameStarted() // 新しいゲームの開始を明示的に通知
			broadcastBoardState()
			broadcastUserListUpdate()
		case "cell_update":
			payload, _ := msg.Payload.(map[string]interface{})
			row, col, value := int(payload["row"].(float64)), int(payload["col"].(float64)), int(payload["value"].(float64))

			var boardCompleted = false
			mu.Lock()
			if currentBoard[row][col].Status != "fixed" && currentBoard[row][col].Status != "correct" {
				if value == 0 {
					currentBoard[row][col] = Cell{Value: 0, Status: "empty"}
				} else if value == currentSolution[row][col] {
					currentBoard[row][col] = Cell{Value: value, Status: "correct", FilledByTeam: player.Team}
					if player.Team == 1 {
						currentScores.Team1++
					} else {
						currentScores.Team2++
					}
				} else {
					currentBoard[row][col] = Cell{Value: value, Status: "wrong", FilledByTeam: player.Team}
					if player.Team == 1 {
						currentScores.Team1--
					} else {
						currentScores.Team2--
					}
				}

				isFull := true
				for r_check := 0; r_check < 9; r_check++ {
					for c_check := 0; c_check < 9; c_check++ {
						if currentBoard[r_check][c_check].Status != "correct" && currentBoard[r_check][c_check].Status != "fixed" {
							isFull = false
							break
						}
					}
				}
				if isFull {
					boardCompleted = true
					isGameOver = true
				}
			}
			mu.Unlock()

			broadcastBoardState()
			broadcastUserListUpdate()

			if boardCompleted {
				log.Println("Game Over!")
				var winner int
				if currentScores.Team1 > currentScores.Team2 {
					winner = 1
				} else if currentScores.Team2 > currentScores.Team1 {
					winner = 2
				} else {
					winner = 0
				}

				gameOverPayload := GameOverPayload{
					WinnerTeam:  winner,
					FinalScores: currentScores,
				}
				mu.Lock()
				lastGameOverPayload = &gameOverPayload
				mu.Unlock()

				broadcastGameOver(gameOverPayload)
			}
		case "change_team":
			payload, _ := msg.Payload.(map[string]interface{})
			team := int(payload["team"].(float64))
			if team == 1 || team == 2 {
				mu.Lock()
				clients[conn].Team = team
				mu.Unlock()
				log.Printf("%s changed to Team %d", player.Name, team)
				broadcastUserListUpdate()
			}
		case "change_name":
			payload, ok := msg.Payload.(map[string]interface{})
			if !ok {
				continue
			}
			newName, ok := payload["name"].(string)
			if !ok || len(newName) == 0 || len(newName) > 15 {
				continue
			}

			mu.Lock()
			originalName := clients[conn].Name
			clients[conn].Name = newName
			mu.Unlock()

			log.Printf("Player name changed: %s -> %s", originalName, newName)
			broadcastUserListUpdate()
		case "send_chat_message":
			payload, ok := msg.Payload.(map[string]interface{})
			if !ok {
				continue
			}
			chatText, ok := payload["message"].(string)
			if !ok || len(chatText) == 0 || len(chatText) > 100 {
				continue
			}

			chatMessage := ChatMessage{
				SenderName: player.Name,
				SenderTeam: player.Team,
				Message:    chatText,
				Timestamp:  time.Now().Format("15:04"),
			}
			broadcastChatMessage(chatMessage)
		}
	}
}

func main() {
	log.Println("Generating initial board...")
	generateNewBoardState()
	if currentBoard == [9][9]Cell{} {
		log.Fatal("Failed to generate initial board. Server cannot start.")
	}
	http.HandleFunc("/ws", handleConnections)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

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
