package hub

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ryo02puyopuyo/sudoku_online/backend/db"
	"github.com/ryo02puyopuyo/sudoku_online/backend/game"
	"github.com/ryo02puyopuyo/sudoku_online/backend/models"
	"gorm.io/gorm"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu         sync.Mutex
	clients    map[*websocket.Conn]*models.Player
	game       *game.Game // Gameの状態への参照を持つ
	DB         *gorm.DB
	upgrader   websocket.Upgrader
	nextUserID int
}

// 修正点: 引数に dbConn を追加
func NewHub(game *game.Game, dbConn *gorm.DB) *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]*models.Player),
		game:    game,
		DB:      dbConn,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		nextUserID: 1,
	}
}

// ServeWs はWebSocketリクエストを処理します
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	var player *models.Player
	var registeredUserID uint = 0 // 修正点: deferで使うためにIDを保持

	// --- 【修正点】ユーザー識別ロジックの追加 ---
	if user, ok := r.Context().Value("user").(*db.User); ok {
		// 【登録ユーザー】
		// DBから取得した本物のユーザー名を使用
		player = &models.Player{
			ID:   fmt.Sprintf("user-%d", user.ID), // IDをプレフィックス付きにしてゲストと区別
			Name: user.Username,
			Team: 1,
			Role: user.Role,
		}
		registeredUserID = user.ID
		log.Printf("[WebSocket] 登録ユーザー接続: %s (ID: %d)", user.Username, user.ID)
	} else {
		// 【ゲストユーザー】
		h.mu.Lock()
		guestName := fmt.Sprintf("Guest%d", h.nextUserID)
		player = &models.Player{
			ID:   guestName,
			Name: guestName,
			Team: 1,
			Role: "guest",
		}
		h.nextUserID++
		h.mu.Unlock()
		log.Printf("[WebSocket] ゲスト接続: %s", guestName)
	}
	// ------------------------------------------

	h.mu.Lock()
	h.clients[conn] = player
	h.mu.Unlock()

	// 接続終了時のクリーンアップ
	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()

		// 【重要】プレイヤーが抜けるとDBのトークンを削除し、再ログインを強制する
		if registeredUserID != 0 {
			log.Printf("registerd user disconnecting, deleting token: userID=%d", registeredUserID)
			db.DeleteUserToken(h.DB, registeredUserID)
		}

		log.Printf("%s disconnected.", player.ID)
		h.broadcastUserListUpdate()
	}()

	// Welcomeメッセージの送信（新しい player 情報を含める）
	welcomePayload := models.WelcomePayload{
		YourPlayer: *player,
		BoardState: h.game.GetBoard(),
	}
	conn.WriteJSON(models.ServerMessage{Type: "welcome", Payload: welcomePayload})

	h.broadcastUserListUpdate()

	for {
		var msg models.ClientMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		h.handleMessage(conn, player, msg)
	}
}

// handleMessage はクライアントからのメッセージを種類に応じて処理します
func (h *Hub) handleMessage(conn *websocket.Conn, player *models.Player, msg models.ClientMessage) {
	isOver, _ := h.game.GetGameOverState()
	if isOver && msg.Type != "new_puzzle" {
		return // ゲーム終了後は新しいパズル要求以外は無視
	}

	switch msg.Type {
	case "new_puzzle":
		h.game.Reset()
		h.broadcastNewGameStarted()
		h.broadcastBoardState()
		h.broadcastUserListUpdate()

	case "cell_update":
		payload, _ := msg.Payload.(map[string]interface{})
		row, col, value := int(payload["row"].(float64)), int(payload["col"].(float64)), int(payload["value"].(float64))

		boardCompleted := h.game.UpdateCell(row, col, value, player.Team)

		h.broadcastBoardState()
		h.broadcastUserListUpdate()

		if boardCompleted {
			_, gameOverPayload := h.game.GetGameOverState()
			h.broadcastGameOver(*gameOverPayload)
		}

	case "change_team":
		payload, _ := msg.Payload.(map[string]interface{})
		team := int(payload["team"].(float64))
		if team == 1 || team == 2 {
			h.mu.Lock()
			player.Team = team
			h.mu.Unlock()
			log.Printf("%s changed to Team %d", player.Name, team)
			h.broadcastUserListUpdate()
		}

	case "change_name":
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return
		}
		newName, ok := payload["name"].(string)
		if !ok || len(newName) == 0 || len(newName) > 15 {
			return
		}
		h.mu.Lock()
		originalName := player.Name
		player.Name = newName
		h.mu.Unlock()
		log.Printf("Player name changed: %s -> %s", originalName, newName)
		h.broadcastUserListUpdate()

	case "send_chat_message":
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return
		}
		chatText, ok := payload["message"].(string)
		if !ok || len(chatText) == 0 || len(chatText) > 100 {
			return
		}
		//チートコマンド (":cheat abc")の形式
		if strings.HasPrefix(chatText, ":cheat") /*&& player.role == "admin"***/ {
			if player.Role == "admin" {
				log.Printf("admin cheat  %s", player.Name)
				return
			}
			log.Printf("%s issued a cheat command: %s", player.Name, chatText)
			h.handleCheatCommand( /***player,***/ conn, player, chatText)
			return
		}

		chatMessage := models.ChatMessage{
			SenderName: player.Name,
			SenderTeam: player.Team,
			Message:    chatText,
			Timestamp:  time.Now().Format("15:04"),
		}
		h.broadcastChatMessage(chatMessage)
	}
}

