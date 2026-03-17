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
	"github.com/gorilla/websocket"
	"github.com/ryo02puyopuyo/sudoku_online/backend/room"
	"gorm.io/gorm"
)

type Hub struct {
	mu          sync.Mutex
	roomManager *room.RoomManager
	DB          *gorm.DB
	upgrader    websocket.Upgrader
	nextUserID  int
}

func NewHub(rm *room.RoomManager, dbConn *gorm.DB) *Hub {
	return &Hub{
		roomManager: rm,
		DB:          dbConn,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		nextUserID: 1,
	}
}

func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		log.Printf("Room ID is missing")
		conn.Close()
		return
	}

	targetRoom, exists := h.roomManager.GetRoom(roomID)
	if !exists {
		log.Printf("Room %s does not exist", roomID)
		conn.Close()
		return
	}

	var player *models.Player
	var registeredUserID uint = 0

	// ユーザー識別: JWT認証済みなら登録ユーザー、それ以外はゲスト
	if user, ok := r.Context().Value("user").(*db.User); ok {
		player = &models.Player{
			ID:   fmt.Sprintf("user-%d", user.ID),
			Name: user.Username,
			Team: 1,
			Role: user.Role,
		}
		registeredUserID = user.ID
		log.Printf("[WebSocket] 登録ユーザー接続: %s (ID: %d)", user.Username, user.ID)
	} else {
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

	targetRoom.AddClient(conn, player)

	defer func() {
		targetRoom.RemoveClient(conn)
		conn.Close()

		if registeredUserID != 0 {
			log.Printf("Registered user disconnecting, deleting token: userID=%d", registeredUserID)
			db.DeleteUserToken(h.DB, registeredUserID)
		}

		log.Printf("%s disconnected from room %s.", player.ID, targetRoom.ID)
		h.broadcastUserListUpdate(targetRoom)
	}()

	welcomePayload := models.WelcomePayload{
		YourPlayer: *player,
		BoardState: targetRoom.Game.GetBoard(),
		RoomID:     targetRoom.ID,
		RoomName:   targetRoom.Name,
	}
	conn.WriteJSON(models.ServerMessage{Type: "welcome", Payload: welcomePayload})

	h.broadcastUserListUpdate(targetRoom)

	for {
		var msg models.ClientMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		h.handleMessage(conn, targetRoom, player, msg)
	}
}

func (h *Hub) handleMessage(conn *websocket.Conn, targetRoom *room.Room, player *models.Player, msg models.ClientMessage) {
	isOver, _ := targetRoom.Game.GetGameOverState()
	if isOver && msg.Type != "new_puzzle" {
		return
	}

	switch msg.Type {
	case "new_puzzle":
		targetRoom.Game.Reset()
		h.broadcastNewGameStarted(targetRoom)
		h.broadcastBoardState(targetRoom)
		h.broadcastUserListUpdate(targetRoom)

	case "cell_update":
		payload, _ := msg.Payload.(map[string]interface{})
		row, col, value := int(payload["row"].(float64)), int(payload["col"].(float64)), int(payload["value"].(float64))

		updateResult, boardCompleted := targetRoom.Game.UpdateCell(row, col, value, player.Team)

		isHotSpotHit := false
		if updateResult == game.ResultHotSpot {
			isHotSpotHit = true
		}

		// コンボ処理
		targetRoom.Mu.Lock()
		if updateResult == game.ResultCorrect || updateResult == game.ResultHotSpot {
			player.ConsecutiveCorrect++
			
			// ボーナス判定: 下一桁5なら+1, 10の倍数なら+2
			bonus := 0
			if player.ConsecutiveCorrect%10 == 5 {
				bonus = 1
			} else if player.ConsecutiveCorrect > 0 && player.ConsecutiveCorrect%10 == 0 {
				bonus = 2
			}
			
			if bonus > 0 {
				// スコア付与
				targetRoom.Mu.Unlock() // AddScoreが中でLockを取るので一旦Unlock
				targetRoom.Game.AddScore(player.Team, bonus)
				
				// チャット通知
				bonusMsg := fmt.Sprintf("%s 選手が怒涛の %d 連続正解！(ボーナス +%d点)", player.Name, player.ConsecutiveCorrect, bonus)
				chatMessage := models.ChatMessage{
					SenderName: "SYSTEM",
					SenderTeam: 0,
					Message:    bonusMsg,
					Timestamp:  time.Now().Format("15:04"),
				}
				h.broadcastChatMessage(targetRoom, chatMessage)
				targetRoom.Mu.Lock() // 戻す
			}
		} else if updateResult == game.ResultIncorrect {
			player.ConsecutiveCorrect = 0
		}
		targetRoom.Mu.Unlock()

		h.broadcastBoardState(targetRoom)
		h.broadcastUserListUpdate(targetRoom)

		if isHotSpotHit {
			chatMessage := models.ChatMessage{
				SenderName: "SYSTEM",
				SenderTeam: 0,
				Message:    fmt.Sprintf("Team %d get hotspot!(+3points)", player.Team),
				Timestamp:  time.Now().Format("15:04"),
			}
			h.broadcastChatMessage(targetRoom, chatMessage)
		}

		if boardCompleted {
			_, gameOverPayload := targetRoom.Game.GetGameOverState()
			h.broadcastGameOver(targetRoom, *gameOverPayload)
		}

	case "change_team":
		payload, _ := msg.Payload.(map[string]interface{})
		team := int(payload["team"].(float64))
		if team == 1 || team == 2 {
			targetRoom.Mu.Lock()
			player.Team = team
			targetRoom.Mu.Unlock()
			log.Printf("%s changed to Team %d in room %s", player.Name, team, targetRoom.ID)
			h.broadcastUserListUpdate(targetRoom)
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
		targetRoom.Mu.Lock()
		originalName := player.Name
		player.Name = newName
		targetRoom.Mu.Unlock()
		log.Printf("Player name changed: %s -> %s in room %s", originalName, newName, targetRoom.ID)
		h.broadcastUserListUpdate(targetRoom)

	case "send_chat_message":
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return
		}
		chatText, ok := payload["message"].(string)
		if !ok || len(chatText) == 0 || len(chatText) > 100 {
			return
		}

		// チートコマンド処理 (":cheat <command>" 形式)
		if strings.HasPrefix(chatText, ":cheat") {
			if player.Role == "admin" {
				log.Printf("Admin cheat by %s in room %s", player.Name, targetRoom.ID)
				return
			}
			log.Printf("%s issued a cheat command: %s in room %s", player.Name, chatText, targetRoom.ID)
			h.handleCheatCommand(conn, targetRoom, player, chatText)
			return
		}

		chatMessage := models.ChatMessage{
			SenderName: player.Name,
			SenderTeam: player.Team,
			Message:    chatText,
			Timestamp:  time.Now().Format("15:04"),
		}
		h.broadcastChatMessage(targetRoom, chatMessage)
	}
}

