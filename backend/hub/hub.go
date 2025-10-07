package hub

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ryo02puyopuyo/sudoku_online/backend/game"
	"github.com/ryo02puyopuyo/sudoku_online/backend/models"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu         sync.Mutex
	clients    map[*websocket.Conn]*models.Player
	game       *game.Game // Gameの状態への参照を持つ
	upgrader   websocket.Upgrader
	nextUserID int
}

func NewHub(game *game.Game) *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]*models.Player),
		game:    game,
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

	// 新しいプレイヤーを作成
	h.mu.Lock()
	playerID := fmt.Sprintf("Player %d", h.nextUserID)
	player := &models.Player{ID: playerID, Name: playerID, Team: 1}
	h.nextUserID++
	h.clients[conn] = player
	h.mu.Unlock()

	log.Printf("%s (%s) connected.", player.ID, player.Name)

	// 接続が終了した際のクリーンアップ
	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
		log.Printf("%s disconnected.", player.ID)
		h.broadcastUserListUpdate()
	}()

	// Welcomeメッセージを送信
	welcomePayload := models.WelcomePayload{
		YourPlayer: *player,
		BoardState: h.game.GetBoard(),
	}
	if err := conn.WriteJSON(models.ServerMessage{Type: "welcome", Payload: welcomePayload}); err != nil {
		log.Printf("Welcome message error: %v", err)
		return
	}

	// 接続時にゲームが終了していたら、その結果を送信
	isOver, gameOverPayload := h.game.GetGameOverState()
	if isOver {
		if err := conn.WriteJSON(models.ServerMessage{Type: "game_over", Payload: *gameOverPayload}); err != nil {
			log.Printf("Could not send late game_over message to %s", player.ID)
		}
	}

	// 全員に最新のユーザーリストを送信
	h.broadcastUserListUpdate()

	// クライアントからのメッセージを待ち受けるループ
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
		chatMessage := models.ChatMessage{
			SenderName: player.Name,
			SenderTeam: player.Team,
			Message:    chatText,
			Timestamp:  time.Now().Format("15:04"),
		}
		h.broadcastChatMessage(chatMessage)
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