func (h *Hub) handleCheatCommand(conn *websocket.Conn, player *models.Player, command string) {
	// 先頭の'/'を除去し、スペースで分割
	log.Printf("in handleCheatCommand: %s", command)
	command = strings.TrimPrefix(command, ":cheat")
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return // コマンドが空
	}

	cmd := parts[0]   // コマンド名 (例: "addscore")
	args := parts[1:] // 引数のスライス (例: ["1", "100"])

	switch cmd {
	case "setscore":
		if len(args) != 2 {
			//h.sendPrivateMessage(conn, "使用法: /setscore <1,2>(int) <points>(int)")
			return
		}
		team, err1 := strconv.Atoi(args[0])
		points, err2 := strconv.Atoi(args[1])
		if err1 != nil || err2 != nil {
			//h.sendPrivateMessage(conn, "引数が不正です。チームとポイントは数字である必要があります。")
			return
		}
		log.Printf("%s is setting Team %d score to %d", player.Name, team, points)
		h.game.SetScore(team, points)
		h.broadcastBoardState()
		h.broadcastUserListUpdate()

	case "godcyclone":
		//使用者のチームの点数を999に、相手チームを-999にする
		log.Printf("%s is using godcyclone cheat!", player.Name)
		h.game.SetScore(player.Team, 999)
		otherTeam := 1
		if player.Team == 1 {
			otherTeam = 2
		}
		h.game.SetScore(otherTeam, -999)
		h.broadcastBoardState()
		h.broadcastUserListUpdate()

	default:
		return // 未知のコマンドはとりあえず無視
	}
}

// --- ブロードキャスト関数群 ---

func (h *Hub) broadcastToAll(message interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("Broadcast error: %v. Removing client.", err)
			client.Close()
			delete(h.clients, client)
		}
	}
}

func (h *Hub) broadcastBoardState() {
	message := models.ServerMessage{Type: "board_state", Payload: h.game.GetBoard()}
	h.broadcastToAll(message)
}

func (h *Hub) broadcastUserListUpdate() {
	h.mu.Lock()
	var userList []models.Player
	for _, p := range h.clients {
		userList = append(userList, *p)
	}
	h.mu.Unlock()

	payload := models.UserListUpdatePayload{Players: userList, Scores: h.game.GetScores()}
	message := models.ServerMessage{Type: "user_list_update", Payload: payload}
	h.broadcastToAll(message)
}

func (h *Hub) broadcastGameOver(payload models.GameOverPayload) {
	msg := models.ServerMessage{Type: "game_over", Payload: payload}
	h.broadcastToAll(msg)
}

func (h *Hub) broadcastChatMessage(payload models.ChatMessage) {
	msg := models.ServerMessage{Type: "new_chat_message", Payload: payload}
	h.broadcastToAll(msg)
}

func (h *Hub) broadcastNewGameStarted() {
	msg := models.ServerMessage{Type: "new_game_started"}
	h.broadcastToAll(msg)
}

// サーバー監視用の関数軍
func (h *Hub) GetConnectionCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// 修正点: プレイヤー情報のスライスを返す (読み取り専用)
func (h *Hub) GetPlayerList() []models.Player {
	h.mu.Lock()
	defer h.mu.Unlock()

	var list []models.Player
	for _, p := range h.clients {
		list = append(list, *p)
	}
	return list
}