func (h *Hub) handleCheatCommand(conn *websocket.Conn, targetRoom *room.Room, player *models.Player, command string) {
	log.Printf("handleCheatCommand: %s", command)
	command = strings.TrimPrefix(command, ":cheat")
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "setscore":
		if len(args) != 2 {
			return
		}
		team, err1 := strconv.Atoi(args[0])
		points, err2 := strconv.Atoi(args[1])
		if err1 != nil || err2 != nil {
			return
		}
		log.Printf("%s is setting Team %d score to %d in room %s", player.Name, team, points, targetRoom.ID)
		targetRoom.Game.SetScore(team, points)
		h.broadcastBoardState(targetRoom)
		h.broadcastUserListUpdate(targetRoom)

	case "godcyclone":
		log.Printf("%s is using godcyclone cheat! in room %s", player.Name, targetRoom.ID)
		targetRoom.Game.SetScore(player.Team, 999)
		otherTeam := 1
		if player.Team == 1 {
			otherTeam = 2
		}
		targetRoom.Game.SetScore(otherTeam, -999)
		h.broadcastBoardState(targetRoom)
		h.broadcastUserListUpdate(targetRoom)

	default:
		return
	}
}

// --- ブロードキャスト関数群 ---

// --- ブロードキャスト関数群 ---

func (h *Hub) broadcastToRoom(targetRoom *room.Room, message interface{}) {
	targetRoom.Mu.Lock()
	defer targetRoom.Mu.Unlock()
	for client := range targetRoom.Clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("Broadcast error: %v. Removing client from room %s.", err, targetRoom.ID)
			client.Close()
			delete(targetRoom.Clients, client)
		}
	}
}

func (h *Hub) broadcastBoardState(targetRoom *room.Room) {
	message := models.ServerMessage{Type: "board_state", Payload: targetRoom.Game.GetBoard()}
	h.broadcastToRoom(targetRoom, message)
}

func (h *Hub) broadcastUserListUpdate(targetRoom *room.Room) {
	targetRoom.Mu.Lock()
	var userList []models.Player
	for _, p := range targetRoom.Clients {
		userList = append(userList, *p)
	}
	targetRoom.Mu.Unlock()

	payload := models.UserListUpdatePayload{Players: userList, Scores: targetRoom.Game.GetScores()}
	message := models.ServerMessage{Type: "user_list_update", Payload: payload}
	h.broadcastToRoom(targetRoom, message)
}

func (h *Hub) broadcastGameOver(targetRoom *room.Room, payload models.GameOverPayload) {
	msg := models.ServerMessage{Type: "game_over", Payload: payload}
	h.broadcastToRoom(targetRoom, msg)
}

func (h *Hub) broadcastChatMessage(targetRoom *room.Room, payload models.ChatMessage) {
	msg := models.ServerMessage{Type: "new_chat_message", Payload: payload}
	h.broadcastToRoom(targetRoom, msg)
}

func (h *Hub) broadcastNewGameStarted(targetRoom *room.Room) {
	msg := models.ServerMessage{Type: "new_game_started"}
	h.broadcastToRoom(targetRoom, msg)
}

func (h *Hub) GetConnectionCount() int {
	count := 0
	h.roomManager.Mu.Lock()
	defer h.roomManager.Mu.Unlock()
	for _, targetRoom := range h.roomManager.Rooms {
		count += targetRoom.GetPlayerCount()
	}
	return count
}

func (h *Hub) GetPlayerList() []models.Player {
	var list []models.Player
	h.roomManager.Mu.Lock()
	defer h.roomManager.Mu.Unlock()
	for _, targetRoom := range h.roomManager.Rooms {
		targetRoom.Mu.Lock()
		for _, p := range targetRoom.Clients {
			list = append(list, *p)
		}
		targetRoom.Mu.Unlock()
	}
	return list
}
