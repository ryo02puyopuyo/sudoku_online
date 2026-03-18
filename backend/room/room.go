package room

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/ryo02puyopuyo/sudoku_online/backend/game"
	"github.com/ryo02puyopuyo/sudoku_online/backend/models"
)

type Room struct {
	ID      string
	Name    string
	Game    *game.Game
	Clients map[*websocket.Conn]*models.Player
	Mu      sync.Mutex
}

func NewRoom(id, name string) *Room {
	return &Room{
		ID:      id,
		Name:    name,
		Game:    game.NewGame(),
		Clients: make(map[*websocket.Conn]*models.Player),
	}
}

func (r *Room) AddClient(conn *websocket.Conn, player *models.Player) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Clients[conn] = player
}

func (r *Room) RemoveClient(conn *websocket.Conn) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	delete(r.Clients, conn)
}

func (r *Room) GetPlayerCount() int {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	return len(r.Clients)
}

type RoomManager struct {
	Rooms map[string]*Room
	Mu    sync.Mutex
}

func NewRoomManager(count int) *RoomManager {
	rm := &RoomManager{
		Rooms: make(map[string]*Room),
	}
	for i := 1; i <= count; i++ {
		id := fmt.Sprintf("%d", i)
		name := fmt.Sprintf("ルーム%d", i)
		rm.Rooms[id] = NewRoom(id, name)
	}
	return rm
}

func (rm *RoomManager) GetRoom(id string) (*Room, bool) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()
	room, exists := rm.Rooms[id]
	return room, exists
}

func (rm *RoomManager) GetAllRoomInfo() []models.RoomInfo {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	var infoList []models.RoomInfo
	for _, room := range rm.Rooms {
		infoList = append(infoList, models.RoomInfo{
			ID:          room.ID,
			Name:        room.Name,
			PlayerCount: room.GetPlayerCount(),
			Scores:      room.Game.GetScores(),
		})
	}

	sort.Slice(infoList, func(i, j int) bool {
		id1, _ := strconv.Atoi(infoList[i].ID)
		id2, _ := strconv.Atoi(infoList[j].ID)
		return id1 < id2
	})

	return infoList
}
